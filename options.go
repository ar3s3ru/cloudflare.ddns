package ddns

import (
	"net/http"
	"time"
)

// Options can be applied to the Request object.
// They'll be used for every API request made by the daemon.
type Options struct {
	timeout time.Duration
	cli     *http.Client
}

// Option identifies a single option that can be applied to the Options.
type Option func(opts *Options)

// Timeout specifies after how much the update must occur.
// Only durations > 0 will be applied.
func Timeout(d time.Duration) Option {
	return func(opts *Options) {
		if d > 0 {
			opts.timeout = d
		}
	}
}

// ClientHTTP specifies which http Client to use to make http requests.
// Only a non-nil client will be applied.
func ClientHTTP(cli *http.Client) Option {
	return func(opts *Options) {
		if cli != nil {
			opts.cli = cli
		}
	}
}
