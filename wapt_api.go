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
	Success     bool
	Msg         string
	RequestTime float64 `json:"request_time"`
}

type WaptPingResponse struct {
	*WaptResponse
	Result struct {
		Version string
	}
}

type WaptHost struct {
	Uuid        string
	HostStatus  string `json:"host_status"`
	Reachable   string
	WaptVersion string `json:"wapt_version"`
}

type WaptHostsResponse struct {
	*WaptResponse
	Result []WaptHost
}

type WaptPackage struct {
	Package string
	Version string
}

type WaptPackagesResponse struct {
	*WaptResponse
	Result []WaptPackage
}

func waptPing(endpoint string) float64 {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to parse endpoint")
		return -1
	}

	u = u.JoinPath("/ping")
	log.Debug().Str("url", u.String()).Msg("Sending API request")
	r, err := http.Get(u.String())
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to ping")
		return -1
	}

	var pingResponse WaptPingResponse
	err = json.NewDecoder(r.Body).Decode(&pingResponse)
	if err != nil {
		log.Error().Err(err).Str("url", u.String()).Msg("Failed to read ping response")
		return -1
	}

	if !pingResponse.Success {
		log.Warn().Msg("Got ping response but success = false")
		return -1
	}

	log.Debug().Float64("time", pingResponse.RequestTime).Msg("Got ping response")
	return pingResponse.RequestTime
}

func isWaptUp(endpoint string) float64 {
	if waptPing(endpoint) > 0 {
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
