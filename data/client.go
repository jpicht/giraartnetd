package data

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"github.com/jpicht/giraartnetd/gira"
)

type Client interface {
	UIConfig() (*UIConfig, error)
	Get(uid string) (*ValueBody, error)
	Set(*ValueBody) error
}

func NewRESTClient(cfg gira.Config) (*RESTClient, error) {
	// FIXME: use local client
	if cfg.IgnoreSSL {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	URL, err := url.Parse(cfg.Server)
	if err != nil {
		return nil, err
	}

	q := URL.Query()
	q.Add("token", cfg.Token)
	URL.RawQuery = q.Encode()

	return &RESTClient{
		Config: cfg,
		URL:    URL,
	}, nil
}

type RESTClient struct {
	Config gira.Config
	URL    *url.URL
	UI     *UIConfig
}

func (rc *RESTClient) request(method string, path string, body io.Reader) (*http.Response, error) {
	URL := *rc.URL
	URL.Path = path

	req, err := http.NewRequest(method, URL.String(), body)
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

func (rc *RESTClient) UIConfig() (*UIConfig, error) {
	if rc.UI != nil {
		return rc.UI, nil
	}
	res, err := rc.request(http.MethodGet, "/api/uiconfig", nil)
	ui, err := gira.Load[UIConfig](res.Body)
	if err != nil {
		return nil, err
	}
	rc.UI = ui
	return rc.UI, nil
}

func (rc *RESTClient) Get(uid string) (*ValueBody, error) {
	res, err := rc.request(http.MethodGet, "/api/values/"+uid, nil)
	if err != nil {
		return nil, err
	}
	return gira.Load[ValueBody](res.Body)
}

func (rc *RESTClient) Set(values *ValueBody) error {
	body, err := json.Marshal(values)
	if err != nil {
		return err
	}

	res, err := rc.request(http.MethodPut, "/api/values", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	_, err = io.Copy(io.Discard, res.Body)
	if err != nil && errors.Is(err, io.EOF) {
		return err
	}
	return nil
}
