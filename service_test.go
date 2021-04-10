package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func saltedHash(salt, obj string) string {

	h := sha256.Sum256(append([]byte(obj), []byte(salt)...))

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func startService(t *testing.T) *sync.WaitGroup {

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		main()
		wg.Done()
	}()

	// wait for service start
	err := errors.New("")
	var resp *http.Response
	timeout := time.After(time.Millisecond * 500)
	for err != nil {
		select {
		case <-timeout:
			t.Fatal("Service has failed to start")
		default:
			// use wrong path request to check the mux
			resp, err = http.Post("http://localhost:8080", "application/json", strings.NewReader(``))
			if err == nil {
				resp.Body.Close()
			}
		}
	}
	return &wg
}

func stopService(wg *sync.WaitGroup) {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	wg.Wait()
}

func Test_serviceWrongENV(t *testing.T) {

	ss := os.Getenv("SSP_SERVERSALT")
	defer func() {
		os.Setenv("SSP_SERVERSALT", ss)
	}()
	os.Unsetenv("SSP_SERVERSALT")
	//godotenv.Load()

	timer := time.NewTimer(time.Second)
	go func(t *time.Timer) {
		<-t.C
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}(timer)

	err := doMain()
	if err == nil {
		t.Error("No error when expected")
	} else {
		timer.Stop()
		t.Logf("Received expected: %v", err)
	}
}

func Test_serviceWrongENV2(t *testing.T) {

	godotenv.Load()
	adrs := os.Getenv("SSP_REDISADDRS")
	defer func() {
		os.Setenv("SSP_REDISADDRS", adrs)
	}()
	os.Setenv("SSP_REDISADDRS", "wrong.addrs:5555")

	timer := time.NewTimer(time.Second)
	go func(t *time.Timer) {
		<-t.C
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}(timer)

	err := doMain()
	if err == nil {
		t.Error("No error when expected")
	} else {
		timer.Stop()
		t.Logf("Received expected: %v", err)
	}
}

func Test_serviceFullGame(t *testing.T) {
	godotenv.Load() // load .env file for test environment

	wg := startService(t)

	player1 := "player1"
	player2 := "player2"

	// New

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
		t.Errorf("Unexpected response: %v", resp)
	}

	t.Logf("Received step1: %s", data)

	// Bet
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
		t.Errorf("Unexpected response: %v", resp)
	}

	t.Logf("Received step2: %s", data)

	// result
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
		t.Errorf("Unexpected response: %v", resp)
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
		t.Errorf("Unexpected response: %v", resp)
	}

	t.Logf("Received disclose1: %s", data)

	// Disclose
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
		t.Errorf("Unexpected response: %v", resp)
	}

	t.Logf("Received disclose2: %s", data)

	// result
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
		t.Errorf("Unexpected response: %v", resp)
	}

	t.Logf("Received result: %s", data)

	// graceful sutdown
	t.Log("Testing graceful sutdown")
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	stopService(wg)

	w.Close()
	log.SetOutput(os.Stdout)

	buf, err := io.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(buf, []byte("Shutdown finished.")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	if !bytes.Contains(buf, []byte("http: Server closed")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	log.Printf("%s", buf)

}

func Test_BadRequests(t *testing.T) {
	godotenv.Load() // load .env file for test environment

	defer stopService(startService(t))

	t.Log("Testing bad requests to /new")
	badRequest("http://localhost:8080/new", t)

	t.Log("Testing bad requests to /bet")
	badRequest("http://localhost:8080/bet", t)

	t.Log("Testing bad requests to /disclose")
	badRequest("http://localhost:8080/disclose", t)

	t.Log("Testing bad requests to /result")
	badRequest("http://localhost:8080/result", t)

}

func badRequest(url string, t *testing.T) {
	resp, err := http.Post(url, "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Bad responce status code: %s", resp.Status)
	} else {
		t.Logf("Received expected responce code: %s", resp.Status)
	}

	resp, err = http.Post(url, "application/json", strings.NewReader(`{~`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Bad responce status code: %s", resp.Status)
	} else {
		t.Logf("Received expected responce code: %s", resp.Status)
	}
}

func Test_serviceBadRound(t *testing.T) {
	godotenv.Load() // load .env file for test environment

	defer stopService(startService(t))

	t.Log("Testing bad round to /new")
	badRound("http://localhost:8080/bet", `{"player":"p1","bet":"jasj","round":"not_existing"}`, t)

	t.Log("Testing bad round to /disclose")
	badRound("http://localhost:8080/disclose", `{"player":"p1","bet":"jasj","secret":"s","round":"not_existing"}`, t)

	t.Log("Testing bad round to /result")
	badRound("http://localhost:8080/result", `{"player":"p1","round":"not_existing"}`, t)

}

func badRound(url, params string, t *testing.T) {
	resp, err := http.Post(url, "application/json", strings.NewReader(params))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

}
