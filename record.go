package ddns

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"golang.org/x/sync/errgroup"
)

const (
	AutomaticTTL uint = 1
	DefaultTTL   uint = 120
)

// RecordType identifies the supported DNS record type from Cloudflare API.
// For more informations: https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record
type RecordType int

const (
	A RecordType = iota
	AAAA
	CNAME
	TXT
	SRV
	LOC
	MX
	NS
	SPF
)

var recordTypeIndices = []int{0, 1}

func (r RecordType) String() string {
	if r < A || r > SPF {
		return fmt.Sprintf("RecordType(%d)", r)
	}
	return "AAAAACNAMETXTSRVLOCMXNSSPF"[r : r+1]
}

var backgroundContext = context.Background()

// Record identifies a DNS Record for Cloudflare API.
// For more informations: https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record
type Record struct {
	ID      string     `json:"id",toml:"id"`
	Type    RecordType `json:"type",toml:"type"`
	Name    *url.URL   `json:"name",toml:"name"`
	Content string     `json:"content",toml:"content"`

	// Optional parameters

	TTL     uint `json:"ttl,omitempty",toml:"ttl,omitempty"`
	Proxied bool `json:"proxied,omitempty",toml:"proxied,omitempty"`
}

// Validate checks for errors in the DNS Record object, and returns them
// in case at least one have been found.
func (r Record) Validate() error {
	group, _ := errgroup.WithContext(backgroundContext)
	group.Go(func() error {
		if r.ID == "" {
			return ErrEmptyIDUnsupported
		}
		return nil
	})
	group.Go(func() error {
		if r.Type < A || r.Type > SPF {
			return fmt.Errorf("invalid DNS record type, %s", r.Type)
		}
		return nil
	})
	group.Go(func() error {
		if r.Name == nil {
			return ErrEmptyRecordName
		}
		return nil
	})
	group.Go(func() error {
		if r.Content == "" {
			return ErrEmptyRecordContent
		}
		return nil
	})
	return group.Wait()
}

// Copy is a copy-constructor that changes only the content field
// of the new copied Record object.
func (r Record) Copy(content string) Record {
	r.Content = content
	return r
}

var (
	// ErrEmptyIDUnsupported happens when an empty ID in the DNS Record
	// has been found. This is not supported yet!
	ErrEmptyIDUnsupported = errors.New("Empty DNS record id is not supported yet")

	// ErrEmptyRecordName is a DNS Record validation error, when the Name field
	// is not setted correctly.
	ErrEmptyRecordName = errors.New("DNS record name url is empty")

	// ErrEmptyRecordContent is a DNS Record validation error, when the Content
	// field is not setted correctly.
	ErrEmptyRecordContent = errors.New("DNS record content is empty")
)
