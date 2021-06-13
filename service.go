package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	HostPort      string   `default:"localhost:8080"`
	RedisAddrs    []string `required:"true"`
	ServerSalt    string   `required:"true"`
	RedisPassword string
}

var (
	db      Database
	version = "test"
	config  = configT{}
)

func main() {
	err := doMain()
	if err != nil {
		log.Fatal(err)
	}
}

func doMain() error {

	err := envconfig.Process("SSP", &config)
	if err != nil {
		return err
	}

	d, err := NewDatabase(redis.UniversalOptions{Addrs: config.RedisAddrs, Password: config.RedisPassword})
	if err != nil {
		return err
	}

	db = NewCache(d, 40*time.Second, 1*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", New)
	mux.HandleFunc("/bet", Bet)
	mux.HandleFunc("/disclose", Disclose)
	mux.HandleFunc("/result", Result)

	server := http.Server{
		Addr:    config.HostPort,
		Handler: mux,
	}

	log.Printf("Stone Scissors Paper game service v.%s\n", version)
	log.Printf("Starting service at %s\n", config.HostPort)

	go func() { log.Println(server.ListenAndServe()) }()

	sig := make(chan (os.Signal), 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	s := <-sig // wait for a signal

	log.Printf("%s received. Starting shutdown...", s.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Printf("Shutdown error:%v", err)
		return err
	}

	log.Println("Shutdown finished.")
	return nil
}

// getInput reds request body and parse it as JSON in to input sruct
func getInput(req *http.Request, input interface{}) error {
	defer req.Body.Close()

	if req.Method != "POST" {
		return fmt.Errorf("wrong method: %s", req.Method)
	}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("request body reading error: %W", err)
	}

	err = json.Unmarshal(buf, input)
	if err != nil {
		return fmt.Errorf("request body parsing error: %W", err)
	}

	return nil
}

// sendResponse writes response struct as JSON into response body
func sendResponse(w http.ResponseWriter, response interface{}) {
	resp, _ := json.Marshal(response)
	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write(resp)
	if err != nil {
		log.Printf("respone writing error: %v", err)
	}
}

// New realizes the request for new round
func New(w http.ResponseWriter, req *http.Request) {

	input := struct {
		Player1 string `json:"player1"`
		Player2 string `json:"player2"`
	}{}
	if err := getInput(req, &input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Player1 == "" || input.Player2 == "" {
		errMsg := fmt.Sprintf("Some mandatory fields are missed: %+v", input)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	round := NewRound(input.Player1, input.Player2)

	err := db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, struct {
		Round string `json:"round"`
	}{
		Round: round.ID,
	})

	log.Printf("new round: %s", round.ID)
}

// Bet realizes the request for the new bet of user
func Bet(w http.ResponseWriter, req *http.Request) {

	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Bet    string `json:"bet"`
	}{}
	if err := getInput(req, &input); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Round == "" || input.Player == "" || input.Bet == "" {
		errMsg := fmt.Sprintf("Some mandatory fields are missed: %+v", input)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := round.Step(input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})
	log.Printf("round: %s - bet result: %s", round.ID, res)
}

// Disclose realizes the request for the disclose bet of user
func Disclose(w http.ResponseWriter, req *http.Request) {

	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Secret string `json:"secret"`
		Bet    string `json:"bet"`
	}{}
	if err := getInput(req, &input); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Round == "" || input.Player == "" || input.Bet == "" || input.Secret == "" {
		errMsg := fmt.Sprintf("Some mandatory fields are missed: %+v", input)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := round.Disclose(input.Secret, input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		log.Printf("Round store error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})
	log.Printf("round: %s - disclose result: %s", round.ID, res)
}

// Result realizes the request for result of round
func Result(w http.ResponseWriter, req *http.Request) {

	input := struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{}

	if err := getInput(req, &input); err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Round == "" || input.Player == "" {
		errMsg := fmt.Sprintf("Some mandatory fields are missed: %+v", input)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	round, err := db.Retrieve(input.Round)
	if err != nil {
		log.Printf("Round retrieve error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := round.Result(input.Player)

	sendResponse(w, struct {
		Respose string `json:"respose"`
	}{
		Respose: res,
	})
	log.Printf("round: %s - result: %s", round.ID, res)
}
