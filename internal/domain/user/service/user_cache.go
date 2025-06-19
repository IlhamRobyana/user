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

func (s *UserServiceImpl) GetSuspendAmountKey(email string) string {
	return fmt.Sprintf("user:suspend:amount:%s", email)
}

func (s *UserServiceImpl) GetSuspendAmount(ctx context.Context, email string) (int, error) {
	attemptStr, err := s.cache.Get(ctx, s.GetSuspendAmountKey(email)).Result()
	if err != nil {
		return 0, err
	}
	if attemptStr == "" {
		return 0, nil
	}
	return strconv.Atoi(attemptStr)
}

func (s *UserServiceImpl) SetSuspendAmount(ctx context.Context, email string, attempt int, ttl time.Duration) error {
	return s.cache.Set(ctx, s.GetSuspendAmountKey(email), attempt, ttl).Err()
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
