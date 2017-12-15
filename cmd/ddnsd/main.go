package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ar3s3ru/cloudflare.ddns"
)

const DefaultConfigPath = "./config.toml"

var stash *log.Logger

func init() {
	stash = log.New(os.Stderr, "[cloudflare.ddnsd] ", log.LstdFlags)
}

var (
	apiKey     = flag.String("k", "", "CloudFlare API Key to use")
	emailKey   = flag.String("e", "", "CloudFlare Email Auth to use")
	zoneID     = flag.String("z", "", "CloudFlare DNS Zone ID")
	recordFile = flag.String("r", DefaultConfigPath, "specify the path of the record file to update")
	ticker     = flag.Duration("t", 120*time.Second, "Refresh duration")
)

type Config struct {
	ddns.Record
}

func main() {
	flag.Parse()

	if recordFile == nil || *recordFile == "" {
		stash.Fatalf("no record file specified...\n")
	}

	file, err := os.Open(*recordFile)
	if err != nil {
		stash.Fatalf("opening record file failed: %s\n", err)
	}
	defer stash.Fatalln(file.Close())

	var config Config
	if _, err := toml.DecodeReader(file, &config); err != nil {
		stash.Fatalf("decoding record file failed: %s\n", err)
	}
	if err := config.Validate(); err != nil {
		stash.Fatalf("validating record file failed: %s\n", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	api := ddns.APIConfig{APIKey: *apiKey, Email: *emailKey, ZoneID: *zoneID}
	ch, err := ddns.NewRequest(ctx, api, config.Record, ddns.Timeout(*ticker))
	if err != nil {
		stash.Fatalf("creating new request failed: %s\n", err)
	}
	for res := range ch {
		stash.Printf("request{STATUS: %s, ERROR: %s}\n", res.Status, res.Error)
	}
	stash.Println("exiting...")
}
