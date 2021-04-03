package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	redis "github.com/go-redis/redis"
)

type config struct {
	HostPort       string
	ConnectOptions redis.UniversalOptions
}

var (
	db      Database
	version string
)

func main() {
	err := doMain()
	if err != nil {
		panic(err)
	}
}

func doMain() error {

	config := config{}
	buf, err := ioutil.ReadFile("cnf.json")
	if err != nil {
		return err
	}
	// parse config file
	err = json.Unmarshal(buf, &config)
	if err != nil {
		return err
	}

	d, err := NewDatabse(config.ConnectOptions)
	if err != nil {
		return err
	}

	db = NewCache(d, 40*time.Second, 1*time.Second)

	http.Handle("/new", http.HandlerFunc(New))
	http.Handle("/bid", http.HandlerFunc(Bid))
	http.Handle("/result", http.HandlerFunc(Result))

	fmt.Printf("Stone Scissors Paper game service v.%s\n", version)
	fmt.Printf("Starting service at %s\n", config.HostPort)
	return http.ListenAndServe(config.HostPort, nil)
}

func New(w http.ResponseWriter, req *http.Request) {
	round := NewRound()
	if err := db.Store(round); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, _ := json.Marshal(struct {
		Round string `json:"round"`
		User1 string `json:"user1"`
		User2 string `json:"user2"`
	}{
		Round: round.ID,
		User1: round.Player1,
		User2: round.Player2,
	})
	w.Header().Add("Content-Type", "application/json")
	w.Write(res)
}

func Bid(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bid   string `json:"bid"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.Bid == "" || input.User == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bid := bidEncode(input.Bid)
	if bid < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrive(input.Round)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := round.Step(bid, input.User)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = db.Store(round)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, _ := json.Marshal(struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

func Result(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	input := struct {
		Round string `json:"round"`
		User  string `json:"user"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.User == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrive(input.Round)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := round.Result(input.User)

	response, _ := json.Marshal(struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

func bidEncode(bid string) int {
	switch strings.ToLower(bid) {
	case "paper":
		return paper
	case "scissors":
		return scissors
	case "stone":
		return stone
	default:
		return -1
	}
}
