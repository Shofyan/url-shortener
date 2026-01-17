package usecase

import (
	"github.com/stretchr/testify/mock"
)

// MockURLRepository is a mock implementation of URLRepository for testing.
type MockURLRepository struct {
	mock.Mock
}

// MockCacheRepository is a mock implementation of CacheRepository for testing.
type MockCacheRepository struct {
	mock.Mock
}
