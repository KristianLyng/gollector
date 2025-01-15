package parser_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/telenornms/skogul/gen/usp"
	"github.com/telenornms/skogul/parser"
)

func readFile(file string, t *testing.T) []byte {
	t.Helper()
	b := make([]byte, 9000)
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("unable to open protobuf packet file: %v", err)
	}
	defer f.Close()
	n, err := f.Read(b)
	if err != nil {
		t.Fatalf("unable to read protobuf packet file: %v", err)
	}
	if n == 0 {
		t.Fatalf("read 0 bytes from protobuf packet file....")
	}
	return b[0:n]
}

func TestUSPParseFile(t *testing.T) {
	d := readFile("testdata/usp.bin", t)

	x := parser.USPParser{}
	container, err := x.Parse(d)
	if err != nil {
		t.Error("Error while parsing", err)
	}

	if container.Metrics[0].Metadata == nil || container.Metrics[0].Data == nil {
		t.Error("Either metadata or data is missing")
	}
}

func TestUSPGetUspRecord(t *testing.T) {
	d := readFile("testdata/usp.bin", t)

	unmarshaledMessage := &usp.Record{}
	if err := proto.Unmarshal(d, unmarshaledMessage); err != nil {
		t.Error("Error while unmarshalling data", err)
	}
}

func TestUSPGetRecordMsgPayload(t *testing.T) {
	msgPayload := &usp.Msg{}

	d := readFile("testdata/usp.bin", t)

	unmarshaledMessage := &usp.Record{}
	if err := proto.Unmarshal(d, unmarshaledMessage); err != nil {
		t.Error("Error while unmarshalling record", err)
	}

	payload := unmarshaledMessage.GetNoSessionContext().GetPayload()
	if err := proto.Unmarshal(payload, msgPayload); err != nil {
		t.Error("Failed to unmarshall record payload")
	}
}

func TestUSPExtractJSON(t *testing.T) {
	msgPayload := &usp.Msg{}

	d := readFile("testdata/usp.bin", t)

	unmarshaledMessage := &usp.Record{}
	if err := proto.Unmarshal(d, unmarshaledMessage); err != nil {
		t.Error("Error while unmarshalling record", err)
	}

	payload := unmarshaledMessage.GetNoSessionContext().GetPayload()
	if err := proto.Unmarshal(payload, msgPayload); err != nil {
		t.Error("Failed to unmarshall record payload")
	}

	input := []byte(msgPayload.Body.GetRequest().GetNotify().GetEvent().GetParams()["Data"])

	var k map[string]interface{}

	if err := json.Unmarshal(input, &k); err != nil {
		t.Error("Failed to unmarshall json")
	}
}
