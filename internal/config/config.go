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
	StreamSecret   string
	TokenMintSecret string
	InstanceID     string
	Bus            string
	RedisURL       string
	NATSURL        string
	ReplayCapacity int
	SubscriberBuf  int
	DefaultTokenTTL int // seconds
}

// FromEnv reads configuration with safe local defaults.
func FromEnv() Config {
	streamSecret := strings.TrimSpace(os.Getenv("STREAM_SECRET"))
	mint := strings.TrimSpace(os.Getenv("TOKEN_MINT_SECRET"))
	hook := strings.TrimSpace(os.Getenv("HOOK_SECRET"))
	if mint == "" {
		mint = hook
	}
	if mint == "" {
		mint = streamSecret
	}
	return Config{
		ListenAddr:      getenv("LISTEN_ADDR", "127.0.0.1:8080"),
		HookSecret:      hook,
		StreamSecret:    streamSecret,
		TokenMintSecret: mint,
		InstanceID:      getenv("INSTANCE_ID", "relay-1"),
		Bus:             getenv("BUS", "memory"),
		RedisURL:        strings.TrimSpace(os.Getenv("REDIS_URL")),
		NATSURL:         getenv("NATS_URL", "nats://127.0.0.1:4222"),
		ReplayCapacity:  getenvInt("REPLAY_CAPACITY", 64),
		SubscriberBuf:   getenvInt("SUBSCRIBER_BUFFER", 16),
		DefaultTokenTTL: getenvInt("TOKEN_TTL_SECONDS", 3600),
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
