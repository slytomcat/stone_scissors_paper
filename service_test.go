package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func Test_success(t *testing.T) {
	go t.Log(doMain())

	resp, err := http.Get("localhost:8080/new")
	if err != nil {
		t.Error(err)
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

	req, _ := json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bid   string `json:"bid"`
	}{
		Round: res.Round,
		User:  res.User1,
		Bid:   "paper",
	})

	resp, err = http.Post("localhost:8080/bid", "application/json", bytes.NewReader(req))
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

	req, _ = json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
		Bid   string `json:"bid"`
	}{
		Round: res.Round,
		User:  res.User2,
		Bid:   "stone",
	})

	resp, err = http.Post("localhost:8080/bid", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"you lose"}` {
		t.Error("Unexpected response")
	}

	req, _ = json.Marshal(struct {
		Round string `json:"round"`
		User  string `json:"user"`
	}{
		Round: res.Round,
		User:  res.User1,
	})

	resp, err = http.Post("localhost:8080/result", "application/json", bytes.NewReader(req))
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	data, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}

	if string(data) != `{"respose":"you won"}` {
		t.Error("Unexpected response")
	}

}
