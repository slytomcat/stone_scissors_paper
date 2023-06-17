package main

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test0_Hash(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	shData := tr.saltedHash("my secret", []byte("stone"))       // "my secret", []byte("paper")
	shDataHash := "L64zOtDB4yPHkd9ieLH8ghGdzDVn-_2X17Oo2bjDE64" // "ukpgoQizajh7pHqNQM4lWCLxnbwtScUQLHhiQzT5u5Y"
	require.Equal(t, shDataHash, shData)

	shDataHash = tr.roundSaltedHash(player1)
	require.Equal(t, shDataHash, tr.Player1)

	shDataHash = tr.roundSaltedHash(player2)
	require.Equal(t, shDataHash, tr.Player2)

	tr.ID = "e500c6d1-93b5-4bd9-8ceb-a4a87fe60cd5" // fixed ID for predictable result of roundSaltedHash

	shData = tr.roundSaltedHash(player1)
	shDataHash = "J24xKbiLVHxdyh4ArORMaVnOKr5VzAtCOuWmPasPeCM"
	require.Equal(t, shDataHash, shData)
}

func Test1_NewRound(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)
	require.NotEqual(t, tr.Player1, tr.Player2)
	p1 := tr.roundSaltedHash(player1)
	p2 := tr.roundSaltedHash(player2)
	storedSig := tr.Signature
	storedID := tr.ID
	tr.Signature = ""
	require.Equal(t, tr.roundSaltedHash(tr), storedSig)
	tr.ID = ""
	require.Equal(t, &Round{
		Player1: p1,
		Player2: p2,
	}, tr)
	tr.Signature = storedSig
	tr.ID = storedID

	res := tr.Result(player1)
	require.Equal(t, "place Your bet, please", res)

	bet1 := tr.saltedHash("my secret", []byte("paper"))
	res = tr.Bet(bet1, player1)
	require.Equal(t, "wait for the rival to place its bet", res)

	require.Equal(t, bet1, tr.HiddenBet1)
	require.Equal(t, "", tr.HiddenBet2)

	res = tr.Disclose("my secret", "paper", player1)
	require.Equal(t, "wait for the rival to place its bet", res)

	res = tr.Bet(tr.saltedHash("my secret", []byte("stone")), player1)
	require.Equal(t, "bet has already been placed", res)

	res = tr.Result(player1)
	require.Equal(t, "wait for the rival to place its bet", res)

	res = tr.Result(player2)
	require.Equal(t, "place Your bet, please", res)

	res = tr.Bet(tr.saltedHash("my 2 secret", []byte("stone")), player2)
	require.Equal(t, "disclose your bet, please", res)

	res = tr.Result(player1)
	require.Equal(t, "disclose your bet, please", res)

	res = tr.Disclose("wrong secret", "paper", player1)
	require.Equal(t, "Your bet is incorrect", res)

	res = tr.Disclose("my secret", "stone", player1)
	require.Equal(t, "Your bet is incorrect", res)

	res = tr.Disclose("my secret", "paper", player1)
	require.Equal(t, "wait for your rival to disclose its bet", res)

	res = tr.Result(player1)
	require.Equal(t, "wait for your rival to disclose its bet", res)

	res = tr.Result(player2)
	require.Equal(t, "disclose your bet, please", res)

	res = tr.Disclose("my 2 secret", "stone", player2)
	require.Equal(t, "You lose: your bet: stone, the rival's bet: paper", res)

	res = tr.Result(player1)
	require.Equal(t, "You won: your bet: paper, the rival's bet: stone", res)

	res = tr.Result(player2)
	require.Equal(t, "You lose: your bet: stone, the rival's bet: paper", res)
}

func Test2_NewRound(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)
	res := tr.Result(player1)
	require.Equal(t, "place Your bet, please", res)

	res = tr.Bet(tr.saltedHash("my secret", []byte("paper")), player1)
	require.Equal(t, "wait for the rival to place its bet", res)

	res = tr.Bet(tr.saltedHash("my 2 secret", []byte("paper")), player2)
	require.Equal(t, "disclose your bet, please", res)

	res = tr.Disclose("my secret", "paper", player1)
	require.Equal(t, "wait for your rival to disclose its bet", res)

	res = tr.Disclose("my 2 secret", "paper", player2)
	require.Equal(t, "draw: your bet: paper, the rival's bet: paper", res)

	res = tr.Result(player1)
	require.Equal(t, "draw: your bet: paper, the rival's bet: paper", res)

	res = tr.Result(player2)
	require.Equal(t, "draw: your bet: paper, the rival's bet: paper", res)
}

func Test5_NewRound_async(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	var wg sync.WaitGroup
	wg.Add(8)
	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("scissors")), player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("paper")), player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("stone")), player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("scissors")), player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("paper")), player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Bet(tr.saltedHash("my secret", []byte("stone")), player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	wg.Wait()

	wg.Add(6)
	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "scissors", player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "paper", player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "stone", player1)
		t.Logf("received S1 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "scissors", player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "stone", player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Disclose("my secret", "paper", player2)
		t.Logf("received S2 result: %s", res)
	}(tr)

	wg.Wait()

	res := tr.Result(player1)
	t.Logf("received S1 result: %s", res)
	res = tr.Result(player2)
	t.Logf("received S2 result: %s", res)
}

func Test6_falsificate(t *testing.T) {
	tr := NewRound("u1", "u2")

	tr.Bet1 = scissors

	res := tr.Result("u2")
	require.Equal(t, "round had been falsificated", res)
}

func Test7_bidEncodeDecode(t *testing.T) {
	tr := NewRound("u1", "u2")

	if tr.betDecode(stone) != "stone" ||
		tr.betDecode(scissors) != "scissors" ||
		tr.betDecode(paper) != "paper" ||
		tr.betDecode(nothing) != "" {
		t.Error("wrong bids decoding")
	}

	if tr.betEncode("Stone") != stone ||
		tr.betEncode("scIssors") != scissors ||
		tr.betEncode("papEr") != paper ||
		tr.betEncode("nOthing") != -1 {
		t.Error("wrong bids encoding")
	}

}

func Test8_NewRoundUnauthorized(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)
	res := tr.Bet(tr.saltedHash("my secret", []byte("stone")), "player3")
	require.Equal(t, "unauthorized", res)

	res = tr.Disclose("my secret", "stone", "player3")
	require.Equal(t, "unauthorized", res)

	res = tr.Result("player3")
	require.Equal(t, "unauthorized", res)
}

func Test9_authorized(t *testing.T) {
	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	err := tr.authorized(player1)
	if err != nil {
		t.Error("first player unauthorized")
	}
	err = tr.authorized(player2)
	if err != nil {
		t.Error("second player unauthorized")
	}
	err = tr.authorized("player3")
	if err == nil {
		t.Error("unauthorized user is authorized")
	}
}
