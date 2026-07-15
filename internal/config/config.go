// Package config loads process settings from the environment.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds runtime settings.
type Config struct {
	ListenAddr     string
	HookSecret     string
	ReplayCapacity int
	SubscriberBuf  int
}

// FromEnv reads configuration with safe local defaults.
func FromEnv() Config {
	return Config{
		ListenAddr:     getenv("LISTEN_ADDR", "127.0.0.1:8080"),
		HookSecret:     strings.TrimSpace(os.Getenv("HOOK_SECRET")),
		ReplayCapacity: getenvInt("REPLAY_CAPACITY", 64),
		SubscriberBuf:  getenvInt("SUBSCRIBER_BUFFER", 16),
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
