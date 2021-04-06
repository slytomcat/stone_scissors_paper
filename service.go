package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
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

	server := http.Server{
		Addr: config.HostPort,
	}

	server.Handler = http.DefaultServeMux
	http.Handle("/new", http.HandlerFunc(New))
	http.Handle("/bet", http.HandlerFunc(Bet))
	http.Handle("/result", http.HandlerFunc(Result))

	fmt.Printf("Stone Scissors Paper game service v.%s\n", version)
	fmt.Printf("Starting service at %s\n", config.HostPort)

	go server.ListenAndServe()

	sig := make(chan (os.Signal))
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP)

	<-sig

	fmt.Println("\nIterupted. Starting shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Shutdown finished.")

	return nil
}

// New realizes the request for new round
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

// Bet realizes the request for the new bet of user
func Bet(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bet   string `json:"bet"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.Bet == "" || input.User == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bet := round.betEncode(input.Bet)
	if bet < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	res := round.Step(bet, input.User)
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

// Result realizes the request for result of round
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

	round, err := db.Retrieve(input.Round)
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
