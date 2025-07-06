package tickets

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

// Ticket fields for use in server
type AutotaskTicket struct {
	ID                 int64  `json:"id"`
	AssignedResourceID string `json:"assignedResourceID"`
	CreateDate         string `json:"createDate"`
	Description        string `json:"description"`
	Title              string `json:"title"`
	From               string `json:"from,omitempty"`
}

// tickets, hash, and mutex
type TicketCollection struct {
	sync.RWMutex
	Tickets *[]AutotaskTicket `json:"tickets"`
	Hash    string            `json:"hash"`
}

// computes hash of titles, returns true if hash has changed
func (tc *TicketCollection) CheckForNewHash() bool {
	unassignedTickets := tc.GetUnassignedTickets()
	titles := ""
	for _, t := range unassignedTickets {
		titles += t.Title
	}
	newHash := sha256.Sum256([]byte(titles))
	newHashStr := hex.EncodeToString(newHash[:])
	tc.Lock()
	defer tc.Unlock()
	if newHashStr != tc.Hash {
		tc.Hash = newHashStr
		return true
	}
	return false
}

// returns slice of tickets with blank resourceid
func (tc *TicketCollection) GetUnassignedTickets() []AutotaskTicket {
	tc.RLock()
	defer tc.RUnlock()
	unassignedTickets := make([]AutotaskTicket, 0)
	for _, ticket := range *tc.Tickets {
		if ticket.AssignedResourceID == "" {
			unassignedTickets = append(unassignedTickets, ticket)
		}
	}
	return unassignedTickets
}

// returns current hash value of tickets
func (tc *TicketCollection) GetCurrentHash() string {
	tc.RLock()
	defer tc.RUnlock()
	return tc.Hash
}

// sets tickets slice to new value
func (tc *TicketCollection) SetTickets(newTickets *[]AutotaskTicket) {
	tc.Lock()
	defer tc.Unlock()
	tc.Tickets = newTickets
}
