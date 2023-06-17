package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"
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

// Round is a single round game provider
type Round struct {
	mx         sync.RWMutex // guard for async updates
	ID         string       `json:"id"`         // round id
	Player1    string       `json:"player1"`    // hash of player1's token
	Player2    string       `json:"player2"`    // hash of player2's token
	HiddenBet1 string       `json:"hiddenbet1"` // hidden bet of player1
	HiddenBet2 string       `json:"hiddenbet2"` // hidden bet of player2
	Bet1       int          `json:"bet1"`       // open bet of player1
	Bet2       int          `json:"bet2"`       // open bet of player2
	Winner     int          `json:"winner"`     // 'nobody' - not all bids done, 'first'|'second'|'draw' - winner selection when all bids done
	Signature  string       `json:"signature"`  // round signature (calculated without itself)
}

// NewRound returns new initialized Round
func NewRound(player1, player2 string) *Round {
	r := &Round{
		ID: uuid.NewString(),
	}

	r.Player1 = r.roundSaltedHash(player1)
	r.Player2 = r.roundSaltedHash(player2)
	r.Signature = r.roundSaltedHash(r)

	return r
}

// saltedHash returns first 32 symbos of BASE64 encoging of sha256(salt + obj)
func (r *Round) saltedHash(salt string, obj []byte) string {

	h := sha256.Sum256(append(obj, []byte(salt)...))

	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(h[:])
}

func (r *Round) roundSaltedHash(obj interface{}) string {

	salt := serverSalt + r.ID // make individual salt for each round object

	bObj, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return r.saltedHash(salt, bObj)
}

func (r *Round) check(player string) string {

	r.mx.Lock()
	defer r.mx.Unlock()

	// clear Signature to calculate round hash without it
	sign := r.Signature
	defer func() { r.Signature = sign }()
	r.Signature = ""

	if sign != r.roundSaltedHash(r) {
		return "round had been falsificated"
	}

	if err := r.authorized(player); err != nil {
		return "unauthorized"
	}
	return ""
}

// Bet makes the user's hidden bid
func (r *Round) Bet(hiddenBet, player string) string {
	if res := r.check(player); res != "" {
		return res
	}

	shPlayer := r.roundSaltedHash(player)

	// data racing prevention
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.Player1 == shPlayer && r.HiddenBet1 != "" ||
		r.Player2 == shPlayer && r.HiddenBet2 != "" {
		return "bet has already been placed"
	}

	if r.Player1 == shPlayer {
		r.HiddenBet1 = hiddenBet
	} else {
		r.HiddenBet2 = hiddenBet
	}

	// recalculate signature
	r.Signature = ""
	r.Signature = r.roundSaltedHash(r)

	return r.result(player)
}

// Disclose used to disclose the user's steps
func (r *Round) Disclose(secret, bet, player string) string {
	if res := r.check(player); res != "" {
		return res
	}

	if r.HiddenBet1 == "" || r.HiddenBet2 == "" {
		r.mx.RLock()
		defer r.mx.RUnlock()
		return r.result(player)
	}

	shPlayer := r.roundSaltedHash(player)

	shBet := r.saltedHash(secret, []byte(bet))

	if shPlayer == r.Player1 && r.HiddenBet1 != shBet ||
		shPlayer == r.Player2 && r.HiddenBet2 != shBet {
		return "Your bet is incorrect"
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	if shPlayer == r.Player1 {
		r.Bet1 = r.betEncode(bet)
	} else {
		r.Bet2 = r.betEncode(bet)
	}

	if r.Bet1 != nothing && r.Bet2 != nothing {
		// find the winner
		r.Winner = rules[r.Bet1][r.Bet2]

	}
	// recalculate signature
	r.Signature = ""
	r.Signature = r.roundSaltedHash(r)

	return r.result(player)

}

// Result returns the round result
func (r *Round) Result(player string) string {
	if res := r.check(player); res != "" {
		return res
	}
	// data racing prevention
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.result(player)
}

// result is not protected against data racing.
// It have to be called after mx.Lock() or mx.RLock()
func (r *Round) result(player string) string {

	hiddenBet, bet := "", nothing
	rHiddenBet, rBet := "", nothing
	cPlayer := 0
	if r.Player1 == r.roundSaltedHash(player) {
		hiddenBet = r.HiddenBet1
		bet = r.Bet1
		rHiddenBet = r.HiddenBet2
		rBet = r.Bet2
		cPlayer = first
	} else {
		hiddenBet = r.HiddenBet2
		bet = r.Bet2
		rHiddenBet = r.HiddenBet1
		rBet = r.Bet1
		cPlayer = second
	}

	if hiddenBet == "" {
		return "place Your bet, please"
	}

	if rHiddenBet == "" {
		return "wait for the rival to place its bet"
	}

	if bet == nothing {
		return "disclose your bet, please"
	}

	if rBet == nothing {
		return "wait for your rival to disclose its bet"
	}

	// there is a game result
	resp := ""

	if r.Winner == draw {
		resp = "draw"
	} else {
		if r.Winner == cPlayer {
			resp = "You won"
		} else {
			resp = "You lose"
		}
	}

	return fmt.Sprintf("%s: your bet: %s, the rival's bet: %s", resp, r.betDecode(bet), r.betDecode(rBet))
}

// authorized checks the user
func (r *Round) authorized(token string) error {
	if r.roundSaltedHash(token) != r.Player1 && r.roundSaltedHash(token) != r.Player2 {
		return errors.New("Unauthorized")
	}
	return nil
}

func (r *Round) betDecode(bet int) string {
	switch bet {
	case paper:
		return "paper"
	case scissors:
		return "scissors"
	case stone:
		return "stone"
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
