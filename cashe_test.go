package main

import (
	"fmt"
	"testing"
	"time"
)

type testDB struct{}

func (d testDB) Store(r *Round) error {
	fmt.Printf("stored: %v", r)
	return nil
}

func (d testDB) Retrive(key string) (*Round, error) {
	fmt.Printf("retrive: %s", key)
	return NewRound(), nil
}

func Test_all(t *testing.T) {
	c := NewCache(testDB{}, 500*time.Millisecond, 50*time.Millisecond)

	r := NewRound()
	t.Log(r)

	err := c.Store(r)
	if err != nil {
		t.Error(err)
	}

	res, err := r.Step(stone, r.Player1)
	if err != nil {
		t.Error(err)
	}
	if res != "wait" {
		t.Error("wrong replay")
	}

	r1, err := c.Retrive(r.ID)
	if err != nil {
		t.Error(err)
	}

	t.Log(r1)
	res, err = r1.Step(paper, r.Player1)
	if err != nil {
		t.Log(err)
	}

	<-time.After(time.Second)

	r2, err := c.Retrive(r.ID)
	if err != nil {
		t.Error(err)
	}

	t.Log(r2)

	if r2.Bid1 == stone {
		t.Error("not deleted")
	}

}
