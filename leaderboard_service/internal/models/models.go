package models

type LeaderboardEntry struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Score    int64  `json:"score"`
	Rank     int64  `json:"rank"`
}

type UserRank struct {
	Score int64
	Rank  int64
}
