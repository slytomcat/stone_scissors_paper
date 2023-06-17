package main

import (
	"testing"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
)

func Test1_Storage(t *testing.T) {

	_, err := NewDatabase(redis.UniversalOptions{Addrs: []string{"wrong.adr:000"}, Password: ""})
	require.Error(t, err)

	envSet(t) // load .env file for local test environment

	config, err := newConfig()

	db, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	require.NoError(t, err)

	r := NewRound("u1")

	require.NoError(t, db.Store(r))

	rr, _ := db.Retrieve(r.ID)

	require.Equal(t, rr, r)

	_, err = db.Retrieve("Non-existing_key")
	require.Error(t, err)
}
