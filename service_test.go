package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func Test_success(t *testing.T) {
	godotenv.Load() // load .env file for test environment

	go doMain()

	time.Sleep(time.Millisecond * 500)

	resp, err := http.Get("http://localhost:8080/new")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	res := struct {
		Round string `json:"round"`
		User1 string `json:"user1"`
		User2 string `json:"user2"`
	}{}

	err = json.Unmarshal(data, &res)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Received new round: %v", res)

	req, _ := json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bet   string `json:"bet"`
	}{
		Round: res.Round,
		User:  res.User1,
		Bet:   "paper",
	})

	resp, err = http.Post("http://localhost:8080/bet", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"wait"}` {
		t.Error("Unexpected response")
	}

	t.Logf("Received step1: %s", data)

	req, _ = json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bet   string `json:"bet"`
	}{
		Round: res.Round,
		User:  res.User2,
		Bet:   "stone",
	})

	resp, err = http.Post("http://localhost:8080/bet", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"You lose: your bet: Stone, the rival's bet: Paper"}` {
		t.Errorf("Unexpected response: %s", data)
	}

	t.Logf("Received step1: %s", data)

	req, _ = json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
	}{
		Round: res.Round,
		User:  res.User1,
	})

	resp, err = http.Post("http://localhost:8080/result", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"You won: your bet: Paper, the rival's bet: Stone"}` {
		t.Errorf("Unexpected response: %s", data)
	}

	t.Logf("Received result: %s", data)

	// logger := os.Stdout
	// r, w, _ := os.Pipe()
	// os.Stdout = w

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(time.Second)

	// w.Close()
	// os.Stdout = logger

	// buf, err := io.ReadAll(r)
	// if err != nil {
	// 	t.Error(err)
	// }
	// if !bytes.Contains(buf, []byte("Shutdown finished.")) {
	// 	t.Errorf("received unexpected output: %s", buf)
	// }
	// log.Printf("%s", buf)

}
