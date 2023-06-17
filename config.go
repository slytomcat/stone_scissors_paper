package main

import (
	"errors"
	"os"
	"strings"
)

type config struct {
	HostPort      string   `default:"localhost:8080"`
	RedisAddrs    []string `required:"true"`
	ServerSalt    string   `required:"true"`
	RedisPassword string
}

const (
	defaultHostPort = "localhost:8080"
)

func newConfig() (*config, error) {
	cfg := config{
		HostPort:      defaultHostPort,
		RedisAddrs:    []string{},
		ServerSalt:    "",
		RedisPassword: "",
	}
	val, ok := os.LookupEnv("SSP_HOST_PORT")
	if ok && len(val) > 0 {
		cfg.HostPort = val
	}
	val, ok = os.LookupEnv("SSP_REDIS_PASSWORD")
	if ok {
		cfg.RedisPassword = val
	}
	val, ok = os.LookupEnv("SSP_REDIS_ADDRS")
	if ok && len(val) > 0 {
		cfg.RedisAddrs = strings.Split(val, ",")
	} else {
		return nil, errors.New("Environment variable SSP_REDIS_ADDRS is not defined")
	}
	val, ok = os.LookupEnv("SSP_SERVER_SALT")
	if ok && len(val) > 0 {
		cfg.ServerSalt = val
	} else {
		return nil, errors.New("Environment variable SSP_SERVER_SALT is not defined")
	}
	return &cfg, nil
}
