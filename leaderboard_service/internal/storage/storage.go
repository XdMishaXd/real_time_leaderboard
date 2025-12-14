package storage

import "errors"

var (
	ErrGameNotFound          = errors.New("Game not found")
	ErrNoResultsFound        = errors.New("No results found")
	ErrUserNotFound          = errors.New("User not found")
	ErrScoreLowerThenCurrent = errors.New("New score is lower then current")
)

const (
	SubmitScoreScript = `
		local game = KEYS[1]
		local userID = ARGV[1]
		local username = ARGV[2]
		local score = tonumber(ARGV[3])

		local leaderboardKey = "leaderboard:" .. game
		local usernamesKey = "usernames:" .. game

		local currentScore = redis.call("ZSCORE", leaderboardKey, userID)

		if currentScore and score <= tonumber(currentScore) then
			return {err = "NEW_SCORE_IS_LOWER_THEN_CURRENT"}
		end

		redis.call("ZADD", leaderboardKey, score, userID)

		redis.call("HSET", usernamesKey, userID, username)

		local rank = redis.call("ZREVRANK", leaderboardKey, userID)
		if not rank then
			return { err = "UNKNOWN_ERROR" }
		end

		return rank
	`

	EnsureGameExistsScript = `
		local game = KEYS[1]
		local exists = redis.call("SISMEMBER", "games", game)
		if exists == 1 then
			return {ok="OK"}
		end
		redis.call("SADD", "games", game)
		return {ok="OK"}
	`
)
