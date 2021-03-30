package main

import (
	"errors"
	"sync"

	uuid "github.com/satori/go.uuid"
)

const (
	// winner selection
	nobody int = iota
	first
	second
	draw
	// bids
	stone
	scissors
	paper
	nothing = nobody
)

var (
	// rules - determines the winner by first and second bids
	rules = map[int]map[int]int{
		stone:    {stone: draw, scissors: first, paper: second},
		scissors: {stone: second, scissors: draw, paper: first},
		paper:    {stone: first, scissors: second, paper: draw},
	}
)

type Round struct {
	mx      sync.Mutex // guard for async updates
	ID      string     `json:"id"`      // round id
	Player1 string     `json:"player1"` // token for player1
	Player2 string     `json:"player2"` // token for player1
	Bid1    int        `json:"bid1"`    // bid of player1
	Bid2    int        `json:"bid2"`    // bid of player1
	Winner  int        `json:"winner"`  // 'nobody' - not all bids done, 'first'|'second'|'draw' - winner selection when all bids done
}

func NewRound() *Round {
	return &Round{
		mx:      sync.Mutex{},
		ID:      uuid.NewV4().String(),
		Player1: uuid.NewV4().String(),
		Player2: uuid.NewV4().String(),
		Bid1:    0,
		Bid2:    0,
		Winner:  0,
	}
}

func (r *Round) Step(bid int, token string) (string, error) {
	if err := r.authorized(token); err != nil {
		return "", errors.New("Unauthorized")
	}

	// data racing prevention
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.Player1 == token {
		if r.Bid1 != 0 {
			return "", errors.New("bid already done")
		}
		r.Bid1 = bid
	}
	if r.Player2 == token {
		if r.Bid2 != nothing {
			return "", errors.New("bid already done")
		}
		r.Bid2 = bid
	}
	if r.Bid1 != nothing && r.Bid2 != nothing {
		// find the winner
		r.Winner = rules[r.Bid1][r.Bid2]

		return r.Result(token)
	}

	return "wait", nil
}

func (r *Round) Result(token string) (string, error) {
	if err := r.authorized(token); err != nil {
		return "", errors.New("Unauthorized")
	}

	// check that winner is determined
	if r.Winner == nobody {
		return "wait", nil
	}
	// check for draw
	if r.Winner == draw {
		return "draw", nil
	}
	if token == r.Player1 {
		if r.Winner == first {
			return "you won", nil
		}
		return "you lose", nil
	}

	// token is authorized and it is not first player -> it is second player
	if r.Winner == second {
		return "you won", nil
	}
	return "you lose", nil

}

func (r *Round) authorized(token string) error {
	if token != r.Player1 && token != r.Player2 {
		return errors.New("Unauthorized")
	}
	return nil
}
