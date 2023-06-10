package main

import (
	"testing"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/stretchr/testify/require"
)

func Test1_Storage(t *testing.T) {

	config := configT{}
	_, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	require.Error(t, err)

	envSet(t) // load .env file for test environment

	require.NoError(t, envconfig.Process("SSP", &config))

	db, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	require.NoError(t, err)

	r := NewRound("u1", "u2")

	require.NoError(t, db.Store(r))

	rr, _ := db.Retrieve(r.ID)

	require.Equal(t, rr, r)

	t.Logf("Stored: %+v \n retrieved: %+v \n", r, rr)

	_, err = db.Retrieve("Non-existing_key")
	require.Error(t, err)
}
