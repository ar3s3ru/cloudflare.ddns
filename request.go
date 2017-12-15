package ddns

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	publicIPURL = "http://myexternalip.com/raw"
)

func getPublicIP(cli *http.Client) (string, error) {
	if cli == nil {
		return "", ErrNilHTTPClient
	}

	resp, err := cli.Get(publicIPURL)
	if err != nil {
		return "", err
	}
	if code := resp.StatusCode; code != http.StatusOK {
		return "", fmt.Errorf("http request for public ip failed (code %d)", code)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes), nil // TODO: maybe use a net.IP?
}

// Status identifies the status of a record refresh iteration.
type Status uint

const (
	Unchanged Status = iota
	Success
	APIError
	Error
)

// Result is the result of a DDNS refresh request.
type Result struct {
	Status
	Error error
}

type requestCtx struct {
	conf    APIConfig
	last    Record
	timeout time.Duration
}

func (r *requestCtx) do(ctx context.Context, cli *http.Client) Result {
	ip, err := getPublicIP(cli) // Retrieve last public IP
	if err != nil {
		return Result{Status: Error, Error: err}
	}
	if ip == r.last.Content { // Check last IP with actual IP
		return Result{Status: Unchanged}
	}
	latest := r.last.Copy(ip)               // New public IP, create request
	req, err := r.conf.Request(ctx, latest) // Passing context to stop request
	if err != nil {
		return Result{Status: Error, Error: err}
	}
	res, err := cli.Do(req)
	if err != nil {
		return Result{Status: Error, Error: err}
	}
	defer res.Body.Close()
	// TODO: handle error case
	r.last = latest
	return Result{Status: Success}
}

func (r *requestCtx) handle(ctx context.Context, cli *http.Client) <-chan Result {
	ch, tick := make(chan Result), time.NewTicker(r.timeout)
	go func(r *requestCtx) {
		defer close(ch)
		for {
			ch <- r.do(ctx, cli)
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
			}
		}
	}(r)
	return ch
}

// NewRequest takes API configuration, DNS record to keep update, and optional parameters
// and returns a channel that will emit Results every time a refresh request will be done.
// A context.Context can be used to cancel the execution of the refresh.
func NewRequest(ctx context.Context, config APIConfig, record Record,
	opts ...Option) (<-chan Result, error) {

	if err := record.Validate(); err != nil {
		return nil, err
	}
	apply := &Options{timeout: 120 * time.Second, cli: http.DefaultClient}
	for _, opt := range opts {
		if opt != nil {
			opt(apply)
		}
	}
	r := &requestCtx{conf: config, last: record, timeout: apply.timeout}
	if r.last.TTL == 0 {
		r.last.TTL = AutomaticTTL // FIXME: don't know if this is correct...
	}
	return r.handle(ctx, apply.cli), nil
}

var (
	// ErrNilHTTPClient happens if the http Client used
	// to make http requests is nil.
	ErrNilHTTPClient = errors.New("http client specified is nil")
)
