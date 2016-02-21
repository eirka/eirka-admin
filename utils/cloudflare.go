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

type CloudFlareBanIpPayload struct {
	Mode          string                       `json:"mode"`
	Notes         string                       `json:"notes"`
	Configuration CloudFlareBanIpConfiguration `json:"configuration"`
}

type CloudFlareBanIpConfiguration struct {
	Target string `json:"target"`
	Value  string `json:"value"`
}

func CloudFlareBanIp(ip, reason string) (err error) {

	if len(ip) == 0 {
		return errors.New("no ip provided")
	}

	// block ip request json
	data := CloudFlareBanIpPayload{
		Mode: "block",
		Configuration: CloudFlareBanIpConfiguration{
			Target: "ip",
			Value:  ip,
		},
		Notes: reason,
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}

	// api endpoint
	cloudflareUrl := &url.URL{
		Scheme: "https",
		Host:   "api.cloudflare.com",
		Path:   "/client/v4/user/firewall/access_rules/rules",
	}

	// our http request
	req, err := http.NewRequest(http.MethodPost, cloudflareUrl.String(), bytes.NewReader(payloadBytes))
	if err != nil {
		return errors.New("Error creating CloudFlare request")
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
		return errors.New("Error reaching CloudFlare")
	}

	return
}
