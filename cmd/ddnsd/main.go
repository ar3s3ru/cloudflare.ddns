package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/ar3s3ru/cloudflare.ddns"
)

var stash *log.Logger

func init() {
	stash = log.New(os.Stderr, "[cloudflare.ddnsd] ", log.LstdFlags)
}

var (
	apiKey     = flag.String("k", "", "CloudFlare API Key to use")
	emailKey   = flag.String("e", "", "CloudFlare Email Auth to use")
	zoneID     = flag.String("z", "", "CloudFlare DNS Zone ID")
	recordFile = flag.String("r", "", "specify the path of the record file to update")
	ticker     = flag.Duration("t", 120*time.Second, "Refresh duration")
)

func main() {
	flag.Parse()

	if recordFile == nil || *recordFile == "" {
		stash.Fatalf("no record file specified...\n")
	}

	file, err := os.Open(*recordFile)
	if err != nil {
		stash.Fatalf("opening record file failed: %s\n", err)
	}

	var record ddns.Record
	if err := json.NewDecoder(file).Decode(&record); err != nil {
		stash.Fatalf("decoding record file failed: %s\n", err)
	}
	if err := record.Validate(); err != nil {
		stash.Fatalf("validating record file failed: %s\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := ddns.APIConfig{APIKey: *apiKey, Email: *emailKey, ZoneID: *zoneID}
	ch, err := ddns.NewRequest(ctx, config, record, ddns.Timeout(*ticker))
	if err != nil {
		stash.Fatalf("creating new request failed: %s\n", err)
	}
	for res := range ch {
		stash.Printf("request{STATUS: %s, ERROR: %s}\n", res.Status, res.Error)
	}
	stash.Println("exiting...")
}
