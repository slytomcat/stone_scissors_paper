package main

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type testDB struct {
	Err bool
}

func (d *testDB) Store(r *Round) error {
	fmt.Printf("stored: %v", r)
	return nil
}

func (d *testDB) Retrieve(key string) (*Round, error) {
	fmt.Printf("retrive: %s", key)
	if d.Err {
		return nil, errors.New("test error")
	}
	return NewRound(), nil
}

func Test_all(t *testing.T) {
	d := &testDB{Err: false}
	c := NewCache(d, 500*time.Millisecond, 50*time.Millisecond)

	r := NewRound()
	t.Log(r)

	err := c.Store(r)
	if err != nil {
		t.Error(err)
	}

	res := r.Step(stone, r.Player1)

	if res != "wait" {
		t.Error("wrong replay")
	}

	r1, err := c.Retrieve(r.ID)
	if err != nil {
		t.Error(err)
	}

	t.Log(r1)
	res = r1.Step(paper, r.Player1)

	<-time.After(time.Second)

	d.Err = true

	r2, err := c.Retrieve(r.ID)
	if err == nil {
		t.Error("no error when expected")
	}

	d.Err = false

	r2, err = c.Retrieve(r.ID)
	if err != nil {
		t.Error(err)
	}

	t.Log(r2)

	if r2.Bet1 == stone {
		t.Error("not deleted")
	}

}
