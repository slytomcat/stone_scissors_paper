package main

import (
	"errors"

	uuid "github.com/satori/go.uuid"
)

const (
	nothing int = iota
	first
	second
	draw

	stone
	scissors
	paper

	nobody = nothing
)

var (
	rules = map[int]map[int]int{
		stone:    {stone: draw, scissors: first, paper: second},
		scissors: {stone: second, scissors: draw, paper: first},
		paper:    {stone: first, scissors: second, paper: draw},
	}
)

type Round struct {
	player1 string // token for player1
	player2 string // token for player1
	bid1    int    // bid of player1
	bid2    int    // bid of player1
	winner  int    // index of the winner or 'nothing' if not all bids done
}

func NewRound() (Round, error) {
	return Round{
		player1: uuid.NewV4().String(),
		player2: uuid.NewV4().String(),
		bid1:    nothing,
		bid2:    nothing,
		winner:  nobody,
	}, nil
}

func (r *Round) Step(bid int, token string) (string, error) {
	if err := r.authorized(token); err != nil {
		return "", errors.New("Unauthorized")
	}
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
