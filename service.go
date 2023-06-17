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
)

var (
	db         Database
	version    = "test_version"
	serverSalt = ""
)

func main() {
	cfg, err := newConfig()
	if err != nil {
		log.Fatal(err)
	}
	err = doMain(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func doMain(cfg *config) error {
	serverSalt = cfg.ServerSalt

	d, err := NewDatabase(redis.UniversalOptions{Addrs: cfg.RedisAddrs, Password: cfg.RedisPassword})
	if err != nil {
		return err
	}

	db = NewCache(d, 40*time.Second, 1*time.Second)

	mux := http.NewServeMux()
	mux.HandleFunc("/new", New)
	mux.HandleFunc("/attach", Attach)
	mux.HandleFunc("/bet", Bet)
	mux.HandleFunc("/disclose", Disclose)
	mux.HandleFunc("/result", Result)

	server := http.Server{
		Addr:    cfg.HostPort,
		Handler: mux,
	}

	log.Printf("Stone Scissors Paper game service v.%s\n", version)
	log.Printf("Starting service at %s\n", cfg.HostPort)

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

// getInput reads request body and parse it as JSON in to input sruct
func getInput(req *http.Request, input interface{}) error {
	defer req.Body.Close()

	if req.Method != "POST" {
		return fmt.Errorf("wrong method: %s", req.Method)
	}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return fmt.Errorf("request body reading error: %v", err)
	}

	err = json.Unmarshal(buf, input)
	if err != nil {
		return fmt.Errorf("request body parsing error: %v", err)
	}

	return nil
}

// sendResponse writes response struct as JSON into response body
func sendResponse(w http.ResponseWriter, response interface{}) {
	resp, _ := json.Marshal(response)
	w.Header().Add("Content-Type", "application/json")
	_, err := w.Write(resp)
	if err != nil {
		log.Printf("response writing error: %v", err)
	}
}

// New realizes the request for new round
func New(w http.ResponseWriter, req *http.Request) {

	input := struct {
		Player string `json:"player"`
	}{}
	if err := getInput(req, &input); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.Player == "" {
		errMsg := fmt.Sprintf("Some mandatory fields are missed: %+v", input)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	round := NewRound(input.Player)

	err := db.Store(round)
	if err != nil {
		storageError("Round store error", err, w)
		return
	}

	sendResponse(w, struct {
		Round string `json:"round"`
	}{
		Round: round.ID,
	})

	log.Printf("new round: %s started by %s", round.ID, input.Player)
}

// Attach realizes the request for attach to existing round
func Attach(w http.ResponseWriter, req *http.Request) {
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
		storageError("Round retrieve error", err, w)
		return
	}

	res := round.Attach(input.Player)

	err = db.Store(round)
	if err != nil {
		storageError("Round store error", err, w)
		return
	}

	sendResponse(w, struct {
		Response string `json:"response"`
	}{
		Response: res,
	})
	log.Printf("round: %s: %s attached", round.ID, input.Player)

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
		storageError("Round retrieve error", err, w)
		return
	}

	res := round.Bet(input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		storageError("Round store error", err, w)
		return
	}

	sendResponse(w, struct {
		Response string `json:"response"`
	}{
		Response: res,
	})
	log.Printf("round: %s:%s - bet result: %s", round.ID, input.Player, res)
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
		storageError("Round retrieve error", err, w)
		return
	}

	res := round.Disclose(input.Secret, input.Bet, input.Player)

	err = db.Store(round)
	if err != nil {
		storageError("Round store error", err, w)
		return
	}

	sendResponse(w, struct {
		Response string `json:"response"`
	}{
		Response: res,
	})
	log.Printf("round: %s:%s - disclose result: %s", round.ID, input.Player, res)
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
		storageError("Round retrieve error", err, w)
		return
	}

	res := round.Result(input.Player)

	sendResponse(w, struct {
		Response string `json:"response"`
	}{
		Response: res,
	})
	log.Printf("round: %s:%s - result: %s", round.ID, input.Player, res)
}

func storageError(msg string, err error, w http.ResponseWriter) {
	log.Printf("%s: %v", msg, err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
