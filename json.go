package main

import (
	"bytes"
	"encoding/json"
)

type JSONFormat struct {
}

func (f *JSONFormat) GetExtensions() []string {
	return []string{
		"json",
		"js",
	}
}

func (f *JSONFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: true,
		Can_scalar:      true,
		Can_array:       true,
		Can_object:      true,
		Is_binary:       false,
		Arbitrary_tree:  true,
		Universal:       true,
	}
}

func (f *JSONFormat) Input(b []byte) (any, error) {
	var data any
	err := json.Unmarshal(b, &data)
	return data, err
}

func (f *JSONFormat) Output(a any, prettyprint bool) ([]byte, error) {
	b := bytes.NewBuffer([]byte{})
	e := json.NewEncoder(b)
	if prettyprint {
		e.SetIndent("", "  ")
	}
	err := e.Encode(a)
	return b.Bytes(), err
}
