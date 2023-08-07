package main

import (
	"bytes"

	"github.com/pelletier/go-toml/v2"
)

type TOMLFormat struct {
}

func (f *TOMLFormat) GetExtensions() []string {
	return []string{
		"toml",
	}
}

func (f *TOMLFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: true,
		Can_scalar:      false,
		Can_array:       false, // can emit arrays but not parse them
		Can_object:      true,
		Is_binary:       false,
		Arbitrary_tree:  true,
		Universal:       true,
	}
}

func (f *TOMLFormat) Input(b []byte) (any, error) {
	var data any
	err := toml.Unmarshal(b, &data)
	return data, err
}

func (f *TOMLFormat) Output(a any, prettyprint bool) ([]byte, error) {
	b := bytes.NewBuffer([]byte{})
	e := toml.NewEncoder(b)
	if prettyprint {
		e.SetIndentTables(true)
	}
	err := e.Encode(a)
	return b.Bytes(), err
}
