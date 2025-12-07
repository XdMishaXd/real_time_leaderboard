package models

type LeaderboardEntry struct {
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
	Rank   int64  `json:"rank"`
}

type UserRank struct {
	Score int64
	Rank  int64
}
