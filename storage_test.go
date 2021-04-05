package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

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
	go rr.Step(stone, rr.Player1)
	go rr.Step(scissors, rr.Player1)
	go rr.Step(paper, rr.Player1)
	go rr.Step(stone, rr.Player2)
	go rr.Step(scissors, rr.Player2)
	go rr.Step(paper, rr.Player2)
	rr.mx.Unlock()

	<-time.After(time.Microsecond)
	rr.mx.Lock()
	rr.mx.Unlock()
	res := rr.Result(rr.Player1)
	t.Logf("Result for first user aster game: %s \n", res)

}
