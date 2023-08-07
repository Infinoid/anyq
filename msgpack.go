package main

import "github.com/vmihailenco/msgpack/v5"

type MsgPackFormat struct {
}

func (f *MsgPackFormat) GetExtensions() []string {
	return []string{
		// https://github.com/msgpack/msgpack/issues/291#issuecomment-1370526984, sure, why not
		"msgpack",
		"mpk",
	}
}

func (f *MsgPackFormat) GetFeatures() FormatFeatures {
	return FormatFeatures{
		Can_input:       true,
		Can_output:      true,
		Can_prettyprint: false,
		Can_scalar:      true,
		Can_array:       true,
		Can_object:      true,
		Is_binary:       true,
		Arbitrary_tree:  true,
		Universal:       true,
	}
}

func (f *MsgPackFormat) Input(b []byte) (any, error) {
	var data any
	err := msgpack.Unmarshal(b, &data)
	return data, err
}

func (f *MsgPackFormat) Output(a any, _ bool) ([]byte, error) {
	rv, err := msgpack.Marshal(a)
	return rv, err
}
