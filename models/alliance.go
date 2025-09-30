package models

import (
	"time"
)

// Alliance representa una alianza en el juego
type Alliance struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Tag         string    `json:"tag" db:"tag"`
	LeaderID    int       `json:"leader_id" db:"leader_id"`
	Level       int       `json:"level" db:"level"`
	Experience  int       `json:"experience" db:"experience"`
	MaxMembers  int       `json:"max_members" db:"max_members"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AllianceMember representa un miembro de una alianza
type AllianceMember struct {
	ID           int       `json:"id" db:"id"`
	AllianceID   int       `json:"alliance_id" db:"alliance_id"`
	PlayerID     int       `json:"player_id" db:"player_id"`
	Role         string    `json:"role" db:"role"` // leader, officer, member
	JoinedAt     time.Time `json:"joined_at" db:"joined_at"`
	Contribution int       `json:"contribution" db:"contribution"`
}

// AllianceWithMembers representa una alianza con sus miembros
type AllianceWithMembers struct {
	Alliance *Alliance        `json:"alliance"`
	Members  []AllianceMember `json:"members"`
	Leader   *Player          `json:"leader"`
}

// AllianceInvitation representa una invitación a una alianza
type AllianceInvitation struct {
	ID         int       `json:"id" db:"id"`
	AllianceID int       `json:"alliance_id" db:"alliance_id"`
	InviterID  int       `json:"inviter_id" db:"inviter_id"`
	InviteeID  int       `json:"invitee_id" db:"invitee_id"`
	Status     string    `json:"status" db:"status"` // pending, accepted, declined
	Message    string    `json:"message" db:"message"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
}

// AllianceWar representa una guerra entre alianzas
type AllianceWar struct {
	ID            int        `json:"id" db:"id"`
	AttackerID    int        `json:"attacker_id" db:"attacker_id"`
	DefenderID    int        `json:"defender_id" db:"defender_id"`
	Status        string     `json:"status" db:"status"` // declared, active, ended
	StartTime     time.Time  `json:"start_time" db:"start_time"`
	EndTime       *time.Time `json:"end_time" db:"end_time"`
	AttackerScore int        `json:"attacker_score" db:"attacker_score"`
	DefenderScore int        `json:"defender_score" db:"defender_score"`
	WinnerID      *int       `json:"winner_id" db:"winner_id"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// AllianceWarParticipant representa un participante en una guerra de alianzas
type AllianceWarParticipant struct {
	ID          int       `json:"id" db:"id"`
	WarID       int       `json:"war_id" db:"war_id"`
	PlayerID    int       `json:"player_id" db:"player_id"`
	AllianceID  int       `json:"alliance_id" db:"alliance_id"`
	Side        string    `json:"side" db:"side"` // attacker, defender
	Score       int       `json:"score" db:"score"`
	BattlesWon  int       `json:"battles_won" db:"battles_won"`
	BattlesLost int       `json:"battles_lost" db:"battles_lost"`
	JoinedAt    time.Time `json:"joined_at" db:"joined_at"`
}

// AllianceRanking representa el ranking de una alianza
type AllianceRanking struct {
	AllianceID   int    `json:"alliance_id" db:"alliance_id"`
	AllianceName string `json:"alliance_name" db:"alliance_name"`
	AllianceTag  string `json:"alliance_tag" db:"alliance_tag"`
	TotalPower   int    `json:"total_power" db:"total_power"`
	TotalScore   int    `json:"total_score" db:"total_score"`
	MemberCount  int    `json:"member_count" db:"member_count"`
	Rank         int    `json:"rank" db:"rank"`
}
