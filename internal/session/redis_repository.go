package session

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{
		client: client,
	}
}

func (r *RedisRepository) Create(ctx context.Context, session *Session) error {
	bld := strings.Builder{}
	bld.WriteString("session:")
	bld.WriteString(session.TokenHash)
	res, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to convert session into json: %w", err)
	}

	if err := r.client.SetArgs(ctx, bld.String(), res, redis.SetArgs{ExpireAt: session.ExpiresAt}).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

func (r *RedisRepository) FindByHash(ctx context.Context, hash string) (*Session, error) {
	bld := strings.Builder{}
	bld.WriteString("session:")
	bld.WriteString(hash)
	sessionJSON, err := r.client.Get(ctx, bld.String()).Result()
	if err == redis.Nil {
		return nil, ErrSessionNotFound
	} else if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to convert json into session: %w", err)
	}

	return &session, nil
}

func (r *RedisRepository) Delete(ctx context.Context, hash string) error {
	bld := strings.Builder{}
	bld.WriteString("session:")
	bld.WriteString(hash)
	return r.client.Del(ctx, bld.String()).Err()
}

func (r *RedisRepository) Extend(ctx context.Context, hash string, expiresAt time.Time) (*Session, error) {
	bld := strings.Builder{}
	bld.WriteString("session:")
	bld.WriteString(hash)
	var updatedSession *Session
	fn := func(tx *redis.Tx) error {
		sessionJSON, err := tx.Get(ctx, bld.String()).Result()
		if err == redis.Nil {
			return ErrSessionNotFound
		} else if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		var session Session
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			return fmt.Errorf("failed to convert json into session: %w", err)
		}

		if session.ExpiresAt.Before(time.Now()) {
			return ErrSessionExpired
		}

		if session.ExpiresAt.Before(expiresAt) {
			session.ExpiresAt = expiresAt
		}

		res, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to convert session into json: %w", err)
		}

		writeFn := func(pipe redis.Pipeliner) error {
			pipe.SetArgs(ctx, bld.String(), res, redis.SetArgs{ExpireAt: session.ExpiresAt})
			return nil
		}
		_, writeErr := tx.TxPipelined(ctx, writeFn)
		if writeErr != nil {
			return writeErr
		}
		updatedSession = &session
		return nil
	}
	err := r.client.Watch(ctx, fn, bld.String())
	if err != nil {
		return nil, fmt.Errorf("failed to extend session: %w", err)
	}

	return updatedSession, nil
}
