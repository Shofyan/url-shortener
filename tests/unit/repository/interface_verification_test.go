package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/Shofyan/url-shortener/internal/domain/entity"
	"github.com/Shofyan/url-shortener/internal/domain/repository"
	"github.com/Shofyan/url-shortener/internal/domain/valueobject"
)

// This ensures that all implementations (PostgreSQL, Redis, In-Memory) behave consistently.
type URLRepositoryTestSuite struct {
	suite.Suite
	repo repository.URLRepository
}

// NewURLRepositoryTestSuite creates a new test suite for the given repository.
func NewURLRepositoryTestSuite(repo repository.URLRepository) *URLRepositoryTestSuite {
	return &URLRepositoryTestSuite{repo: repo}
}

// SetupTest prepares each test case.
func (suite *URLRepositoryTestSuite) SetupTest() {
	// Clean up any existing test data
	// Implementation would depend on the repository type
}

// TestSave_Success tests successful URL saving.
func (suite *URLRepositoryTestSuite) TestSave_Success() {
	ctx := context.Background()

	// Create test URL
	shortKey, err := valueobject.NewShortKey("test123")
	require.NoError(suite.T(), err)

	longURL, err := valueobject.NewLongURL("https://example.com")
	require.NoError(suite.T(), err)

	url := entity.NewURL(shortKey, longURL)
	url.ID = 12345
	expiresAt := time.Now().Add(time.Hour)
	url.ExpiresAt = &expiresAt

	// Save URL
	err = suite.repo.Save(ctx, url)
	assert.NoError(suite.T(), err)

	// Verify it can be retrieved
	retrieved, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), retrieved)

	assert.Equal(suite.T(), url.ID, retrieved.ID)
	assert.Equal(suite.T(), url.ShortKey.Value(), retrieved.ShortKey.Value())
	assert.Equal(suite.T(), url.LongURL.Value(), retrieved.LongURL.Value())
	assert.Equal(suite.T(), url.VisitCount, retrieved.VisitCount)
	assert.WithinDuration(suite.T(), url.CreatedAt, retrieved.CreatedAt, time.Second)

	if url.ExpiresAt != nil {
		require.NotNil(suite.T(), retrieved.ExpiresAt)
		assert.WithinDuration(suite.T(), *url.ExpiresAt, *retrieved.ExpiresAt, time.Second)
	}
}

// TestFindByShortKey_Success tests successful retrieval by short key.
func (suite *URLRepositoryTestSuite) TestFindByShortKey_Success() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("find123")
	longURL, _ := valueobject.NewLongURL("https://findme.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 54321

	// Save first
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Test retrieval
	found, err := suite.repo.FindByShortKey(ctx, shortKey)

	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), found)
	assert.Equal(suite.T(), "find123", found.ShortKey.Value())
	assert.Equal(suite.T(), "https://findme.com", found.LongURL.Value())
}

// TestFindByShortKey_NotFound tests retrieval of non-existent URL.
func (suite *URLRepositoryTestSuite) TestFindByShortKey_NotFound() {
	ctx := context.Background()

	nonExistentKey, _ := valueobject.NewShortKey("notfound")

	found, err := suite.repo.FindByShortKey(ctx, nonExistentKey)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestFindByLongURL_Success tests successful retrieval by long URL.
func (suite *URLRepositoryTestSuite) TestFindByLongURL_Success() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("long123")
	longURL, _ := valueobject.NewLongURL("https://longurl.example.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 98765

	// Save first
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Test retrieval
	found, err := suite.repo.FindByLongURL(ctx, longURL)

	assert.NoError(suite.T(), err)
	require.NotNil(suite.T(), found)
	assert.Equal(suite.T(), "long123", found.ShortKey.Value())
	assert.Equal(suite.T(), "https://longurl.example.com", found.LongURL.Value())
}

// TestUpdate_Success tests successful URL updates.
func (suite *URLRepositoryTestSuite) TestUpdate_Success() {
	ctx := context.Background()

	// Setup initial data
	shortKey, _ := valueobject.NewShortKey("update123")
	longURL, _ := valueobject.NewLongURL("https://update.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 11111

	// Save initially
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Modify and update
	url.VisitCount = 42
	lastAccessed := time.Now()
	url.LastAccessedAt = &lastAccessed

	err = suite.repo.Update(ctx, url)
	assert.NoError(suite.T(), err)

	// Verify update
	updated, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), int64(42), updated.VisitCount)

	if url.LastAccessedAt != nil {
		require.NotNil(suite.T(), updated.LastAccessedAt)
		assert.WithinDuration(suite.T(), *url.LastAccessedAt, *updated.LastAccessedAt, time.Second)
	}
}

// TestDelete_Success tests successful URL deletion.
func (suite *URLRepositoryTestSuite) TestDelete_Success() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("delete123")
	longURL, _ := valueobject.NewLongURL("https://delete.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 22222

	// Save first
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Verify it exists
	found, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), found)

	// Delete
	err = suite.repo.Delete(ctx, shortKey)
	assert.NoError(suite.T(), err)

	// Verify it's gone
	found, err = suite.repo.FindByShortKey(ctx, shortKey)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestExistsByShortKey_Success tests existence checking.
func (suite *URLRepositoryTestSuite) TestExistsByShortKey_Success() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("exists123")
	longURL, _ := valueobject.NewLongURL("https://exists.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 33333

	// Initially should not exist
	exists, err := suite.repo.ExistsByShortKey(ctx, shortKey)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), exists)

	// Save
	err = suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Now should exist
	exists, err = suite.repo.ExistsByShortKey(ctx, shortKey)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
}

// TestIncrementVisitCount_Success tests atomic visit count increment.
func (suite *URLRepositoryTestSuite) TestIncrementVisitCount_Success() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("increment123")
	longURL, _ := valueobject.NewLongURL("https://increment.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 44444

	// Save initially
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Get initial state
	initial, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	initialCount := initial.VisitCount

	// Increment
	err = suite.repo.IncrementVisitCount(ctx, shortKey)
	assert.NoError(suite.T(), err)

	// Verify increment
	updated, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), initialCount+1, updated.VisitCount)
	assert.NotNil(suite.T(), updated.LastAccessedAt, "LastAccessedAt should be set after increment")
}

// TestIncrementVisitCount_NotFound tests increment on non-existent URL.
func (suite *URLRepositoryTestSuite) TestIncrementVisitCount_NotFound() {
	ctx := context.Background()

	nonExistentKey, _ := valueobject.NewShortKey("noincrement")

	err := suite.repo.IncrementVisitCount(ctx, nonExistentKey)
	assert.Error(suite.T(), err)
}

// TestConcurrentIncrementVisitCount tests thread safety of visit count increment.
func (suite *URLRepositoryTestSuite) TestConcurrentIncrementVisitCount() {
	if testing.Short() {
		suite.T().Skip("Skipping concurrency test in short mode")
	}

	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("concurrent123")
	longURL, _ := valueobject.NewLongURL("https://concurrent.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 55555

	// Save initially
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Get initial count
	initial, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	initialCount := initial.VisitCount

	// Perform concurrent increments
	concurrentIncrements := 50
	errChan := make(chan error, concurrentIncrements)

	for i := 0; i < concurrentIncrements; i++ {
		go func() {
			errChan <- suite.repo.IncrementVisitCount(ctx, shortKey)
		}()
	}

	// Collect results
	var incrementErrors []error

	for i := 0; i < concurrentIncrements; i++ {
		if err := <-errChan; err != nil {
			incrementErrors = append(incrementErrors, err)
		}
	}

	// Verify no errors
	assert.Empty(suite.T(), incrementErrors, "Concurrent increments should not fail")

	// Verify final count
	final, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	expectedCount := initialCount + int64(concurrentIncrements)
	assert.Equal(suite.T(), expectedCount, final.VisitCount,
		"Visit count should be exactly %d after %d concurrent increments",
		expectedCount, concurrentIncrements)
}

// TestLastAccessedAt_Persistence tests that LastAccessedAt is properly persisted.
func (suite *URLRepositoryTestSuite) TestLastAccessedAt_Persistence() {
	ctx := context.Background()

	// Setup test data
	shortKey, _ := valueobject.NewShortKey("lastaccess123")
	longURL, _ := valueobject.NewLongURL("https://lastaccess.com")
	url := entity.NewURL(shortKey, longURL)
	url.ID = 66666

	// Set initial LastAccessedAt
	initialAccess := time.Now().Add(-time.Hour)
	url.LastAccessedAt = &initialAccess

	// Save
	err := suite.repo.Save(ctx, url)
	require.NoError(suite.T(), err)

	// Retrieve and verify
	retrieved, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	require.NotNil(suite.T(), retrieved.LastAccessedAt)
	assert.WithinDuration(suite.T(), initialAccess, *retrieved.LastAccessedAt, time.Second)

	// Increment visit count (should update LastAccessedAt)
	err = suite.repo.IncrementVisitCount(ctx, shortKey)
	require.NoError(suite.T(), err)

	// Verify LastAccessedAt was updated
	updated, err := suite.repo.FindByShortKey(ctx, shortKey)
	require.NoError(suite.T(), err)

	require.NotNil(suite.T(), updated.LastAccessedAt)
	assert.True(suite.T(), updated.LastAccessedAt.After(initialAccess),
		"LastAccessedAt should be updated after increment")
}

// RunURLRepositoryTests runs the complete test suite against a repository implementation.
func RunURLRepositoryTests(t *testing.T, repo repository.URLRepository) {
	suite.Run(t, NewURLRepositoryTestSuite(repo))
}

// Example of how to use this test suite with different implementations:

// TestPostgreSQLRepository runs the interface tests against PostgreSQL implementation
/*
func TestPostgreSQLRepository(t *testing.T) {
	// Setup PostgreSQL test database
	db := setupPostgreSQLTestDB(t) // Implementation specific
	defer db.Close()

	repo := postgres.NewURLRepository(db)
	RunURLRepositoryTests(t, repo)
}
*/

// TestInMemoryRepository runs the interface tests against in-memory implementation
/*
func TestInMemoryRepository(t *testing.T) {
	repo := memory.NewURLRepository() // Implementation specific
	RunURLRepositoryTests(t, repo)
}
*/

// MockRepositoryCompliance tests that our mock implementations conform to the interface.
func TestMockRepositoryCompliance(t *testing.T) {
	// This test ensures our mocks properly implement the interface
	var _ repository.URLRepository = new(MockURLRepository)

	// Test basic mock functionality
	mock := new(MockURLRepository)
	ctx := context.Background()
	shortKey, _ := valueobject.NewShortKey("mock123")

	// Test that methods can be called (even if they return errors)
	mock.On("FindByShortKey", ctx, shortKey).Return(nil, errors.New("mock error"))

	_, err := mock.FindByShortKey(ctx, shortKey)
	assert.Error(t, err)
	mock.AssertExpectations(t)
}

// MockURLRepository for testing using testify/mock.
type MockURLRepository struct {
	mock.Mock
}

func (m *MockURLRepository) Save(ctx context.Context, url *entity.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) FindByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (*entity.URL, error) {
	args := m.Called(ctx, shortKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.URL), args.Error(1)
}

func (m *MockURLRepository) FindByLongURL(ctx context.Context, longURL *valueobject.LongURL) (*entity.URL, error) {
	args := m.Called(ctx, longURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.URL), args.Error(1)
}

func (m *MockURLRepository) Update(ctx context.Context, url *entity.URL) error {
	args := m.Called(ctx, url)
	return args.Error(0)
}

func (m *MockURLRepository) Delete(ctx context.Context, shortKey *valueobject.ShortKey) error {
	args := m.Called(ctx, shortKey)
	return args.Error(0)
}

func (m *MockURLRepository) ExistsByShortKey(ctx context.Context, shortKey *valueobject.ShortKey) (bool, error) {
	args := m.Called(ctx, shortKey)
	return args.Bool(0), args.Error(1)
}

func (m *MockURLRepository) IncrementVisitCount(ctx context.Context, shortKey *valueobject.ShortKey) error {
	args := m.Called(ctx, shortKey)
	return args.Error(0)
}
