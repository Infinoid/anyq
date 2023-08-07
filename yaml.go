package main

import (
	yaml "gopkg.in/yaml.v3"
)

type YAMLFormat struct {
}

func (f *YAMLFormat) GetExtensions() []string {
	return []string{
		"yaml",
		"yml",
	}
}

func (f *YAMLFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: true, // indentation is structural in yaml, so technically you can't *not* indent it
		Can_scalar:      true,
		Can_array:       true,
		Can_object:      true,
		Is_binary:       false,
		Arbitrary_tree:  true,
		Universal:       true,
	}
}

func (f *YAMLFormat) Input(b []byte) (any, error) {
	var data any
	err := yaml.Unmarshal(b, &data)
	return data, err
}

func (f *YAMLFormat) Output(a any, _ bool) ([]byte, error) {
	rv, err := yaml.Marshal(a)
	return rv, err
}
