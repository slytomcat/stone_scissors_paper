package main

import (
	"sync"
	"testing"
)

func Test0_Hash(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	shData := tr.saltedHash("my secret", []byte("paper"))
	shDataHash := "ukpgoQizajh7pHqNQM4lWCLxnbwtScUQLHhiQzT5u5"
	if shData != shDataHash {
		t.Errorf("unexpected result from saltedHash: '%s'  while expecting: '%s'", shData, shDataHash)
	}

	shDataHash = tr.roundSaltedHash(player1)
	if tr.Player1 != shDataHash {
		t.Errorf("unexpected Player1: '%s'  while expecting: '%s'", tr.Player1, shDataHash)
	}

	shDataHash = tr.roundSaltedHash(player2)
	if tr.Player2 != shDataHash {
		t.Errorf("unexpected Player2: '%s'  while expecting: '%s'", tr.Player1, shDataHash)
	}

	tr.ID = "e500c6d1-93b5-4bd9-8ceb-a4a87fe60cd5" // fixed ID for predictable result of roundSaltedHash

	shData = tr.roundSaltedHash(player1)
	shDataHash = "J24xKbiLVHxdyh4ArORMaVnOKr5VzAtCOuWmPasPeC"
	if shData != shDataHash {
		t.Errorf("unexpected result from saltedHash: '%s'  while expecting: '%s'", shData, shDataHash)
	}

}

func Test1_NewRound(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	if tr.Player1 == tr.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v", tr)

	res := tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Step(tr.saltedHash("my secret", []byte("paper")), player1)

	t.Logf("received result after fist player step:%s", res)

	t.Logf("updated round:%v", tr)

	res = tr.Disclose("my secret", "paper", player1)

	t.Logf("received result after discosure of first player:%s", res)

	res = tr.Step(tr.saltedHash("my secret", []byte("stone")), player1)

	t.Logf("received result after fist player double step:%s", res)

	res = tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Result(player2)

	t.Logf("received result for second player :%s", res)

	res = tr.Step(tr.saltedHash("my 2 secret", []byte("stone")), player2)

	t.Logf("received result after second player step:%s", res)

	res = tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Disclose("wrong secret", "paper", player1)

	t.Logf("received result after wrong discosure of first player:%s", res)

	res = tr.Disclose("my secret", "stone", player1)

	t.Logf("received result after wrong discosure of first player:%s", res)

	res = tr.Disclose("my secret", "paper", player1)

	t.Logf("received result after correct discosure of first player:%s", res)

	res = tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Result(player2)

	t.Logf("received result for second player :%s", res)

	res = tr.Disclose("my 2 secret", "stone", player2)

	t.Logf("received result after correct discosure of second player:%s", res)

	res = tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Result(player2)

	t.Logf("received result for second player :%s", res)

}

func Test2_NewRound(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	t.Logf("received round:%v+", tr)

	res := tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Step(tr.saltedHash("my secret", []byte("paper")), player1)

	t.Logf("received result after fist player step:%s", res)

	res = tr.Step(tr.saltedHash("my 2 secret", []byte("paper")), player2)

	t.Logf("received result after second player step:%s", res)

	res = tr.Disclose("my secret", "paper", player1)

	t.Logf("received result after correct discosure of first player:%s", res)

	res = tr.Disclose("my 2 secret", "paper", player2)

	t.Logf("received result after correct discosure of second player:%s", res)

	res = tr.Result(player1)

	t.Logf("received result for first player :%s", res)

	res = tr.Result(player2)

	t.Logf("received result for second player :%s", res)

}

func Test5_NewRound_async(t *testing.T) {

	player1 := "player1"
	player2 := "player2"

	tr := NewRound(player1, player2)

	if tr.Player1 == tr.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", tr)

	//test_round.mx.Lock()

	var wg sync.WaitGroup
	wg.Add(8)
	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("scissors")), player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("paper")), player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("stone")), player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("scissors")), player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("paper")), player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Step(tr.saltedHash("my secret", []byte("stone")), player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player1)
		t.Logf("received R result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player2)
		t.Logf("received R result:%s", res)
	}(tr)

	//test_round.mx.Unlock()

	wg.Wait()

	t.Log(tr)

	wg.Add(8)
	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "scissors", player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "paper", player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "stone", player1)
		t.Logf("received S1 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "scissors", player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "stone", player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := tr.Disclose("my secret", "paper", player2)
		t.Logf("received S2 result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player1)
		t.Logf("received R result:%s", res)
	}(tr)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(player2)
		t.Logf("received R result:%s", res)
	}(tr)

	//test_round.mx.Unlock()

	wg.Wait()

	t.Log(tr)
}

func Test6_bidEncodeDecode(t *testing.T) {
	tr := NewRound("u1", "u2")

	tr.Bet1 = scissors

	res := tr.Result("u2")
	t.Logf("received R result:%s", res)

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

	t.Logf("received round:%v+", tr)

	res := tr.Step(tr.saltedHash("my secret", []byte("stone")), "player3")

	t.Logf("received result for unauthorized user step:%s", res)

	res = tr.Disclose("my secret", "stone", "player3")

	t.Logf("received result for unauthorized user disclose:%s", res)

	res = tr.Result("player3")

	t.Logf("received result for unauthorized user:%s", res)
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
