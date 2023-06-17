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

	"github.com/stretchr/testify/require"
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
	require.Eventually(t, func() bool {
		// use wrong path to check the mux
		resp, err := http.Post("http://localhost:8080/wrong_new", "application/json", nil)
		if err == nil {
			resp.Body.Close()
			return true
		}
		return false

	}, time.Millisecond*500, time.Millisecond*50)

	// wg.Wait()
	return &wg
}

func stopService(wg *sync.WaitGroup) {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	wg.Wait()
}

func envSet(t testing.TB) {
	if os.Getenv("CI") != "" {
		return
	}
	file, err := os.ReadFile(".env")
	require.NoError(t, err)
	for _, l := range strings.Split(string(file), "\n") {
		env := strings.Split(l, "=")
		if len(env) == 2 {
			t.Setenv(strings.Trim(env[0], " \t"), strings.Trim(env[1], " \t"))
		}
	}
}

func Test_serviceMissingENV(t *testing.T) {
	timer := time.AfterFunc(time.Second, func() { syscall.Kill(syscall.Getpid(), syscall.SIGINT) })
	defer timer.Stop()

	err := doMain(&config{RedisAddrs: []string{""}})
	require.Error(t, err)
}

func Test_gracefulShutdown(t *testing.T) {
	envSet(t)
	wg := startService(t)
	// graceful shutdown
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	stopService(wg)

	w.Close()
	log.SetOutput(os.Stdout)

	buf, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(buf), "Shutdown finished.")
	require.Contains(t, string(buf), "http: Server closed")
}

func request(path string, req []byte) ([]byte, error) {
	resp, err := http.Post("http://localhost:8080/"+path, "application/json", bytes.NewReader(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
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

	data, err := request("new", req)
	res := struct {
		Round string `json:"round"`
	}{}

	err = json.Unmarshal(data, &res)
	require.NoError(t, err)

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

	data, err = request("bet", req)

	require.Equal(t, `{"response":"wait for the rival to place its bet"}`, string(data))

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

	data, err = request("bet", req)

	require.Equal(t, `{"response":"disclose your bet, please"}`, string(data))

	// result
	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{
		Round:  res.Round,
		Player: player1,
	})

	data, err = request("result", req)

	require.Equal(t, `{"response":"disclose your bet, please"}`, string(data))

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

	data, err = request("disclose", req)

	require.Equal(t, `{"response":"wait for your rival to disclose its bet"}`, string(data))

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

	data, err = request("disclose", req)

	require.Equal(t, `{"response":"You lose: your bet: stone, the rival's bet: paper"}`, string(data))

	// result
	req, _ = json.Marshal(struct {
		Round  string `json:"round"`
		Player string `json:"player"`
	}{
		Round:  res.Round,
		Player: player1,
	})

	data, err = request("result", req)

	require.Equal(t, `{"response":"You won: your bet: paper, the rival's bet: stone"}`, string(data))
}

func Test_BadRequests(t *testing.T) {
	envSet(t) // load .env file for test environment
	defer stopService(startService(t))

	badRequest("http://localhost:8080/new", t)

	badRequest("http://localhost:8080/bet", t)

	badRequest("http://localhost:8080/disclose", t)

	badRequest("http://localhost:8080/result", t)
}

func badRequest(url string, t *testing.T) {
	resp, err := http.Post(url, "application/json", strings.NewReader(`{}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	resp, err = http.Post(url, "application/json", strings.NewReader(`{~`))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func Test_serviceBadRound(t *testing.T) {
	envSet(t) // load .env file for test environment
	defer stopService(startService(t))

	badRound("http://localhost:8080/bet", `{"player":"p1","bet":"jasj","round":"not_existing"}`, t)

	badRound("http://localhost:8080/disclose", `{"player":"p1","bet":"jasj","secret":"s","round":"not_existing"}`, t)

	badRound("http://localhost:8080/result", `{"player":"p1","round":"not_existing"}`, t)
}

func badRound(url, params string, t *testing.T) {
	resp, err := http.Post(url, "application/json", strings.NewReader(params))
	require.NoError(t, err)
	defer resp.Body.Close()
}
