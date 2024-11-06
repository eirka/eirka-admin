package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/eirka/eirka-libs/config"
)

type cloudFlareBanIPPayload struct {
	Mode          string                       `json:"mode"`
	Notes         string                       `json:"notes"`
	Configuration cloudFlareBanIPConfiguration `json:"configuration"`
}

type cloudFlareBanIPConfiguration struct {
	Target string `json:"target"`
	Value  string `json:"value"`
}

// CloudFlareBanIP will query the CloudFlare API and add the IP to ban to all zones
func CloudFlareBanIP(ip, reason string) (err error) {

	// noop if cloudflare is not configured
	if !config.Settings.CloudFlare.Configured {
		return
	}

	if len(ip) == 0 {
		return errors.New("no ip provided")
	}

	// block ip request json
	data := cloudFlareBanIPPayload{
		Mode: "block",
		Configuration: cloudFlareBanIPConfiguration{
			Target: "ip",
			Value:  ip,
		},
		Notes: reason,
	}

	payloadBytes, _ := json.Marshal(data)

	// api endpoint
	cloudflareURL := &url.URL{
		Scheme: "https",
		Host:   "api.cloudflare.com",
		Path:   "/client/v4/user/firewall/access_rules/rules",
	}

	// our http request
	req, err := http.NewRequest(http.MethodPost, cloudflareURL.String(), bytes.NewReader(payloadBytes))
	if err != nil {
		return errors.New("error creating cloudflare request")
	}

	req.Header.Set("X-Auth-Email", config.Settings.CloudFlare.Email)
	req.Header.Set("X-Auth-Key", config.Settings.CloudFlare.Key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Eirka/1.2")

	// a client with a timeout
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	// do the request
	// TODO: add errors here to a system log
	_, err = netClient.Do(req)
	if err != nil {
		return errors.New("error reaching cloudflare")
	}

	return
}
