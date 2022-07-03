package parser

import (
	"os"

	"github.com/hamba/avro"
	"github.com/telenornms/skogul"
)

type AVRO struct {
	Schema avro.Schema
	In     skogul.Container
}

// Parser accepts the byte buffer of GOB
func (x AVRO) Parse(b []byte) (*skogul.Container, error) {
	var A AVRO
	s, _ := os.ReadFile("./schema/avro_schema")
	A.Schema = avro.MustParse(string(s))
	var c skogul.Container
	err := avro.Unmarshal(A.Schema, b, &c)
	return &c, err

}
