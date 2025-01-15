/*
 * skogul, generic receiver
 *
 * Copyright (c) 2019 Telenor Norge AS
 * Author(s):
 *  - Kristian Lyngstøl <kly@kly.no>
 *  - Håkon Solbjørg <hakon.solbjorg@telenor.com>
 *
 * This library is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 2.1 of the License, or (at your option) any later version.
 *
 * This library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public
 * License along with this library; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA
 * 02110-1301  USA
 */

package receiver

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"github.com/telenornms/skogul"
)

var httpLog = skogul.Logger("receiver", "http")

// HTTPAuth contains ways to authenticate a HTTP request, e.g. Username/Password for Basic Auth.
type HTTPAuth struct {
	Username              string        `doc:"Username for basic authentication. No authentication is required if left blank."`
	Password              skogul.Secret `doc:"Password for basic authentication."`
	SANDNSName            string        `doc:"DNS name which has to be present in SAN extension of x509 certificate when using Client Certificate authentication"`
	SkipCertificateVerify bool          `doc:"Skip verifying certificate. (default: false)"`
	path                  string
}

/*
HTTP accepts HTTP connections on the Address specified, and requires at
least one handler to be set up, using Handle. This is done implicitly
if the HTTP receiver is created using New()
*/
type HTTP struct {
	Address              string                        `doc:"Address to listen to." example:"[::1]:80 [2001:db8::1]:443"`
	Handlers             map[string]*skogul.HandlerRef `doc:"Paths to handlers. Need at least one." example:"{\"/\": \"someHandler\" }"`
	Auth                 map[string]*HTTPAuth          `doc:"A map corresponding to Handlers; specifying authentication for the given path, if required."`
	Certfile             string                        `doc:"Path to certificate file for TLS. If left blank, un-encrypted HTTP is used."`
	Keyfile              string                        `doc:"Path to key file for TLS."`
	ClientCertificateCAs []string                      `doc:"Paths to files containing CAs which are accepted for Client Certificate authentication."`
	Log204OK             bool                          `doc:"Log successful requests as well as failed. Failed requests are always logged as a warning.Successful requests are logged as info-level."`
	stats                *httpStats
}

// httpStats contains the internal stats of the HTTP receiver.
type httpStats struct {
	Received      uint64 // Number of valid received HTTP request
	NoData        uint64 // If Content-Length is zero
	ReadFailed    uint64 // Number of read failures of http request body
	HandlerErrors uint64 // Number of errors from upstream handlers
	Sent          uint64 // Number of sent skogul.Containers
}

// For each path we handle, we set up a receiver such as this
// to simplify things.
// FIXME: This should almost certianly have a more descriptive name to
// avoid collisions and confusion.
type receiver struct {
	Handler  *skogul.Handler
	settings *HTTP
	auth     *HTTPAuth
}

// fallback is used to handle the / path if it isn't defined, mainly to
// unify logging.
type fallback struct {
	hasAuth bool // if any auth handler is present, we return 401 instead of 404.
}

type httpReturn struct {
	Message string
}

func (auth *HTTPAuth) auth(r *http.Request) error {
	if auth.Username != "" && auth.Password != "" {
		username, pw, ok := r.BasicAuth()
		success := ok && auth.Username == username && auth.Password.Expose() == pw
		if !success {
			return fmt.Errorf("Invalid credentials")
		}

		return nil
	}

	if auth.SANDNSName != "" {
		httpLog.Trace("Verifying request using client certificates")
		if err := auth.verifyPeerCertificate(nil, r.TLS.VerifiedChains); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("no matching authentication method")
}

func answer(w http.ResponseWriter, r *http.Request, code int, inerr error) {
	answer := "OK"

	w.WriteHeader(code)
	if code == 204 {
		return
	}
	if inerr != nil {
		answer = inerr.Error()
	}

	b, err := json.Marshal(httpReturn{Message: answer})
	skogul.Assert(err == nil, err)
	fmt.Fprintf(w, "%s\n", b)
}

func (rcvr receiver) handle(w http.ResponseWriter, r *http.Request) (int, error) {
	if rcvr.auth != nil {
		if err := rcvr.auth.auth(r); err != nil {
			return 401, err
		}
	}

	atomic.AddUint64(&rcvr.settings.stats.Received, 1)
	if r.ContentLength == 0 {
		atomic.AddUint64(&rcvr.settings.stats.NoData, 1)
		return 400, fmt.Errorf("no body in HTTP request")
	}

	b := make([]byte, r.ContentLength)

	if _, err := io.ReadFull(r.Body, b); err != nil {
		atomic.AddUint64(&rcvr.settings.stats.ReadFailed, 1)
		return 400, fmt.Errorf("read error on http body: %w", err)
	}

	if err := rcvr.Handler.Handle(b); err != nil {
		atomic.AddUint64(&rcvr.settings.stats.HandlerErrors, 1)
		return 400, err
	}

	atomic.AddUint64(&rcvr.settings.stats.Sent, 1)
	return 204, nil
}

// Core HTTP handler
func (rcvr receiver) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := rcvr.handle(w, r)
	if err != nil {
		httpLog.WithFields(log.Fields{
			"code":          code,
			"remoteAddress": r.RemoteAddr,
			"requestUri":    r.RequestURI,
			"ContentLength": r.ContentLength,
		}).WithError(err).Warnf("HTTP request failed")
	} else if rcvr.settings.Log204OK {
		httpLog.WithFields(log.Fields{
			"code":          code,
			"remoteAddress": r.RemoteAddr,
			"requestUri":    r.RequestURI,
			"ContentLength": r.ContentLength,
		}).Infof("HTTP request ok")
	}
	answer(w, r, code, err)
}

// Fallback HTTP handler
func (f fallback) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code := 404
	err := fmt.Errorf("File not found")
	extra := ""
	if f.hasAuth {
		extra = " Authenticated handlers present, masking 404 as 401."
	}
	httpLog.WithFields(log.Fields{
		"code":          code,
		"remoteAddress": r.RemoteAddr,
		"requestUri":    r.RequestURI,
		"ContentLength": r.ContentLength,
	}).WithError(err).Warnf("HTTP request failed%s", extra)
	if f.hasAuth {
		code = 401
		err = fmt.Errorf("Invalid credentials")
	}
	answer(w, r, code, err)
}

// loadClientCertificateCAs loads a given list of strings as paths of
// acceptable CAs for use in Client Certificate authentication.
func loadClientCertificateCAs(paths []string) (*x509.CertPool, error) {
	httpLog.Debugf("Loading Client Certificates from %d file(s)", len(paths))
	pool := x509.NewCertPool()
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file: %w", err)
		}
		pool.AppendCertsFromPEM(data)
	}
	return pool, nil
}

// Start never returns.
func (htt *HTTP) Start() error {
	server := http.Server{}
	serveMux := http.NewServeMux()
	server.Handler = serveMux
	for idx, h := range htt.Handlers {
		httpLog.WithFields(log.Fields{
			"configuredHandler": idx,
			"selectedHandler":   h.Name,
			"hasAuth":           htt.Auth[idx] != nil,
		}).Debug("Adding handler")

		serveMux.Handle(idx, receiver{Handler: h.H, settings: htt, auth: htt.Auth[idx]})
	}
	if htt.Handlers["/"] == nil {
		f := fallback{}
		if htt.Auth != nil && len(htt.Auth) > 0 {
			f.hasAuth = true
		} else {
			f.hasAuth = false
		}
		serveMux.Handle("/", f)
	}

	if len(htt.ClientCertificateCAs) > 0 {
		pool, err := loadClientCertificateCAs(htt.ClientCertificateCAs)
		if err != nil {
			return err
		}
		server.TLSConfig = &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.VerifyClientCertIfGiven,
		}
		httpLog.Info("Configured HTTP receiver with Client Certificate authentication")
		for _, auth := range htt.Auth {
			if auth.SANDNSName != "" {
				httpLog.Info("Configured HTTP receiver with Client Certificate verification")
				break
			}
		}
	}

	htt.stats = &httpStats{
		Received:      0,
		NoData:        0,
		ReadFailed:    0,
		HandlerErrors: 0,
		Sent:          0,
	}

	server.Addr = htt.Address
	if htt.Certfile != "" {
		httpLog.WithField("address", htt.Address).Info("Starting http receiver with TLS")
		httpLog.Fatal(server.ListenAndServeTLS(htt.Certfile, htt.Keyfile))
	} else {
		httpLog.WithField("address", htt.Address).Info("Starting INSECURE http receiver (no TLS)")
		httpLog.Fatal(server.ListenAndServe())
	}
	return fmt.Errorf("unreachable")
}

// verifyPeerCertificate verifies a client certificate presented to us
// during TLS handshake by comparing its extensions (such as SAN) to
// some expected value(s)
func (auth *HTTPAuth) verifyPeerCertificate(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	if auth.SkipCertificateVerify || auth.SANDNSName == "" {
		httpLog.WithFields(log.Fields{"skip": auth.SkipCertificateVerify, "dns_name": auth.SANDNSName}).Trace("Skipping verifying certificate")
		return nil
	}

	certLogger := httpLog
	certLogger.Tracef("Verifying %d certificate chains", len(verifiedChains))
	for _, chain := range verifiedChains {
		if len(chain) < 1 {
			// Chain has no certificates (?)
			continue
		}
		cert := chain[0]
		certDebugLogger := certLogger.WithFields(log.Fields{
			"issuer":           cert.Issuer,
			"subject":          cert.Subject,
			"num_dns_names":    len(cert.DNSNames),
			"num_emails":       len(cert.EmailAddresses),
			"num_ip_addresses": len(cert.IPAddresses),
			"num_uris":         len(cert.URIs),
			"num_x509_ext":     len(cert.Extensions),
		})
		certDebugLogger.Trace("Verifying certificate")
		for _, dnsName := range cert.DNSNames {
			if auth.SANDNSName != "" && strings.ToLower(dnsName) == strings.ToLower(auth.SANDNSName) {
				// If we find a matching DNS name in the SANs, we return non-error
				// which specifies that we're done verifying with access granted.
				return nil
			}
		}
	}
	// If no checks until now have returned a non-error,
	// we return an error to tell the verifying function that this certificate
	// is not verified, and access is denied.
	// This will present the user with a 'bad certificate' alert.
	return fmt.Errorf("failed to verify x509 SAN DNS Name")
}

// Verify verifies the configuration for the HTTP receiver
func (htt *HTTP) Verify() error {
	if htt.Handlers == nil || len(htt.Handlers) == 0 {
		return skogul.MissingArgument("Handlers")
	}

	if htt.Address == "" {
		httpLog.Warn("Missing listen address for http receiver, using Go default")
	}

	if htt.Certfile == "" && htt.Auth != nil {
		httpLog.Warn("HTTP receiver configured with authentication but not with TLS! Auth will happen in the open!")
	}
	if (htt.Certfile != "" && htt.Keyfile == "") || (htt.Certfile == "" && htt.Keyfile != "") {
		return fmt.Errorf("Specify both Certfile AND Keyfile or none at all")
	}
	cas, err := loadClientCertificateCAs(htt.ClientCertificateCAs)
	if err != nil {
		return fmt.Errorf("unable to load client certificate CAs: %w", err)
	}
	for _, auth := range htt.Auth {
		if auth.Username != "" && auth.Password == "" {
			return fmt.Errorf("Username specified but no password.")
		}
		if auth.Username == "" && auth.Password != "" {
			return fmt.Errorf("Password specified but no username.")
		}
		if auth.SANDNSName != "" && cas == nil {
			return fmt.Errorf("No Client Certificate CAs defined, but DNS Name for SAN specified. Specify ClientCertificateCAs configuration element.")
		}
	}

	return nil
}

// GetStats exposes stats about the HTTP receiver.
func (htt *HTTP) GetStats() *skogul.Metric {
	now := skogul.Now()
	httpLog.WithField("time", now).Trace("Getting stats")
	metric := skogul.Metric{
		Time:     &now,
		Metadata: make(map[string]interface{}),
		Data:     make(map[string]interface{}),
	}

	metric.Metadata["component"] = "receiver"
	metric.Metadata["type"] = "HTTP"
	metric.Metadata["identity"] = skogul.Identity[htt]

	metric.Data["received"] = htt.stats.Received
	metric.Data["no_data"] = htt.stats.NoData
	metric.Data["read_failed"] = htt.stats.ReadFailed
	metric.Data["handler_errors"] = htt.stats.HandlerErrors
	metric.Data["sent"] = htt.stats.Sent

	return &metric
}
