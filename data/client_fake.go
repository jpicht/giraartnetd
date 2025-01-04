package data

import (
	"encoding/json"
	"os"
)

func NewFakeClient(ui *UIConfig) *fakeClient {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return &fakeClient{
		ui:     ui,
		values: make(map[string]string),
		enc:    enc,
	}
}

type fakeClient struct {
	ui     *UIConfig
	values map[string]string
	enc    *json.Encoder
}

func (fc *fakeClient) UIConfig() (*UIConfig, error) {
	return fc.ui, nil
}

func (fc *fakeClient) Get(uid string) (*ValueBody, error) {
	return &ValueBody{}, nil
}

func (fc *fakeClient) Set(b *ValueBody) error {
	fc.enc.Encode(b)
	return nil
}
