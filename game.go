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
	mx      *sync.Mutex // guard for async updates
	player1 string      // token for player1
	player2 string      // token for player1
	bid1    int         // bid of player1
	bid2    int         // bid of player1
	winner  int         // 'nobody' - not all bids done, 'first'|'second'|'draw' - winner selection when all bids done
}

func NewRound() *Round {
	return &Round{
		mx:      new(sync.Mutex),
		player1: uuid.NewV4().String(),
		player2: uuid.NewV4().String(),
		bid1:    nothing,
		bid2:    nothing,
		winner:  nobody,
	}
}

func (r *Round) Step(bid int, token string) (string, error) {
	if err := r.authorized(token); err != nil {
		return "", errors.New("Unauthorized")
	}

	// data racing prevention
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.player1 == token {
		if r.bid1 != 0 {
			return "", errors.New("bid already done")
		}
		r.bid1 = bid
	}
	if r.player2 == token {
		if r.bid2 != nothing {
			return "", errors.New("bid already done")
		}
		r.bid2 = bid
	}
	if r.bid1 != nothing && r.bid2 != nothing {
		// find the winner
		r.winner = rules[r.bid1][r.bid2]

		return r.Result(token)
	}

	return "wait", nil
}

func (r *Round) Result(token string) (string, error) {
	if err := r.authorized(token); err != nil {
		return "", errors.New("Unauthorized")
	}

	// check that winner is determined
	if r.winner == nobody {
		return "wait", nil
	}
	// check for draw
	if r.winner == draw {
		return "draw", nil
	}
	if token == r.player1 {
		if r.winner == first {
			return "you won", nil
		}
		return "you lose", nil
	}

	// token is authorized and it is not first player -> it is second player
	if r.winner == second {
		return "you won", nil
	}
	return "you lose", nil

}

func (r *Round) authorized(token string) error {
	if token != r.player1 && token != r.player2 {
		return errors.New("Unauthorized")
	}
	return nil
}
