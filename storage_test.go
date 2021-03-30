package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-redis/redis"
)

type storage_config struct {
	ConnectOptions redis.UniversalOptions
}

func Test1_Storage(t *testing.T) {

	config := storage_config{}
	buf, err := ioutil.ReadFile("cnf.json")
	if err != nil {
		t.Error(err)
	}
	// parse config file
	err = json.Unmarshal(buf, &config)
	if err != nil {
		t.Error(err)
	}

	db, err := NewDatabse(config.ConnectOptions)
	if err != nil {
		t.Error(err)
	}

	r := NewRound()

	err = db.Store(r)
	if err != nil {
		t.Error(err)
	}

	rr, err := db.Retrive(r.ID)

	if rr.ID != r.ID ||
		rr.Bid1 != r.Bid1 ||
		rr.Bid2 != r.Bid2 ||
		rr.Player1 != r.Player1 ||
		rr.Player2 != r.Player2 ||
		rr.Winner != r.Winner {
		t.Error("Retrived not the same as stored")
	}

	t.Logf("Stored: %v \n retrived: %v \n", r, rr)

}
