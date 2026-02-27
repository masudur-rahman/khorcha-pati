package wizard

import (
	"sync"
	"time"
)

// Step identifies the current step in the interactive transaction wizard.
type Step int

const (
	StepType        Step = iota // choose transaction type
	StepCategory                // choose category
	StepSubcategory             // choose subcategory
	StepWallet                  // choose wallet (from / to)
	StepAmount                  // enter amount
	StepDate                    // enter date
	StepNote                    // enter note
	StepConfirm                 // confirm and save
)

// State holds all data collected during an active wizard session.
// It lives server-side so callback_data stays tiny (≤ 64 bytes).
type State struct {
	Step        Step
	TxnType     string
	Category    string
	Subcategory string
	FromWallet  string
	ToWallet    string
	ContactID   int64
	Amount      float64
	Date        string
	Note        string
	ExpiresAt   time.Time
}

// Store is a thread-safe, in-memory wizard state store keyed by Telegram UserID.
type Store struct {
	mu     sync.Mutex
	states map[int64]*State
}

// NewStore creates an empty wizard Store.
func NewStore() *Store {
	return &Store{states: make(map[int64]*State)}
}

// Set stores (or replaces) state for a user with a 10-minute TTL.
func (s *Store) Set(userID int64, state *State) {
	state.ExpiresAt = time.Now().Add(10 * time.Minute)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[userID] = state
}

// Get retrieves state for a user. Returns (nil, false) if not found or expired.
func (s *Store) Get(userID int64) (*State, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.states[userID]
	if !ok {
		return nil, false
	}
	if time.Now().After(st.ExpiresAt) {
		delete(s.states, userID)
		return nil, false
	}
	return st, true
}

// Clear removes wizard state for a user (call on confirm or cancel).
func (s *Store) Clear(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, userID)
}

// PurgeExpired removes all expired entries. Call periodically if needed.
func (s *Store) PurgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, st := range s.states {
		if now.After(st.ExpiresAt) {
			delete(s.states, id)
		}
	}
}
