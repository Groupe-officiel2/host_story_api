package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/ovh/go-ovh/ovh"
)

func getOVHCredentials() (string, string, string) {
	return os.Getenv("OVH_APPLICATION_KEY"),
		os.Getenv("OVH_APPLICATION_SECRET"),
		os.Getenv("OVH_CONSUMER_KEY")
}

func newOVHClient() *ovh.Client {
	applicationKey, applicationSecret, consumerKey := getOVHCredentials()

	client, err := ovh.NewClient(
		"ovh-eu",
		applicationKey,
		applicationSecret,
		consumerKey,
	)
	if err != nil {
		return nil
	}

	return client
}

func srvRecordExists(client *ovh.Client, zone, subDomain, expectedTarget string, expectedTTL int) bool {
	var recordIDs []int64
	path := fmt.Sprintf("/domain/zone/%s/record?fieldType=SRV&subDomain=%s",
		zone, url.QueryEscape(subDomain))
	if err := client.Get(path, &recordIDs); err != nil {
		return false
	}

	for _, id := range recordIDs {
		var rec struct {
			Target string `json:"target"`
			TTL    int    `json:"ttl"`
		}
		if err := client.Get("/domain/zone/"+zone+"/record/"+fmt.Sprintf("%d", id),
			&rec); err != nil {
			continue
		}

		if rec.Target == expectedTarget && rec.TTL == expectedTTL {
			return true
		}
	}

	return false
}

func AddSRVRecord(port int, name string) {
	const (
		priority = 10
		weight   = 10
		ttl      = 300
		target   = "hoststory.fr."
		zoneEnv  = "OVH_DNS_ZONE"
		zoneDef  = "hoststory.fr"
	)

	if name == "" {
		return
	}

	client := newOVHClient()
	if client == nil {
		return
	}

	zone := os.Getenv(zoneEnv)
	if zone == "" {
		zone = zoneDef
	}

	subDomain := fmt.Sprintf("_vintagestory._tcp.%s", name)

	srvValue := fmt.Sprintf("%d %d %d %s", priority, weight, port, target)
	if srvRecordExists(client, zone, subDomain, srvValue, ttl) {
		fmt.Println("DNS record already exists")
		return
	}

	params := map[string]any{
		"fieldType": "SRV",
		"subDomain": subDomain,
		"target":    srvValue,
		"ttl":       ttl,
	}

	var recordResp struct {
		ID int64 `json:"id"`
	}
	if err := client.Post("/domain/zone/"+zone+"/record", params, &recordResp); err != nil {
		return
	}

	if err := client.Post("/domain/zone/"+zone+"/refresh", nil, nil); err != nil {
		return
	}
}

func DeleteSRVRecord(name string) {
	const (
		zoneEnv = "OVH_DNS_ZONE"
		zoneDef = "hoststory.fr"
	)

	if name == "" {
		return
	}

	client := newOVHClient()
	if client == nil {
		return
	}

	zone := os.Getenv(zoneEnv)
	if zone == "" {
		zone = zoneDef
	}

	subDomain := fmt.Sprintf("_vintagestory._tcp.%s", name)
	path := fmt.Sprintf("/domain/zone/%s/record?fieldType=SRV&subDomain=%s",
		zone, url.QueryEscape(subDomain))

	var recordIDs []int64
	if err := client.Get(path, &recordIDs); err != nil {
		return
	}

	if len(recordIDs) == 0 {
		return
	}

	recordID := recordIDs[0]
	if err := client.Delete("/domain/zone/"+zone+"/record/"+fmt.Sprintf("%d",
		recordID), nil); err != nil {
		return
	}

	if err := client.Post("/domain/zone/"+zone+"/refresh", nil, nil); err != nil {
		return
	}
}
