package repository

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	entity_profile "github.com/Tomelin/dashfin-backend-app/internal/core/entity/profile"
	"github.com/Tomelin/dashfin-backend-app/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock Firebase DB
type MockFirebaseDB struct {
	mock.Mock
}

func (m *MockFirebaseDB) Create(ctx context.Context, data map[string]interface{}, collection string) ([]byte, error) {
	args := m.Called(ctx, data, collection)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFirebaseDB) Update(ctx context.Context, id string, data map[string]interface{}, collection string) error {
	args := m.Called(ctx, id, data, collection)
	return args.Error(0)
}

func (m *MockFirebaseDB) Get(ctx context.Context, collection string) ([]byte, error) {
	args := m.Called(ctx, collection)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFirebaseDB) GetByFilter(ctx context.Context, filter map[string]interface{}, collection string) ([]byte, error) {
	args := m.Called(ctx, filter, collection)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFirebaseDB) Delete(ctx context.Context, id string, collection string) error {
	args := m.Called(ctx, id, collection)
	return args.Error(0)
}

// Test Suite
type ProfileRepositoryTestSuite struct {
	suite.Suite
	mockDB *MockFirebaseDB
	repo   ProfileRepositoryInterface
	ctx    context.Context
}

func TestProfileRepositorySuite(t *testing.T) {
	suite.Run(t, new(ProfileRepositoryTestSuite))
}

func (suite *ProfileRepositoryTestSuite) SetupTest() {
	suite.mockDB = new(MockFirebaseDB)
	suite.ctx = context.Background()
	
	repo, err := InicializeProfileRepository(suite.mockDB)
	assert.NoError(suite.T(), err)
	suite.repo = repo
}

func (suite *ProfileRepositoryTestSuite) TestInicializeProfileRepository() {
	tests := []struct {
		name        string
		db          database.FirebaseDBInterface
		expectError bool
	}{
		{
			name:        "Valid DB",
			db:          suite.mockDB,
			expectError: false,
		},
		{
			name:        "Nil DB",
			db:          nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			repo, err := InicializeProfileRepository(tt.db)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), repo)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), repo)
			}
		})
	}
}

func (suite *ProfileRepositoryTestSuite) TestCreateProfile() {
	tests := []struct {
		name        string
		profile     *entity_profile.Profile
		mockReturn  []byte
		mockError   error
		expectError bool
	}{
		{
			name: "Successful creation",
			profile: &entity_profile.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			mockReturn:  []byte(`{"id":"123"}`),
			mockError:   nil,
			expectError: false,
		},
		{
			name: "Successful creation with ID field",
			profile: &entity_profile.Profile{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane@example.com",
			},
			mockReturn:  []byte(`{"ID":"456"}`),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Nil profile",
			profile:     nil,
			expectError: true,
		},
		{
			name: "Database error",
			profile: &entity_profile.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			mockError:   errors.New("database error"),
			expectError: true,
		},
		{
			name: "Invalid JSON response",
			profile: &entity_profile.Profile{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
			},
			mockReturn:  []byte(`invalid json`),
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.profile != nil && tt.mockReturn != nil {
				suite.mockDB.On("Create", suite.ctx, mock.Anything, "profiles").
					Return(tt.mockReturn, tt.mockError).Once()
			}

			result, err := suite.repo.CreateProfile(suite.ctx, tt.profile)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.NotEmpty(suite.T(), result.ID)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileRepositoryTestSuite) TestGetProfileByID() {
	profiles := []entity_profile.Profile{
		{
			ID:             "123",
			FirstName:      "John",
			LastName:       "Doe",
			Email:          "john@example.com",
			UserProviderID: "provider123",
		},
	}
	
	tests := []struct {
		name        string
		id          *string
		mockReturn  []byte
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful retrieval",
			id:          stringPtr("provider123"),
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "User not found",
			id:          stringPtr("nonexistent"),
			mockReturn:  []byte(`[]`),
			mockError:   nil,
			expectError: true,
		},
		{
			name:        "Nil ID",
			id:          nil,
			expectError: true,
		},
		{
			name:        "Database error",
			id:          stringPtr("provider123"),
			mockError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "Invalid JSON response",
			id:          stringPtr("provider123"),
			mockReturn:  []byte(`invalid json`),
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.id != nil && tt.mockReturn != nil {
				expectedFilter := map[string]interface{}{
					"userProviderID": *tt.id,
				}
				suite.mockDB.On("GetByFilter", suite.ctx, expectedFilter, "profiles").
					Return(tt.mockReturn, tt.mockError).Once()
			}

			result, err := suite.repo.GetProfileByID(suite.ctx, tt.id)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Equal(suite.T(), "John", result.FirstName)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileRepositoryTestSuite) TestGetByFilter() {
	profiles := []entity_profile.Profile{
		{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
	}

	tests := []struct {
		name        string
		filter      map[string]interface{}
		mockReturn  []byte
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful filter",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Nil filter",
			filter:      nil,
			expectError: true,
		},
		{
			name:        "Database error",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "Invalid JSON response",
			filter:      map[string]interface{}{"email": "john@example.com"},
			mockReturn:  []byte(`invalid json`),
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.filter != nil && tt.mockReturn != nil {
				suite.mockDB.On("GetByFilter", suite.ctx, tt.filter, "profiles").
					Return(tt.mockReturn, tt.mockError).Once()
			}

			result, err := suite.repo.GetByFilter(suite.ctx, tt.filter)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Len(suite.T(), result, 1)
				assert.Equal(suite.T(), "John", result[0].FirstName)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileRepositoryTestSuite) TestGetProfile() {
	profiles := []entity_profile.Profile{
		{
			ID:        "123",
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john@example.com",
		},
		{
			ID:        "456",
			FirstName: "Jane",
			LastName:  "Smith",
			Email:     "jane@example.com",
		},
	}

	tests := []struct {
		name        string
		mockReturn  []byte
		mockError   error
		expectError bool
		expected    int
	}{
		{
			name:        "Successful retrieval",
			mockReturn:  mustMarshal(profiles),
			mockError:   nil,
			expectError: false,
			expected:    2,
		},
		{
			name:        "Database error",
			mockError:   errors.New("database error"),
			expectError: true,
		},
		{
			name:        "Invalid JSON response",
			mockReturn:  []byte(`invalid json`),
			mockError:   nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.mockDB.On("Get", suite.ctx, "profiles").
				Return(tt.mockReturn, tt.mockError).Once()

			result, err := suite.repo.GetProfile(suite.ctx)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Len(suite.T(), result, tt.expected)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

func (suite *ProfileRepositoryTestSuite) TestUpdateProfile() {
	profile := &entity_profile.Profile{
		ID:             "123",
		FirstName:      "John",
		LastName:       "Doe Updated",
		Email:          "john@example.com",
		UserProviderID: "provider123",
	}

	profiles := []entity_profile.Profile{*profile}

	tests := []struct {
		name         string
		profile      *entity_profile.Profile
		mockReturn   []byte
		mockError    error
		filterReturn []byte
		filterError  error
		expectError  bool
	}{
		{
			name:         "Successful update",
			profile:      profile,
			mockError:    nil,
			filterReturn: mustMarshal(profiles),
			filterError:  nil,
			expectError:  false,
		},
		{
			name:        "Nil profile",
			profile:     nil,
			expectError: true,
		},
		{
			name:        "Update error",
			profile:     profile,
			mockError:   errors.New("update failed"),
			expectError: true,
		},
		{
			name:         "Filter error after update",
			profile:      profile,
			mockError:    nil,
			filterError:  errors.New("filter failed"),
			expectError:  true,
		},
		{
			name:         "No user found after update",
			profile:      profile,
			mockError:    nil,
			filterReturn: []byte(`[]`),
			filterError:  nil,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.profile != nil {
				suite.mockDB.On("Update", suite.ctx, tt.profile.ID, mock.Anything, "profiles").
					Return(tt.mockError).Once()
				
				if tt.mockError == nil {
					expectedFilter := map[string]interface{}{
						"userProviderID": tt.profile.UserProviderID,
					}
					suite.mockDB.On("GetByFilter", suite.ctx, expectedFilter, "profiles").
						Return(tt.filterReturn, tt.filterError).Once()
				}
			}

			result, err := suite.repo.UpdateProfile(suite.ctx, tt.profile)
			
			if tt.expectError {
				assert.Error(suite.T(), err)
				assert.Nil(suite.T(), result)
			} else {
				assert.NoError(suite.T(), err)
				assert.NotNil(suite.T(), result)
				assert.Equal(suite.T(), tt.profile.FirstName, result.FirstName)
			}

			suite.mockDB.AssertExpectations(suite.T())
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// Benchmark tests
func BenchmarkCreateProfile(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	
	profile := &entity_profile.Profile{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
	}

	mockDB.On("Create", ctx, mock.Anything, "profiles").
		Return([]byte(`{"id":"123"}`), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.CreateProfile(ctx, profile)
	}
}

func BenchmarkGetProfileByID(b *testing.B) {
	mockDB := new(MockFirebaseDB)
	repo, _ := InicializeProfileRepository(mockDB)
	ctx := context.Background()
	
	profiles := []entity_profile.Profile{{ID: "123", FirstName: "John"}}
	id := "provider123"

	mockDB.On("GetByFilter", ctx, mock.Anything, "profiles").
		Return(mustMarshal(profiles), nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetProfileByID(ctx, &id)
	}
}