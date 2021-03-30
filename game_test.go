package main

import "testing"

var test_round Round

func Test1_NewRound(t *testing.T) {
	var err error
	test_round, err = NewRound()
	if err != nil {
		t.Error(err)
	}
	if test_round.player1 == test_round.player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+, err:%v", test_round, err)

	res, err := test_round.Step(paper, test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

}

func Test2_NewRound(t *testing.T) {
	var err error
	test_round, err = NewRound()
	if err != nil {
		t.Error(err)
	}
	if test_round.player1 == test_round.player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+, err:%v", test_round, err)

	res, err := test_round.Step(scissors, test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

}

func Test3_NewRound(t *testing.T) {
	var err error
	test_round, err = NewRound()
	if err != nil {
		t.Error(err)
	}
	if test_round.player1 == test_round.player2 {
		t.Error("two tokens are equal")
	}

	t.Logf("received round:%v+, err:%v", test_round, err)

	res, err := test_round.Step(scissors, test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(stone, test_round.player1)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(scissors, test_round.player2)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Step(paper, test_round.player2)
	if err == nil {
		t.Error("no error when expected")
	}

	t.Logf("received result:%s, err:%v", res, err)

	res, err = test_round.Result(test_round.player1)
	if err != nil {
		t.Error(err)
	}

	t.Logf("received result:%s, err:%v", res, err)

}
func Test9_authorized(t *testing.T) {
	err := test_round.authorized(test_round.player1)
	if err != nil {
		t.Error("first user unauthorized")
	}
	err = test_round.authorized(test_round.player2)
	if err != nil {
		t.Error("first user unauthorized")
	}
	err = test_round.authorized("56ccbb4c-5673-4c88-9d76-edde2f240052")
	if err == nil {
		t.Error("unauthorized user is authorized")
	}

}
