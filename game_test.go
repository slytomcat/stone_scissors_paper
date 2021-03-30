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

	res, err := test_round.Result(test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.Player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.Player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.Player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

}

func Test2_NewRound(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	res, err := test_round.Step(scissors, test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.Player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.Player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.Player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

}

func Test3_NewRound(t *testing.T) {

	test_round = NewRound()

	if test_round.Player1 == test_round.Player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+", test_round)

	res, err := test_round.Step(scissors, test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.Player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(scissors, test_round.Player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.Player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.Player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)
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
		res, err := r.Step(scissors, test_round.Player1)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Step(paper, test_round.Player1)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Step(stone, test_round.Player1)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Step(scissors, test_round.Player2)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Step(stone, test_round.Player2)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Step(paper, test_round.Player2)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Result(test_round.Player1)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	go func(r *Round) {
		defer wg.Done()
		res, err := r.Result(test_round.Player2)
		t.Logf("received result:%s, err:%v", res, err)
	}(test_round)

	wg.Wait()

	t.Log(test_round)
}

func Test8_NewRoundUnauthorized(t *testing.T) {

	test_round = NewRound()

	t.Logf("received round:%v+", test_round)

	res, err := test_round.Step(scissors, uuid.NewV4().String())
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(uuid.NewV4().String())
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)
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
