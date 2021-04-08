package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func saltedHash(salt, obj string) string {

	h := sha256.Sum256(append([]byte(obj), []byte(salt)...))

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func Test_success(t *testing.T) {
	godotenv.Load() // load .env file for test environment

	go doMain()

	time.Sleep(time.Millisecond * 500)

	player1 := "player1"
	player2 := "player2"

	req, _ := json.Marshal(struct {
		Player1 string `json:"player1"`
		Player2 string `json:"player2"`
	}{
		Player1: player1,
		Player2: player2,
	})

	resp, err := http.Post("http://localhost:8080/new", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	t.Logf("responce: %s", data)
	res := struct {
		Round string `json:"round"`
	}{}

	err = json.Unmarshal(data, &res)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Received new round: %v", res)

	// place bets

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Bet    string `json:"bet"`
	}{
		Round:  res.Round,
		Player: player1,
		Bet:    saltedHash("p1 secret", "paper"),
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
	t.Logf("responce: %s", data)

	if string(data) != `{"respose":"wait for the rival to place its bet"}` {
		t.Error("Unexpected response")
	}

	t.Logf("Received step1: %s", data)

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Bet    string `json:"bet"`
	}{
		Round:  res.Round,
		Player: player2,
		Bet:    saltedHash("p2 secret", "stone"),
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

	if string(data) != `{"respose":"disclose your bet, please"}` {
		t.Errorf("Unexpected response: %s", data)
	}

	t.Logf("Received step2: %s", data)

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{
		Round:  res.Round,
		Player: player1,
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
	t.Logf("Received result: %s", data)

	if string(data) != `{"respose":"disclose your bet, please"}` {
		t.Errorf("Unexpected response: %s", data)
	}

	t.Logf("Received result: %s", data)

	// Disclose

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Secret string `json:"secret"`
		Bet    string `json:"bet"`
	}{
		Round:  res.Round,
		Player: player1,
		Secret: "p1 secret",
		Bet:    "paper",
	})

	resp, err = http.Post("http://localhost:8080/disclose", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	t.Logf("responce: %s", data)

	if string(data) != `{"respose":"wait for your rival to disclose its bet"}` {
		t.Error("Unexpected response")
	}

	t.Logf("Received disclose1: %s", data)

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
		Secret string `json:"secret"`
		Bet    string `json:"bet"`
	}{
		Round:  res.Round,
		Player: player2,
		Secret: "p2 secret",
		Bet:    "stone",
	})

	resp, err = http.Post("http://localhost:8080/disclose", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"You lose: your bet: stone, the rival's bet: paper"}` {
		t.Errorf("Unexpected response: %s", data)
	}

	t.Logf("Received disclose2: %s", data)

	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{
		Round:  res.Round,
		Player: player1,
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
	t.Logf("Received result: %s", data)

	if string(data) != `{"respose":"You won: your bet: paper, the rival's bet: stone"}` {
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
