// Package mocks provides mock implementations for cache interfaces.
package mocks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Tomelin/dashfin-backend-app/pkg/cache"
	"github.com/stretchr/testify/mock"
)

// MockRedisCache is a mock implementation of cache.RedisCacheInterface.
type MockRedisCache struct {
	mock.Mock
}

// Get mocks the Get method.
// It simulates unmarshaling by marshaling the pre-configured mock return value (if it's not an error)
// and then unmarshaling into the 'data' argument.
func (m *MockRedisCache) Get(ctx context.Context, key string, data interface{}) error {
	args := m.Called(ctx, key, data)
	err := args.Error(0)

	if err == nil {
		// If no error, 'data' should be populated.
		// The mock should be configured with the actual data structure to return.
		// args.Get(1) would be the pre-configured data for successful unmarshal.
		// This is a bit tricky with testify/mock for unmarshaling into an argument.
		// A common way is to have the mock return the data directly, and the test does the unmarshal,
		// or the mock itself handles the unmarshal if 'data' is passed to 'Return'.
		// For simplicity, let's assume the mock is configured to return a specific error or nil.
		// If nil, the test setup needs to ensure 'data' is somehow populated or checked.
		// A better way for 'Get' mocks is to allow configuring the output for 'data'.
		//
		// Let's try to simulate the behavior of the actual Get:
		// The mock is configured with the object that should be "retrieved" from cache.
		// e.g., mockCache.On("Get", ..., mock.Anything).Return(nil, expectedDataOutput)
		// However, standard mock.Return only returns values, it doesn't modify arguments by pointer.
		//
		// Alternative: The mock call can configure what `data` should be populated with.
		// We can provide a function in Return that populates `data`.
		// For example:
		// mockCache.On("Get", mock.Anything, mock.AnythingOfType("*entity_dashboard.DashboardSummary")).
		//   Run(func(args mock.Arguments) {
		//		    arg := args.Get(2).(*entity_dashboard.DashboardSummary)
		//		    *arg = yourExpectedCachedSummary
		//   }).
		//   Return(nil)

		// For this implementation, we'll rely on the test to set up the `data` argument
		// if the mock is to simulate a cache hit by returning specific data.
		// The test can do this:
		// `mockRedis.On("Get", mock.Anything, mock.AnythingOfType("*types.MyType")).Run(func(args mock.Arguments) { ... }).Return(nil)`
		// The `Run` function can then populate `args.Get(2)`.
		//
		// If the error is cache.ErrCacheMiss or other, `data` is not touched.
		// If the error is nil, it implies data was found and unmarshalled.
		// The test case will need to handle how `data` gets its value for a cache hit scenario.
		// One way is for the mock to store the expected data and the test to check it.
		// Here, we assume that if err is nil, the test setup has arranged for `data` to be filled.
		// This often involves using `mock.AnythingOfType` for the `data` argument in `On`
		// and then using `Run` to populate it.
		//
		// Let's simplify: the mock stores the object it's supposed to "return" via unmarshalling.
		// This means the mock needs to be aware of the type.
		// This is what testify does if you pass the object to Return for the argument.
		// However, `data` is `interface{}`.
		//
		// A more robust way for this mock:
		retVal := args.Get(1) // This would be the object to unmarshal into `data`
		if err == nil && retVal != nil {
			sourceBytes, marshalErr := json.Marshal(retVal)
			if marshalErr != nil {
				return marshalErr // Error during mock's simulation of unmarshal
			}
			return json.Unmarshal(sourceBytes, data)
		}
	}
	return err
}

// Set mocks the Set method.
func (m *MockRedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

// Delete mocks the Delete method.
func (m *MockRedisCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

// Helper for tests to configure the Get mock for successful cache hit.
// It configures the mock to populate the 'data' argument with 'cachedItem'.
func (m *MockRedisCache) OnGetSuccess(ctx context.Context, key string, dataArgMatcher interface{}, cachedItem interface{}) *mock.Call {
    return m.On("Get", ctx, key, dataArgMatcher).Run(func(args mock.Arguments) {
        out := args.Get(2) // This is the 'data interface{}' argument
        bytes, _ := json.Marshal(cachedItem)
        json.Unmarshal(bytes, out)
    }).Return(nil)
}

// Helper for tests to configure Get for cache miss
func (m *MockRedisCache) OnGetCacheMiss(ctx context.Context, key string, dataArgMatcher interface{}) *mock.Call {
    return m.On("Get", ctx, key, dataArgMatcher).Return(cache.ErrCacheMiss)
}

// Helper for tests to configure Get for other errors
func (m *MockRedisCache) OnGetError(ctx context.Context, key string, dataArgMatcher interface{}, err error) *mock.Call {
    return m.On("Get", ctx, key, dataArgMatcher).Return(err)
}

// Helper for tests to configure Set
func (m *MockRedisCache) OnSet(ctx context.Context, key string, value interface{}, ttl time.Duration, err error) *mock.Call {
	return m.On("Set", ctx, key, value, ttl).Return(err)
}
