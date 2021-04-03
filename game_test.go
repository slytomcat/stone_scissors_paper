package main

import (
	"sync"
	"testing"

	uuid "github.com/satori/go.uuid"
)

var test_round *Round

func Test1_NewRound(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	res := test_round.Result(test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(paper, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(stone, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(stone, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Step(paper, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Result(test_round.Player1)

	t.Logf("received result:%s", res)

}

func Test2_NewRound(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	res := test_round.Step(-1, test_round.Player1)

	t.Logf("received round:%v+", res)

	res = test_round.Step(scissors, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(stone, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(stone, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Step(paper, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Result(test_round.Player1)

	t.Logf("received result:%s", res)

}

func Test3_NewRound(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	res := test_round.Step(scissors, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(stone, test_round.Player1)

	t.Logf("received result:%s", res)

	res = test_round.Step(scissors, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Step(paper, test_round.Player2)

	t.Logf("received result:%s", res)

	res = test_round.Result(test_round.Player1)

	t.Logf("received result:%s", res)
}

func Test5_NewRound_async(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	var wg sync.WaitGroup
	wg.Add(8)
	go func(r *Round) {
		defer wg.Done()
		res := r.Step(scissors, test_round.Player1)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Step(paper, test_round.Player1)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Step(stone, test_round.Player1)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Step(scissors, test_round.Player2)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Step(stone, test_round.Player2)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Step(paper, test_round.Player2)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(test_round.Player1)
		t.Logf("received result:%s", res)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res := r.Result(test_round.Player2)
		t.Logf("received result:%s", res)
	}(test_round)

	wg.Wait()

	t.Log(test_round)
}

func Test8_NewRoundUnauthorized(t *testing.T) {

	test_round = NewRound()

	t.Logf("received round:%v+", test_round)

	res := test_round.Step(scissors, uuid.NewV4().String())

	t.Logf("received result:%s", res)

	res = test_round.Result(uuid.NewV4().String())

	t.Logf("received result:%s", res)
}

func Test9_authorized(t *testing.T) {
	err := test_round.authorized(test_round.Player1)
	if err != nil {
		t.Error("first user unauthorized")
	}
	err = test_round.authorized(test_round.Player2)
	if err != nil {
		t.Error("first user unauthorized")
	}
	err = test_round.authorized(uuid.NewV4().String())
	if err == nil {
		t.Error("unauthorized user is authorized")
	}
}
