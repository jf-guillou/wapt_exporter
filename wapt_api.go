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
	Success      bool
	Msg          string
	Request_time float64
}

type WaptPingResponse struct {
	*WaptResponse
	Result struct {
		Version string
	}
}

type WaptHost struct {
	Uuid       string
	HostStatus string
	Reachable  string
}

type WaptHostsResponse struct {
	*WaptResponse
	Result []WaptHost
}

func waptPing(endpoint string) float64 {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to parse endpoint")
		return -1
	}

	u = u.JoinPath("/ping")
	r, err := http.Get(u.String())
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to ping")
		return -1
	}

	var pingResponse WaptPingResponse
	err = json.NewDecoder(r.Body).Decode(&pingResponse)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to read ping response")
		return -1
	}

	if !pingResponse.Success {
		log.Warn().Str("endpoint", endpoint).Msg("Got ping response but success = false")
		return -1
	}

	log.Debug().Str("endpoint", endpoint).Float64("time", pingResponse.Request_time).Msg("Got ping response")
	return pingResponse.Request_time
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
	r, err := http.Get(u.String())
	query := u.Query()
	// Max 10k hosts
	query.Add("limit", fmt.Sprintf("%d", MAX_HOSTS))
	u.RawQuery = query.Encode()
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to get hosts")
		return nil
	}

	var hostsResponse WaptHostsResponse
	err = json.NewDecoder(r.Body).Decode(&hostsResponse)
	if err != nil {
		log.Error().Err(err).Str("endpoint", endpoint).Msg("Failed to read hosts response")
		return nil
	}

	if !hostsResponse.Success {
		log.Warn().Str("endpoint", endpoint).Msg("Got hosts response but success = false")
		return nil
	}

	return &hostsResponse
}
