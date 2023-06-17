package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testDB struct {
	Err bool
}

func (d *testDB) Store(r *Round) error {
	return nil
}

func (d *testDB) Retrieve(key string) (*Round, error) {
	if d.Err {
		return nil, errors.New("test error")
	}
	return NewRound("", ""), nil
}

func Test_all(t *testing.T) {
	d := &testDB{Err: false}
	c := NewCache(d, 500*time.Millisecond, 50*time.Millisecond)

	player1 := "player1"
	player2 := "player2"
	r := NewRound(player1, player2)

	err := c.Store(r)
	require.NoError(t, err)

	res := r.Bet(r.saltedHash("my secret", []byte("paper")), player1)

	require.Equal(t, "wait for the rival to place its bet", res)

	r1, err := c.Retrieve(r.ID)
	require.NoError(t, err)

	_ = r1.Bet(r.saltedHash("my secret", []byte("paper")), player1)

	time.Sleep(time.Second)

	d.Err = true

	_, err = c.Retrieve(r.ID)
	require.Error(t, err)

	d.Err = false

	r2, err := c.Retrieve(r.ID)
	require.NoError(t, err)

	require.NotEqual(t, r2.Bet1, stone)
}
