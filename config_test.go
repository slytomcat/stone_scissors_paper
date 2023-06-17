package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigOk(t *testing.T) {
	t.Setenv("SSP_HOST_PORT", "localhost:1234")
	t.Setenv("SSP_REDIS_PASSWORD", "some.password")
	t.Setenv("SSP_REDIS_ADDRS", "some.redis.adr:1234")
	t.Setenv("SSP_SERVER_SALT", "some.salt")
	cfg, err := newConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestConfigFail(t *testing.T) {
	t.Setenv("SSP_REDIS_ADDRS", "")
	t.Setenv("SSP_SERVER_SALT", "some.salt")
	cfg, err := newConfig()
	require.Error(t, err)
	require.Nil(t, cfg)
	t.Setenv("SSP_REDIS_ADDRS", "some.redis.adr:1234")
	t.Setenv("SSP_SERVER_SALT", "")
	cfg, err = newConfig()
	require.Error(t, err)
	require.Nil(t, cfg)
}
