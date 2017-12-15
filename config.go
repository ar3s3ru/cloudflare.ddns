package ddns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	cloudflareURL    = "https://api.cloudflare.com/client/v4"
	apiEmailHeader   = "X-Auth-Email"
	apiAuthKeyHeader = "X-Auth-Key"
)

// APIConfig represent the configuration for CloudFlare API.
type APIConfig struct {
	Email  string
	APIKey string
	ZoneID string
}

// APIDNSPath returns the url path to which make the API request.
func (conf APIConfig) APIDNSPath(id string) string {
	return fmt.Sprintf("%s/zones/%s/dns_recors/%s", cloudflareURL, conf.ZoneID, id)
}

// Request creates an http Request from a Record object.
// The generated request can be used to update the DNS record.
// A context.Context is used in order to cancel the http Request, eventually.
func (conf APIConfig) Request(ctx context.Context, body Record) (*http.Request, error) {
	if err := body.Validate(); err != nil {
		return nil, err
	}
	method := http.MethodPost
	if body.ID != "" { // Using an already-existing record
		method = http.MethodPut
	}
	bytz, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, conf.APIDNSPath(body.ID), bytes.NewBuffer(bytz))
	if err == nil {
		req.Header.Add(apiEmailHeader, conf.Email)
		req.Header.Add(apiAuthKeyHeader, conf.APIKey)
		req = req.WithContext(ctx) // In order to stop the request eventually
	}
	return req, err
}
