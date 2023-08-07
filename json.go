package main

import "encoding/json"

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
		Can_input:      true,
		Can_output:     true,
		Arbitrary_tree: true,
		Opinionated:    true,
	}
}

func (f *JSONFormat) Input(b []byte) (DataNode, error) {
	data := DataNode{}
	err := json.Unmarshal(b, data)
	return data, err
}
