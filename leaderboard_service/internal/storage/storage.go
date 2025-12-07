package storage

import "errors"

var (
	ErrGameAlreadyExists = errors.New("Game already exists")
	ErrGameNotFound      = errors.New("Game not found")
)

const (
	SubmitScoreScript = `
		local game = KEYS[1]
		local user = ARGV[1]
		local score = tonumber(ARGV[2])
		local exists = redis.call("SISMEMBER", "games", game)
		if exists == 0 then
			return {err="GAME_NOT_FOUND"}
		end
		redis.call("ZADD", "leaderboard:"..game, score, user)
		return {ok="OK"}
	`

	EnsureGameExistsScript = `
		local game = KEYS[1]
		local exists = redis.call("SISMEMBER", "games", game)
		if exists == 1 then
			return {err="GAME_ALREADY_EXISTS"}
		end
		redis.call("SADD", "games", game)
		return {ok="OK"}
	`
)
