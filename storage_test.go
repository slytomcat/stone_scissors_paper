package main

import (
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

func Test1_Storage(t *testing.T) {

	config := configT{}

	godotenv.Load() // load .env file for test environment

	err := envconfig.Process("SSP", &config)
	if err != nil {
		t.Error(err)
	}

	db, err := NewDatabse(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	if err != nil {
		t.Fatal(err)
	}

	r := NewRound("u1", "u2")

	err = db.Store(r)
	if err != nil {
		t.Error(err)
	}

	rr, err := db.Retrieve(r.ID)

	if rr.ID != r.ID ||
		rr.Bet1 != r.Bet1 ||
		rr.Bet2 != r.Bet2 ||
		rr.Player1 != r.Player1 ||
		rr.Player2 != r.Player2 ||
		rr.Winner != r.Winner {
		t.Error("Retrived not the same as stored")
	}

	t.Logf("Stored: %v \n retrived: %v \n", r, rr)

	// check mutex in retrived round
	rr.mx.Lock()
	go rr.Step("stone", "u1")
	go rr.Step("scissors", "u1")
	go rr.Step("paper", "u1")
	go rr.Step("stone", "u2")
	go rr.Step("scissors", "u2")
	go rr.Step("paper", "u2")
	rr.mx.Unlock()

	<-time.After(time.Microsecond)
	rr.mx.Lock()
	rr.mx.Unlock()
	res := rr.Result("u1")
	t.Logf("Result for first user aster game: %s \n", res)

}
