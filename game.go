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
	mx      sync.RWMutex // guard for async updates
	ID      string       `json:"id"`      // round id
	Player1 string       `json:"player1"` // token for player1
	Player2 string       `json:"player2"` // token for player1
	Bid1    int          `json:"bid1"`    // bid of player1
	Bid2    int          `json:"bid2"`    // bid of player1
	Winner  int          `json:"winner"`  // 'nobody' - not all bids done, 'first'|'second'|'draw' - winner selection when all bids done
}

// NewRound returns new initialized Round
func NewRound() *Round {
	return &Round{
		mx:      sync.RWMutex{},
		ID:      uuid.NewV4().String(),
		Player1: uuid.NewV4().String(),
		Player2: uuid.NewV4().String(),
		Bid1:    0,
		Bid2:    0,
		Winner:  0,
	}
}

// Step makes the user's bid
func (r *Round) Step(bid int, user string) string {
	if err := r.authorized(user); err != nil {
		return "Unauthorized"
	}
	if bid != paper && bid != scissors && bid != stone {
		return "wrong bid"
	}

	// data racing prevention
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.Player1 == user {
		if r.Bid1 != nothing {
			return "bid already done"
		}
		r.Bid1 = bid
	}
	if r.Player2 == user {
		if r.Bid2 != nothing {
			return "bid already done"
		}
		r.Bid2 = bid
	}
	if r.Bid1 != nothing && r.Bid2 != nothing {
		// find the winner
		r.Winner = rules[r.Bid1][r.Bid2]

		return r.result(user)
	}

	return "wait"
}

// Result returns the round result
func (r *Round) Result(user string) string {
	if err := r.authorized(user); err != nil {
		return "Unauthorized"
	}
	// data racing prevention
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.result(user)
}

// result is not protected from data racing. It should called after mx.Lock() or mx.RLock()
func (r *Round) result(user string) string {
	// check that winner is determined
	if r.Winner == nobody {
		return "wait"
	}
	// check for draw
	if r.Winner == draw {
		return "draw"
	}
	if user == r.Player1 {
		if r.Winner == first {
			return "you won"
		}
		return "you lose"
	}

	// token is authorized and it is not first player -> it is second player
	if r.Winner == second {
		return "you won"
	}
	return "you lose"

}

// authorized checks the user
func (r *Round) authorized(token string) error {
	if token != r.Player1 && token != r.Player2 {
		return errors.New("Unauthorized")
	}
	return nil
}
