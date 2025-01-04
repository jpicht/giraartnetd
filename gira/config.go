package gira

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Server    string   `json:"server"`
	IgnoreSSL bool     `json:"ignore_ssl"`
	ClientID  string   `json:"client_id"`
	User      string   `json:"user"`
	Password  string   `json:"password"`
	Token     string   `json:"token"`
	ArtNet    ArtNet   `json:"artnet"`
	UIDs      []string `json:"uids"`
	AutoOnOff bool     `json:"auto_on_off"`
}

type ArtNet struct {
	Network string `json:"network"`
	Net     int    `json:"net"`
	SubUni  int    `json:"sub_uni"`
}

func LoadConfig(r io.Reader) (*Config, error) {
	return Load[Config](r)
}

func LoadConfigFile(f string) (*Config, error) {
	return LoadFile[Config](f)
}

func Load[T any](r io.Reader) (*T, error) {
	c := new(T)
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return c, dec.Decode(c)
}

func LoadFile[T any](f string) (*T, error) {
	fp, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return Load[T](fp)
}
