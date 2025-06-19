package service

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (s *UserServiceImpl) GetLoginAttemptKey(email string) string {
	return fmt.Sprintf("user:login:attempt:%s", email)
}

func (s *UserServiceImpl) GetLoginAttempt(ctx context.Context, email string) (int, error) {
	attemptStr, err := s.cache.Get(ctx, s.GetLoginAttemptKey(email)).Result()
	if err != nil {
		return 0, err
	}
	if attemptStr == "" {
		return 0, nil
	}
	return strconv.Atoi(attemptStr)
}

func (s *UserServiceImpl) SetLoginAttempt(ctx context.Context, email string, attempt int, ttl time.Duration) error {
	return s.cache.Set(ctx, s.GetLoginAttemptKey(email), attempt, ttl).Err()
}
