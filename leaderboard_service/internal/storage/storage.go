package storage

import "errors"

var (
	ErrGameNotFound   = errors.New("Game not found")
	ErrNoResultsFound = errors.New("No results found")
	ErrUserNotFound   = errors.New("User not found")
)

const (
	SubmitScoreScript = `
		local game = KEYS[1]
    local user = ARGV[1]
    local username = ARGV[2]  
    local score = tonumber(ARGV[3])

    local exists = redis.call("SISMEMBER", "games", game)
    if exists == 0 then
      return {err="GAME_NOT_FOUND"}
    end

    local leaderboardKey = "leaderboard:" .. game

    redis.call("ZADD", leaderboardKey, score, user)

    local rank = redis.call("ZREVRANK", leaderboardKey, user)
    if not rank then
      return {err="UNKNOWN_ERROR"}
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
