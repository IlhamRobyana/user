package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/IlhamRobyana/user/configs"
	"github.com/IlhamRobyana/user/internal/domain/user/model"
	"github.com/IlhamRobyana/user/internal/domain/user/model/dto"
	"github.com/IlhamRobyana/user/internal/domain/user/repository"
	"github.com/IlhamRobyana/user/internal/domain/user/service"
	"github.com/IlhamRobyana/user/shared/failure"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *model.User, fieldsInsert ...repository.UserField) error {
	args := m.Called(ctx, user, fieldsInsert)
	return args.Error(0)
}

func (m *MockUserRepository) ResolveUserByID(ctx context.Context, userID uuid.UUID, selectFields ...repository.UserField) (model.User, error) {
	args := m.Called(ctx, userID, selectFields)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) ResolveUserByEmail(ctx context.Context, email string, selectFields ...repository.UserField) (model.User, error) {
	args := m.Called(ctx, email, selectFields)
	return args.Get(0).(model.User), args.Error(1)
}

func (m *MockUserRepository) IsExistUserByID(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func TestCreateUser(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockRedis := new(MockRedisClient)

	cfg := &configs.Config{
		Internal: configs.Internal{
			MaxLoginAttempt: 3,
			LoginAttemptTTL: time.Minute * 2,
		},
	}

	userService := &service.UserServiceImpl{
		UserRepository: mockRepo,
		cfg:            cfg,
		cache:          mockRedis,
	}

	ctx := context.Background()
	testID := uuid.New()

	// Test case 1: Successful user creation
	t.Run("Success", func(t *testing.T) {
		// Input
		createRequest := dto.UserCreateRequest{
			Email:    "test@example.com",
			Password: "password123",
			Fullname: "Test User",
		}

		// Expected model after conversion
		expectedUser := model.User{
			Id:        testID,
			Email:     "test@example.com",
			Password:  "hashed_password", // This would be the hashed password
			Fullname:  "Test User",
			CreatedBy: testID.String(),
			UpdatedBy: testID.String(),
		}

		// Mock behavior
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		// Execute
		response, err := userService.CreateUser(ctx, createRequest)

		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, response)
		mockRepo.AssertExpectations(t)
	})

	// Test case 2: Repository error
	t.Run("RepositoryError", func(t *testing.T) {
		// Input
		createRequest := dto.UserCreateRequest{
			Email:    "test@example.com",
			Password: "password123",
			Fullname: "Test User",
		}

		// Mock behavior
		mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*model.User")).Return(errors.New("database error"))

		// Execute
		response, err := userService.CreateUser(ctx, createRequest)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dto.UserResponse{}, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestResolveUserByID(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockRedis := new(MockRedisClient)

	cfg := &configs.Config{
		Internal: configs.Internal{
			MaxLoginAttempt: 3,
			LoginAttemptTTL: time.Minute * 2,
		},
	}

	userService := &service.UserServiceImpl{
		UserRepository: mockRepo,
		cfg:            cfg,
		cache:          mockRedis,
	}

	ctx := context.Background()
	testID := uuid.New()

	// Test case 1: User found
	t.Run("UserFound", func(t *testing.T) {
		// Expected user
		expectedUser := model.User{
			Id:       testID,
			Email:    "test@example.com",
			Password: "hashed_password",
			Fullname: "Test User",
		}

		// Mock behavior
		mockRepo.On("ResolveUserByID", ctx, testID, mock.Anything).Return(expectedUser, nil)

		// Execute
		response, err := userService.ResolveUserByID(ctx, testID)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, testID, response.Id)
		assert.Equal(t, "test@example.com", response.Email)
		assert.Equal(t, "Test User", response.Fullname)
		mockRepo.AssertExpectations(t)
	})

	// Test case 2: User not found
	t.Run("UserNotFound", func(t *testing.T) {
		// Mock behavior
		notFoundErr := failure.NotFound("user with id not found")
		mockRepo.On("ResolveUserByID", ctx, testID, mock.Anything).Return(model.User{}, notFoundErr)

		// Execute
		response, err := userService.ResolveUserByID(ctx, testID)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, dto.UserResponse{}, response)
		mockRepo.AssertExpectations(t)
	})
}

func TestLoginUser(t *testing.T) {
	// Setup
	mockRepo := new(MockUserRepository)
	mockRedis := new(MockRedisClient)

	cfg := &configs.Config{
		Internal: configs.Internal{
			MaxLoginAttempt: 3,
			LoginAttemptTTL: time.Minute * 2,
		},
	}

	userService := &service.UserServiceImpl{
		UserRepository: mockRepo,
		cfg:            cfg,
		cache:          mockRedis,
	}

	ctx := context.Background()

	// Test case 1: Successful login
	t.Run("SuccessfulLogin", func(t *testing.T) {
		// Input
		loginRequest := dto.UserLoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock user from repository
		user := model.User{
			Id:       uuid.New(),
			Email:    "test@example.com",
			Password: "hashed_password", // This would be the hashed password
			Fullname: "Test User",
		}

		// Mock GetLoginAttempt behavior - no previous attempts
		nilRedisErr := redis.NewStringResult("", redis.Nil)
		mockRedis.On("Get", ctx, "login_attempt:test@example.com").Return(nilRedisErr)

		// Mock SetLoginAttempt behavior - initial setup with 0 attempts
		statusCmd := redis.NewStatusResult("OK", nil)
		mockRedis.On("Set", ctx, "login_attempt:test@example.com", "0", cfg.Internal.LoginAttemptTTL).Return(statusCmd)

		// Mock ResolveUserByEmail behavior
		mockRepo.On("ResolveUserByEmail", ctx, "test@example.com").Return(user, nil)

		// We need to mock the password comparison
		// Since we can't mock the ComparePassword method directly (it's on model.User),
		// we'll need to use a custom implementation or mock strategy

		// Execute with mock password comparison
		isPasswordMatch, err := userService.LoginUser(ctx, loginRequest)

		// Assert
		assert.NoError(t, err)
		assert.True(t, isPasswordMatch)
		mockRepo.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	// Test case 2: User not found
	t.Run("UserNotFound", func(t *testing.T) {
		// Input
		loginRequest := dto.UserLoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		// Mock GetLoginAttempt behavior - no previous attempts
		nilRedisErr := redis.NewStringResult("", redis.Nil)
		mockRedis.On("Get", ctx, "login_attempt:nonexistent@example.com").Return(nilRedisErr)

		// Mock SetLoginAttempt behavior - initial setup with 0 attempts
		statusCmd := redis.NewStatusResult("OK", nil)
		mockRedis.On("Set", ctx, "login_attempt:nonexistent@example.com", "0", cfg.Internal.LoginAttemptTTL).Return(statusCmd)

		// Mock ResolveUserByEmail behavior - user not found
		notFoundErr := failure.NotFound("user not found")
		mockRepo.On("ResolveUserByEmail", ctx, "nonexistent@example.com").Return(model.User{}, notFoundErr)

		// Mock SetLoginAttempt behavior - increment after failed login
		mockRedis.On("Set", ctx, "login_attempt:nonexistent@example.com", "1", cfg.Internal.LoginAttemptTTL).Return(statusCmd)

		// Execute
		isPasswordMatch, err := userService.LoginUser(ctx, loginRequest)

		// Assert
		assert.Error(t, err)
		assert.False(t, isPasswordMatch)
		mockRepo.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})

	// Test case 3: Max login attempts exceeded
	t.Run("MaxLoginAttemptsExceeded", func(t *testing.T) {
		// Input
		loginRequest := dto.UserLoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Mock GetLoginAttempt behavior - max attempts reached
		stringCmd := redis.NewStringResult("3", nil) // 3 attempts already made
		mockRedis.On("Get", ctx, "login_attempt:test@example.com").Return(stringCmd)

		// Execute
		isPasswordMatch, err := userService.LoginUser(ctx, loginRequest)

		// Assert
		assert.Error(t, err)
		assert.False(t, isPasswordMatch)
		assert.Equal(t, http.StatusForbidden, failure.GetCode(err))
		mockRedis.AssertExpectations(t)
	})

	// Test case 4: Invalid password
	t.Run("InvalidPassword", func(t *testing.T) {
		// Input
		loginRequest := dto.UserLoginRequest{
			Email:    "test@example.com",
			Password: "wrong_password",
		}

		// Mock user from repository
		user := model.User{
			Id:       uuid.New(),
			Email:    "test@example.com",
			Password: "hashed_password", // This would be the hashed password
			Fullname: "Test User",
		}

		// Mock GetLoginAttempt behavior - no previous attempts
		nilRedisErr := redis.NewStringResult("", redis.Nil)
		mockRedis.On("Get", ctx, "login_attempt:test@example.com").Return(nilRedisErr)

		// Mock SetLoginAttempt behavior - initial setup with 0 attempts
		statusCmd := redis.NewStatusResult("OK", nil)
		mockRedis.On("Set", ctx, "login_attempt:test@example.com", "0", cfg.Internal.LoginAttemptTTL).Return(statusCmd)

		// Mock ResolveUserByEmail behavior
		mockRepo.On("ResolveUserByEmail", ctx, "test@example.com").Return(user, nil)

		// Mock SetLoginAttempt behavior - increment after failed login
		mockRedis.On("Set", ctx, "login_attempt:test@example.com", "1", cfg.Internal.LoginAttemptTTL).Return(statusCmd)

		// Execute with mock password comparison (returning false)
		isPasswordMatch, err := userService.LoginUser(ctx, loginRequest)

		// Assert
		assert.Error(t, err)
		assert.False(t, isPasswordMatch)
		assert.Equal(t, http.StatusUnauthorized, failure.GetCode(err))
		mockRepo.AssertExpectations(t)
		mockRedis.AssertExpectations(t)
	})
}

// Mock implementations for the cache methods
func (s *MockUserServiceImpl) GetLoginAttempt(ctx context.Context, email string) (int, error) {
	args := s.Called(ctx, email)
	return args.Int(0), args.Error(1)
}

func (s *MockUserServiceImpl) SetLoginAttempt(ctx context.Context, email string, attempt int, ttl time.Duration) error {
	args := s.Called(ctx, email, attempt, ttl)
	return args.Error(0)
}
