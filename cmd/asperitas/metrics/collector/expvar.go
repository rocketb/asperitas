package collector

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

// Expvar provides the ability to receive metrics
// from internal services using expvar.
type Expvar struct {
	host   string
	tr     *http.Transport
	client http.Client
}

// New creates new Expvar for metrics collection.
func New(host string) (*Expvar, error) {
	tr := http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	return &Expvar{
		host: host,
		tr:   &tr,
		client: http.Client{
			Transport: &tr,
			Timeout:   1 * time.Second,
		},
	}, nil
}

// Collect capture metrics on the addr configure to this endpoint.
func (exp *Expvar) Collect() (map[string]any, error) {
	req, err := http.NewRequest(http.MethodGet, exp.host, nil)
	if err != nil {
		return nil, err
	}

	resp, err := exp.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, errors.New(string(msg))
	}

	data := make(map[string]any)
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
