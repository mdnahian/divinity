package faction

import (
	"fmt"
	"math/rand"
	"sync/atomic"
)

var factionCounter atomic.Int64
var contractCounter atomic.Int64

type FactionTypeDef struct {
	StatKey   string
	Threshold int
	Label     string
}

var FactionTypes = map[string]FactionTypeDef{
	"political": {StatKey: "dominance", Threshold: 60, Label: "Political"},
	"economic":  {StatKey: "greed", Threshold: 55, Label: "Economic Guild"},
	"religious": {StatKey: "spiritual_sensitivity", Threshold: 60, Label: "Religious Order"},
	"criminal":  {StatKey: "aggression", Threshold: 55, Label: "Criminal Ring"},
	"scholarly": {StatKey: "intelligence", Threshold: 60, Label: "Scholarly Circle"},
}

type Faction struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Type                   string   `json:"type"`
	LeaderID               string   `json:"leaderId"`
	LeaderName             string   `json:"leaderName"`
	MemberIDs              []string `json:"memberIds"`
	FoundedDay             int      `json:"foundedDay"`
	Treasury               int      `json:"treasury"`
	MembershipFee          int      `json:"membershipFee"`
	Goal                   string   `json:"goal"`
	FactionCutPct          int      `json:"factionCutPct"`
	AllowExternalContracts bool     `json:"allowExternalContracts"`
}

// FactionContract represents a free-form commissioned job within or for a faction.
type FactionContract struct {
	ID                   string `json:"id"`
	FactionID            string `json:"factionId"`
	RequesterID          string `json:"requesterId"`
	RequesterName        string `json:"requesterName"`
	WorkerID             string `json:"workerId"`
	WorkerName           string `json:"workerName"`
	Description          string `json:"description"`
	FulfillmentCondition string `json:"fulfillmentCondition"`
	Payment              int    `json:"payment"`
	FactionCut           int    `json:"factionCut"`
	EscrowGold           int    `json:"escrowGold"`
	PostedDay            int    `json:"postedDay"`
	DueDay               int    `json:"dueDay"`
	RejectionCount       int    `json:"rejectionCount"`
	WorkerReport         string `json:"workerReport"`
	Status               string `json:"status"` // "open"|"accepted"|"pending_review"|"completed"|"expired"|"abandoned"
}

func NewContract(factionID, requesterID, requesterName, description, condition string, payment, factionCut, postedDay int) *FactionContract {
	escrow := payment + factionCut
	return &FactionContract{
		ID:                   fmt.Sprintf("contract_%d", contractCounter.Add(1)),
		FactionID:            factionID,
		RequesterID:          requesterID,
		RequesterName:        requesterName,
		Description:          description,
		FulfillmentCondition: condition,
		Payment:              payment,
		FactionCut:           factionCut,
		EscrowGold:           escrow,
		PostedDay:            postedDay,
		DueDay:               postedDay + 7,
		Status:               "open",
	}
}

func GenerateName(ftype, leaderName string) string {
	prefixes := map[string][]string{
		"political": {"Council of", "Order of", "Alliance of"},
		"economic":  {"Guild of", "Trade House of", "Merchants of"},
		"religious": {"Followers of", "Shrine Keepers of", "Brothers of"},
		"criminal":  {"Shadow of", "Rats of", "Night Watch of"},
		"scholarly": {"Circle of", "Learned of", "Scribes of"},
	}
	opts := prefixes[ftype]
	if len(opts) == 0 {
		opts = prefixes["political"]
	}
	return fmt.Sprintf("%s %s", opts[rand.Intn(len(opts))], leaderName)
}

func NewFaction(ftype, leaderID, leaderName string, memberIDs []string, day int) *Faction {
	return &Faction{
		ID:            fmt.Sprintf("faction_%d", factionCounter.Add(1)),
		Name:          GenerateName(ftype, leaderName),
		Type:          ftype,
		LeaderID:      leaderID,
		LeaderName:    leaderName,
		MemberIDs:     memberIDs,
		FoundedDay:    day,
		FactionCutPct: 10,
	}
}
