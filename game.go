package main

import (
	"errors"
	"fmt"
	"strings"
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

// Round is a sigle raund game provider
type Round struct {
	mx      sync.RWMutex // guard for async updates
	ID      string       `json:"id"`      // round id
	Player1 string       `json:"player1"` // token for player1
	Player2 string       `json:"player2"` // token for player1
	Bet1    int          `json:"bid1"`    // bid of player1
	Bet2    int          `json:"bid2"`    // bid of player1
	Winner  int          `json:"winner"`  // 'nobody' - not all bids done, 'first'|'second'|'draw' - winner selection when all bids done
}

// NewRound returns new initialized Round
func NewRound() *Round {
	return &Round{
		mx:      sync.RWMutex{},
		ID:      uuid.NewV4().String(),
		Player1: uuid.NewV4().String(),
		Player2: uuid.NewV4().String(),
		Bet1:    nothing,
		Bet2:    nothing,
		Winner:  nobody,
	}
}

// Step makes the user's bid
func (r *Round) Step(bet int, user string) string {
	if err := r.authorized(user); err != nil {
		return "Unauthorized"
	}
	if bet != paper && bet != scissors && bet != stone {
		return "wrong bet"
	}

	// data racing prevention
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.Player1 == user && r.Bet1 != nothing || r.Player2 == user && r.Bet2 != nothing {
		return "bet has already been placed"
	}

	if r.Player1 == user {
		r.Bet1 = bet
	} else {
		r.Bet2 = bet
	}
	if r.Bet1 != nothing && r.Bet2 != nothing {
		// find the winner
		r.Winner = rules[r.Bet1][r.Bet2]

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

// result is not protected against data racing.
// It have to be called after mx.Lock() or mx.RLock()
func (r *Round) result(user string) string {
	// check that winner is determined
	if r.Winner == nobody {
		return "wait"
	}
	// check for draw
	if r.Winner == draw {
		return "draw"
	}

	resp := ""
	if r.Winner == first && user == r.Player1 ||
		r.Winner == second && user == r.Player2 {
		resp = "You won"
	} else {
		resp = "You lose"
	}

	ybet, rbet := nothing, nothing
	if user == r.Player1 {
		ybet, rbet = r.Bet1, r.Bet2
	} else {
		ybet, rbet = r.Bet2, r.Bet1

	}

	return fmt.Sprintf("%s: your bet: %s, the rival's bet: %s", resp, r.betDecode(ybet), r.betDecode(rbet))
}

// authorized checks the user
func (r *Round) authorized(token string) error {
	if token != r.Player1 && token != r.Player2 {
		return errors.New("Unauthorized")
	}
	return nil
}

func (r *Round) betDecode(bet int) string {
	switch bet {
	case paper:
		return "Paper"
	case scissors:
		return "Scissors"
	case stone:
		return "Stone"
	default:
		return ""
	}
}
func (r *Round) betEncode(bet string) int {
	switch strings.ToLower(bet) {
	case "paper":
		return paper
	case "scissors":
		return scissors
	case "stone":
		return stone
	default:
		return -1
	}
}
