package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
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
	"github.com/stretchr/testify/assert"
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
	assert.Eventually(t, func() bool {
		// use wrong method to check the mux
		resp, err := http.Get("http://localhost:8080/new")
		if err == nil {
			resp.Body.Close()
			return true
		}
		return false

	}, time.Millisecond*500, time.Millisecond*50)

	return &wg
}

func stopService(wg *sync.WaitGroup) {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	wg.Wait()
}

func envSet(t testing.TB, files ...string) {
	envs, _ := godotenv.Read(files...)
	for k, v := range envs {
		t.Setenv(k, v)
	}
}

func Test_serviceMissingENV(t *testing.T) {

	envSet(t)
	os.Unsetenv("SSP_REDISADDRS")

	timer := time.NewTimer(time.Second)
	go func(t *time.Timer) {
		<-t.C
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}(timer)

	assert.Error(t, doMain())
	timer.Stop()
}

func Test_serviceWrongEnv(t *testing.T) {

	envSet(t)
	os.Setenv("SSP_REDISADDRS", "wrong.addrs:5555")

	timer := time.NewTimer(time.Second)
	go func(t *time.Timer) {
		<-t.C
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}(timer)

	assert.Error(t, doMain())
	timer.Stop()
}

func Test_gracefulSutdown(t *testing.T) {
	envSet(t)
	wg := startService(t)
	// graceful sutdown
	t.Log("Testing graceful sutdown")
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	stopService(wg)

	w.Close()
	log.SetOutput(os.Stdout)

	buf, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Contains(t, string(buf), "Shutdown finished.")
	assert.Contains(t, string(buf), "http: Server closed")
}

func Test_serviceFullGame(t *testing.T) {
	envSet(t) // load .env file for test environment
	defer stopService(startService(t))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("responce: %s", data)
	res := struct {
		Round string `json:"round"`
	}{}

	err = json.Unmarshal(data, &res)
	assert.NoError(t, err)

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("responce: %s", data)

	assert.Equal(t, `{"respose":"wait for the rival to place its bet"}`, string(data))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, `{"respose":"disclose your bet, please"}`, string(data))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("Received result: %s", data)

	assert.Equal(t, `{"respose":"disclose your bet, please"}`, string(data))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("responce: %s", data)

	assert.Equal(t, `{"respose":"wait for your rival to disclose its bet"}`, string(data))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, `{"respose":"You lose: your bet: stone, the rival's bet: paper"}`, string(data))

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
	assert.NoError(t, err)
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	t.Logf("Received result: %s", data)

	assert.Equal(t, `{"respose":"You won: your bet: paper, the rival's bet: stone"}`, string(data))

	t.Logf("Received result: %s", data)

}

func Test_BadRequests(t *testing.T) {
	envSet(t) // load .env file for test environment
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
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(url, "application/json", strings.NewReader(`{~`))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func Test_serviceBadRound(t *testing.T) {
	envSet(t) // load .env file for test environment

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
	assert.NoError(t, err)
	defer resp.Body.Close()

}
