package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
)

type configT struct {
	HostPort      string `default:"localhost:8080"`
	RedisAddrs    []string
	RedisPassword string
	ServerSalt    string
}

var (
	db      Database
	version string
	config  = configT{}
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}

func doMain() error {

	// config := config{}
	err := envconfig.Process("SSP", &config)
	if err != nil {
		return err
	}

	d, err := NewDatabse(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
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
	http.Handle("/disclose", http.HandlerFunc(Disclose))
	http.Handle("/result", http.HandlerFunc(Result))

	log.Printf("Stone Scissors Paper game service v.%s\n", version)
	log.Printf("Starting service at %s\n", config.HostPort)

	go server.ListenAndServe()

	sig := make(chan (os.Signal))
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP)

	<-sig

	log.Println("\nInterrupted. Starting shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		return err
	}

	log.Println("Shutdown finished.")

	return nil
}

// New realizes the request for new round
func New(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Body reading error: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := struct {
		Player1 string `json:"player1"`
		Player2 string `json:"player2"`
	}{}

	err = json.Unmarshal(buf, &input)
	if err != nil || input.Player1 == "" || input.Player2 == "" {
		log.Printf("Wrong input params: '%s'", buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round := NewRound(input.Player1, input.Player2)

	err = db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, _ := json.Marshal(struct {
		Round string `json:"round"`
	}{
		Round: round.ID,
	})
	log.Printf("new round: %s", res)
	w.Header().Add("Content-Type", "application/json")
	w.Write(res)
}

// Bet realizes the request for the new bet of user
func Bet(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Body reading error: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Bet    string `json:"bet"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.Bet == "" || input.Player == "" {
		log.Printf("Wrong input params: '%s'", buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := round.Step(input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, _ := json.Marshal(struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})

	log.Printf("round: %s - bet result: %s", round.ID, res)
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// Disclose realizes the request for the disclose bet of user
func Disclose(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Body reading error: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Secret string `json:"secret"`
		Bet    string `json:"bet"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.Bet == "" || input.Secret == "" || input.Player == "" {
		log.Printf("Wrong input params: '%s'", buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := round.Disclose(input.Secret, input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, _ := json.Marshal(struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})

	log.Printf("round: %s - disclose result: %s", round.ID, res)
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// Result realizes the request for result of round
func Result(w http.ResponseWriter, req *http.Request) {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("Body reading error: %+v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{}
	err = json.Unmarshal(buf, &input)
	if err != nil || input.Round == "" || input.Player == "" {
		log.Printf("Wrong input params: '%s'", buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res := round.Result(input.Player)

	response, _ := json.Marshal(struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})

	log.Printf("round: %s - result: %s", round.ID, res)
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}
