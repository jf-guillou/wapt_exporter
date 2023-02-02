package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

const MAX_HOSTS = 10000

type WaptResponse struct {
	Success     bool    `json:"success"`
	Msg         string  `json:"msg"`
	RequestTime float64 `json:"request_time"`
}

type WaptPingResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Version string
	} `json:"result"`
}

type WaptHost struct {
	Uuid        string `json:"uuid"`
	HostStatus  string `json:"host_status"`
	Reachable   string `json:"reachable"`
	WaptVersion string `json:"wapt_version"`
}

type WaptHostsResponse struct {
	*WaptResponse
	Result []WaptHost `json:"result"`
}

type WaptPackage struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

type WaptPackagesResponse struct {
	*WaptResponse
	Result []WaptPackage `json:"result"`
}

func waptPing(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to parse endpoint")
		return false
	}

	u = u.JoinPath("/ping")
	log.Debug().Str("url", u.String()).Msg("Sending API request")
	r, err := http.Get(u.String())
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to ping")
		return false
	}

	var pingResponse WaptPingResponse
	err = json.NewDecoder(r.Body).Decode(&pingResponse)
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to read ping response")
		return false
	}

	if !pingResponse.Success {
		log.Warn().Bool("success", pingResponse.Success).Msg("Got ping error response")
		return false
	}

	log.Debug().Bool("success", pingResponse.Success).Msg("Got ping response")
	return pingResponse.Success
}

func isWaptUp(endpoint string) float64 {
	if waptPing(endpoint) {
		return 1
	}
	return 0
}

func waptHosts(endpoint, username, password string) *WaptHostsResponse {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to parse endpoint")
		return nil
	}

	u = u.JoinPath("/api/v3/hosts")
	u.User = url.UserPassword(username, password)
	query := u.Query()
	// Max 10k hosts
	query.Add("limit", fmt.Sprintf("%d", MAX_HOSTS))
	u.RawQuery = query.Encode()
	log.Debug().Str("url", u.String()).Msg("Sending API request")
	r, err := http.Get(u.String())
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to get hosts")
		return nil
	}

	var hostsResponse WaptHostsResponse
	err = json.NewDecoder(r.Body).Decode(&hostsResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read hosts response")
		return nil
	}

	if !hostsResponse.Success {
		log.Warn().Msg("Got hosts response but success = false")
		return nil
	}

	log.Debug().Msg("Got hosts response")
	return &hostsResponse
}

func waptPackages(endpoint, username, password string) *WaptPackagesResponse {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to parse endpoint")
		return nil
	}

	u = u.JoinPath("/api/v3/packages")
	u.User = url.UserPassword(username, password)
	log.Debug().Str("url", u.String()).Msg("Sending API request")
	r, err := http.Get(u.String())
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to get packages")
		return nil
	}

	var packagesResponse WaptPackagesResponse
	err = json.NewDecoder(r.Body).Decode(&packagesResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read packages response")
		return nil
	}

	if !packagesResponse.Success {
		log.Warn().Msg("Got packages response but success = false")
		return nil
	}

	log.Debug().Msg("Got packages response")
	return &packagesResponse
}
