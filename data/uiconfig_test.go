package data_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"

	"github.com/jpicht/giraartnetd/data"
	"github.com/stretchr/testify/require"
)

//go:embed uiconfig.json
var uiconfig_json []byte

func TestUIConfig(t *testing.T) {
	dec := json.NewDecoder(bytes.NewBuffer(uiconfig_json))
	dec.DisallowUnknownFields()
	var cfg data.UIConfig
	require.NoError(t, dec.Decode(&cfg))
}
