package service_finance

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	entity_finance "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/stretchr/testify/mock"
)

// MockIncomeRecordRepository is a mock type for the IncomeRecordRepositoryInterface
type MockIncomeRecordRepository struct {
	mock.Mock
}

func (m *MockIncomeRecordRepository) CreateIncomeRecord(ctx context.Context, data *entity_finance.IncomeRecord) (*entity_finance.IncomeRecord, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_finance.IncomeRecord), args.Error(1)
}

func (m *MockIncomeRecordRepository) GetIncomeRecordByID(ctx context.Context, id string) (*entity_finance.IncomeRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_finance.IncomeRecord), args.Error(1)
}

func (m *MockIncomeRecordRepository) GetIncomeRecords(ctx context.Context, params *entity_finance.GetIncomeRecordsQueryParameters) ([]entity_finance.IncomeRecord, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entity_finance.IncomeRecord), args.Error(1)
}

func (m *MockIncomeRecordRepository) UpdateIncomeRecord(ctx context.Context, id string, data *entity_finance.IncomeRecord) (*entity_finance.IncomeRecord, error) {
	args := m.Called(ctx, id, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity_finance.IncomeRecord), args.Error(1)
}

func (m *MockIncomeRecordRepository) DeleteIncomeRecord(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestIncomeRecordService_CreateIncomeRecord(t *testing.T) {
	mockRepo := new(MockIncomeRecordRepository)
	service, _ := InitializeIncomeRecordService(mockRepo)
	ctx := context.WithValue(context.Background(), "UserID", "user123")
	validDate := time.Now().Format("2006-01-02")
	one := 1

	tests := []struct {
		name          string
		inputRecord   *entity_finance.IncomeRecord
		setupMock     func(input *entity_finance.IncomeRecord)
		wantErrMsg    string
		wantRecord    *entity_finance.IncomeRecord // Only for non-recurring simple case
		timesRepoCalled int
	}{
		{
			name: "successful creation - non-recurring",
			inputRecord: &entity_finance.IncomeRecord{
				Category:      "salary",
				BankAccountID: "bank001",
				Amount:        2000,
				ReceiptDate:   validDate,
				// UserID will be set by service from context
			},
			setupMock: func(input *entity_finance.IncomeRecord) {
				// Clone and set UserID as service would
				expectedSave := *input
				expectedSave.UserID = "user123"
				// ID, CreatedAt, UpdatedAt will be set by repo/service
				mockRepo.On("CreateIncomeRecord", mock.Anything, mock.MatchedBy(func(r *entity_finance.IncomeRecord) bool {
					return r.Category == expectedSave.Category && r.UserID == expectedSave.UserID
				})).Return(&expectedSave, nil).Once()
			},
			wantErrMsg: "",
			wantRecord: &entity_finance.IncomeRecord{Category: "salary", UserID: "user123"}, // Simplified check
			timesRepoCalled: 1,
		},
		{
			name: "successful creation - recurring 3 times",
			inputRecord: &entity_finance.IncomeRecord{
				Category:        "rent_received",
				BankAccountID:   "bank002",
				Amount:          500,
				ReceiptDate:     validDate,
				IsRecurring:     true,
				RecurrenceCount: &one, // Set to 1 for simplicity, changed to 3 in test logic
			},
			setupMock: func(input *entity_finance.IncomeRecord) {
				// This setup needs to expect multiple calls for recurring
				// count := 3 // This is the actual recurrence count we'll test - set in test run

				// Adjust input for test logic
				// input.RecurrenceCount = &count // Set in test run logic

				// Expect 3 calls because the test logic will set RecurrenceCount to 3
				for i := 0; i < 3; i++ {
					mockRepo.On("CreateIncomeRecord", mock.Anything, mock.MatchedBy(func(r *entity_finance.IncomeRecord) bool {
						return r.Category == input.Category && r.UserID == "user123" && r.IsRecurring // Check a few key fields
					})).Return(func(ctx context.Context, rec *entity_finance.IncomeRecord) *entity_finance.IncomeRecord {
						// Return a copy of what was passed, simulating DB save
						saved := *rec
						saved.ID = "genID" + string(rune(i)) // Simulate generated ID
						return &saved
					}, nil).Once()
				}
			},
			wantErrMsg: "",
			// wantRecord check is harder for recurring, focus on no error and times called
			timesRepoCalled: 3,
		},
		{
			name: "validation error from record",
			inputRecord: &entity_finance.IncomeRecord{
				Category:      "", // Invalid
				BankAccountID: "bank001",
				Amount:        2000,
				ReceiptDate:   validDate,
			},
			setupMock: func(input *entity_finance.IncomeRecord) {
				// No call to repo expected
			},
			wantErrMsg: "validation failed: category is required",
			timesRepoCalled: 0,
		},
		{
			name: "repository error on create",
			inputRecord: &entity_finance.IncomeRecord{
				Category:      "salary",
				BankAccountID: "bank001",
				Amount:        2000,
				ReceiptDate:   validDate,
			},
			setupMock: func(input *entity_finance.IncomeRecord) {
				mockRepo.On("CreateIncomeRecord", mock.Anything, mock.AnythingOfType("*entity_finance.IncomeRecord")).Return(nil, errors.New("repo create error")).Once()
			},
			wantErrMsg: "repo create error",
			timesRepoCalled: 1,
		},
		{
            name: "missing userID in context",
            inputRecord: &entity_finance.IncomeRecord{
                Category:      "salary",
                BankAccountID: "bank001",
                Amount:        2000,
                ReceiptDate:   validDate,
            },
            setupMock: func(input *entity_finance.IncomeRecord) {},
            wantErrMsg: "userID not found in context",
            timesRepoCalled: 0,
            // Override context for this test
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.Mock.ExpectedCalls = nil // Reset mocks for each test run

			// Make a deep copy of inputRecord to avoid modification across tests, especially for RecurrenceCount
			currentInputRecord := *tt.inputRecord
			if tt.inputRecord.RecurrenceCount != nil {
				rc := *tt.inputRecord.RecurrenceCount
				currentInputRecord.RecurrenceCount = &rc
			}


			// Special handling for recurrence count in test logic for "recurring 3 times"
            if tt.name == "successful creation - recurring 3 times" {
                 count := 3
                 currentInputRecord.RecurrenceCount = &count // Ensure this is set before calling service
            }
			tt.setupMock(&currentInputRecord)


			currentCtx := ctx
            if tt.name == "missing userID in context" {
                currentCtx = context.Background() // Use context without UserID
            }



			_, err := service.CreateIncomeRecord(currentCtx, &currentInputRecord)

			if tt.wantErrMsg == "" {
				if err != nil {
					t.Errorf("IncomeRecordService.CreateIncomeRecord() error = %v, wantErr nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("IncomeRecordService.CreateIncomeRecord() error = nil, wantErrMsg %q", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IncomeRecordService.CreateIncomeRecord() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
				}
			}
			mockRepo.AssertNumberOfCalls(t, "CreateIncomeRecord", tt.timesRepoCalled)
		})
	}
}


func TestIncomeRecordService_GetIncomeRecordByID(t *testing.T) {
	mockRepo := new(MockIncomeRecordRepository)
	service, _ := InitializeIncomeRecordService(mockRepo)
	ctx := context.WithValue(context.Background(), "UserID", "user123")
	recordID := "recordXYZ"

	tests := []struct {
		name       string
		recordID   string
		setupMock  func()
		wantErrMsg string
		wantRecord *entity_finance.IncomeRecord
	}{
		{
			name:     "successful get",
			recordID: recordID,
			setupMock: func() {
				expected := &entity_finance.IncomeRecord{ID: recordID, UserID: "user123", Category: "salary"}
				mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(expected, nil).Once()
			},
			wantErrMsg: "",
			wantRecord: &entity_finance.IncomeRecord{ID: recordID, UserID: "user123", Category: "salary"},
		},
		{
			name:     "repository returns not found",
			recordID: recordID,
			setupMock: func() {
				mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(nil, errors.New("income record not found")).Once()
			},
			wantErrMsg: "income record not found",
		},
		{
			name:     "record belongs to another user",
			recordID: recordID,
			setupMock: func() {
				// Repo returns a record, but service should deny access
				returnedRecord := &entity_finance.IncomeRecord{ID: recordID, UserID: "user999", Category: "other"}
				mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(returnedRecord, nil).Once()
			},
			wantErrMsg: "income record not found or access denied",
		},
		{
            name: "missing userID in context",
            recordID: recordID,
            setupMock: func() {},
            wantErrMsg: "userID not found in context",
        },
        {
            name: "empty record ID",
            recordID: "",
            setupMock: func() {},
            wantErrMsg: "id is empty",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo.Mock.ExpectedCalls = nil
			tt.setupMock()

			currentCtx := ctx
            if tt.name == "missing userID in context" {
                currentCtx = context.Background()
            }


			gotRecord, err := service.GetIncomeRecordByID(currentCtx, tt.recordID)

			if tt.wantErrMsg == "" {
				if err != nil {
					t.Errorf("IncomeRecordService.GetIncomeRecordByID() error = %v, wantErr nil", err)
				}
				if !reflect.DeepEqual(gotRecord, tt.wantRecord) {
					t.Errorf("IncomeRecordService.GetIncomeRecordByID() got = %v, want %v", gotRecord, tt.wantRecord)
				}
			} else {
				if err == nil {
					t.Errorf("IncomeRecordService.GetIncomeRecordByID() error = nil, wantErrMsg %q", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IncomeRecordService.GetIncomeRecordByID() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}


func TestIncomeRecordService_GetIncomeRecords(t *testing.T) {
    mockRepo := new(MockIncomeRecordRepository)
    service, _ := InitializeIncomeRecordService(mockRepo)
    userID := "user123"
    ctx := context.WithValue(context.Background(), "UserID", userID)

    descFilter := "freelance work"
    startDateFilter := "2023-01-01"
    endDateFilter := "2023-01-31"
    sortKeyFilter := "amount"
    sortDirFilter := "desc"

    tests := []struct {
        name           string
        userIDParam    string // UserID passed as param to service method
        descParam      string
        startDateParam string
        endDateParam   string
        sortKeyParam   string
        sortDirParam   string
        setupMock      func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters)
        wantErrMsg     string
        wantRecords    []entity_finance.IncomeRecord
    }{
        {
            name:        "successful get - no filters",
            userIDParam: userID,
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                // UserID in expectedParams is set by service based on context
                mockRepo.On("GetIncomeRecords", mock.AnythingOfType("*context.valueCtx"), expectedParams).
                    Return([]entity_finance.IncomeRecord{{ID: "rec1", UserID: userID}}, nil).Once()
            },
            wantErrMsg:  "",
            wantRecords: []entity_finance.IncomeRecord{{ID: "rec1", UserID: userID}},
        },
        {
            name:           "successful get - with all filters",
            userIDParam:    userID,
            descParam:      descFilter,
            startDateParam: startDateFilter,
            endDateParam:   endDateFilter,
            sortKeyParam:   sortKeyFilter,
            sortDirParam:   sortDirFilter,
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                mockRepo.On("GetIncomeRecords", mock.AnythingOfType("*context.valueCtx"), expectedParams).
                    Return([]entity_finance.IncomeRecord{{ID: "rec2", UserID: userID, Description: &descFilter}}, nil).Once()
            },
            wantErrMsg:  "",
            wantRecords: []entity_finance.IncomeRecord{{ID: "rec2", UserID: userID, Description: &descFilter}},
        },
        {
            name:        "repository error",
            userIDParam: userID,
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                mockRepo.On("GetIncomeRecords", mock.AnythingOfType("*context.valueCtx"), expectedParams).Return(nil, errors.New("repo list error")).Once()
            },
            wantErrMsg: "repo list error",
        },
        {
            name:        "invalid sortDirection param",
            userIDParam: userID,
            sortDirParam: "sideways",
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                // No call to repo expected
            },
            wantErrMsg: "invalid sortDirection value",
        },
         {
            name:        "userID in context missing",
            userIDParam: userID, // This param won't matter if context is missing UserID
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                // No call to repo
            },
            wantErrMsg: "userID not found in context",
        },
        {
            name:        "userID param mismatch with context",
            userIDParam: "anotherUser", // Different from context UserID
            setupMock: func(expectedParams *entity_finance.GetIncomeRecordsQueryParameters) {
                // No call to repo
            },
            wantErrMsg: "user ID parameter mismatch",
        },

    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo.Mock.ExpectedCalls = nil

            currentCtx := ctx // Default context with UserID
            if tt.name == "userID in context missing" {
                currentCtx = context.Background() // Override for this specific test case
            }

            // Construct the GetIncomeRecordsQueryParameters that the service is expected to build and pass to the repo.
            // This is what we need to configure the mock with.
            expectedRepoParams := &entity_finance.GetIncomeRecordsQueryParameters{}
            // Service layer sets UserID from context, so we use the context's UserID for mock expectation
            if currentCtx.Value("UserID") != nil {
                 expectedRepoParams.UserID = currentCtx.Value("UserID").(string)
            }

            if tt.descParam != "" { expectedRepoParams.Description = &tt.descParam }
            if tt.startDateParam != "" { expectedRepoParams.StartDate = &tt.startDateParam }
            if tt.endDateParam != "" { expectedRepoParams.EndDate = &tt.endDateParam }
            if tt.sortKeyParam != "" { expectedRepoParams.SortKey = &tt.sortKeyParam }
            // Service validates sortDirection before creating queryParams for repo
            if tt.sortDirParam != "" && (strings.ToLower(tt.sortDirParam) == "asc" || strings.ToLower(tt.sortDirParam) == "desc") {
                dir := strings.ToLower(tt.sortDirParam)
                expectedRepoParams.SortDirection = &dir
            }


            tt.setupMock(expectedRepoParams)


            gotRecords, err := service.GetIncomeRecords(currentCtx, tt.userIDParam, tt.descParam, tt.startDateParam, tt.endDateParam, tt.sortKeyParam, tt.sortDirParam)

            if tt.wantErrMsg == "" {
                if err != nil {
                    t.Errorf("IncomeRecordService.GetIncomeRecords() error = %v, wantErr nil", err)
                }
                if !reflect.DeepEqual(gotRecords, tt.wantRecords) {
                    t.Errorf("IncomeRecordService.GetIncomeRecords() got = %v, want %v", gotRecords, tt.wantRecords)
                }
            } else {
                if err == nil {
                    t.Errorf("IncomeRecordService.GetIncomeRecords() error = nil, wantErrMsg %q", tt.wantErrMsg)
                } else if !strings.Contains(err.Error(), tt.wantErrMsg) {
                    t.Errorf("IncomeRecordService.GetIncomeRecords() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
                }
            }
        })
    }
}


func TestIncomeRecordService_UpdateIncomeRecord(t *testing.T) {
    mockRepo := new(MockIncomeRecordRepository)
    service, _ := InitializeIncomeRecordService(mockRepo)
    userID := "user123"
    ctx := context.WithValue(context.Background(), "UserID", userID)
    recordID := "recordToUpdate"
    validDate := time.Now().Format("2006-01-02")

    updateData := &entity_finance.IncomeRecord{
        Category:      "freelance",
        BankAccountID: "bank789",
        Amount:        1250,
        ReceiptDate:   validDate,
        // UserID will be set by service
    }

    existingRecord := &entity_finance.IncomeRecord{
        ID:            recordID,
        UserID:        userID,
        Category:      "salary",
        BankAccountID: "bankOld",
        Amount:        1000,
        ReceiptDate:   "2023-01-01",
        CreatedAt:     time.Now().Add(-24 * time.Hour),
    }

    tests := []struct {
        name        string
        recordID    string
        updateData  *entity_finance.IncomeRecord
        setupMock   func(dataToUpdate *entity_finance.IncomeRecord)
        wantErrMsg  string
        wantRecord  *entity_finance.IncomeRecord // Simplified check on updated fields
    }{
        {
            name:       "successful update",
            recordID:   recordID,
            updateData: updateData,
            setupMock: func(dataToUpdate *entity_finance.IncomeRecord) {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(existingRecord, nil).Once()

                expectedRepoUpdate := *dataToUpdate
                expectedRepoUpdate.UserID = userID
                expectedRepoUpdate.ID = recordID
                expectedRepoUpdate.CreatedAt = existingRecord.CreatedAt

                mockRepo.On("UpdateIncomeRecord", mock.AnythingOfType("*context.valueCtx"), recordID,
                    mock.MatchedBy(func(r *entity_finance.IncomeRecord) bool {
                        return r.ID == recordID && r.UserID == userID && r.Category == expectedRepoUpdate.Category && r.Amount == expectedRepoUpdate.Amount && r.CreatedAt == expectedRepoUpdate.CreatedAt
                    })).Return(&expectedRepoUpdate, nil).Once()
            },
            wantErrMsg: "",
            wantRecord: &entity_finance.IncomeRecord{Category: "freelance", Amount: 1250},
        },
        {
            name:       "update with validation error",
            recordID:   recordID,
            updateData: &entity_finance.IncomeRecord{Category: ""},
             setupMock: func(dataToUpdate *entity_finance.IncomeRecord) {

            },
            wantErrMsg: "validation failed: category is required",
        },
        {
            name:       "repo GetIncomeRecordByID fails (not found)",
            recordID:   recordID,
            updateData: updateData,
            setupMock: func(dataToUpdate *entity_finance.IncomeRecord) {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(nil, errors.New("record not found for update")).Once()
            },
            wantErrMsg: "record not found for update",
        },
        {
            name:       "user unauthorized (record belongs to another user)",
            recordID:   recordID,
            updateData: updateData,
            setupMock: func(dataToUpdate *entity_finance.IncomeRecord) {
                anotherUserRecord := *existingRecord
                anotherUserRecord.UserID = "user999"
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(&anotherUserRecord, nil).Once()
            },
            wantErrMsg: "income record not found or access denied for update",
        },
        {
            name:       "repo UpdateIncomeRecord fails",
            recordID:   recordID,
            updateData: updateData,
            setupMock: func(dataToUpdate *entity_finance.IncomeRecord) {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(existingRecord, nil).Once()
                mockRepo.On("UpdateIncomeRecord", mock.AnythingOfType("*context.valueCtx"), recordID, mock.AnythingOfType("*entity_finance.IncomeRecord")).Return(nil, errors.New("repo update error")).Once()
            },
            wantErrMsg: "repo update error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo.Mock.ExpectedCalls = nil

            // Deep copy updateData for this test run
            currentUpdateData := *tt.updateData
            if tt.updateData.Description != nil {
                desc := *tt.updateData.Description
                currentUpdateData.Description = &desc
            }
             if tt.updateData.RecurrenceCount != nil {
                rc := *tt.updateData.RecurrenceCount
                currentUpdateData.RecurrenceCount = &rc
            }
            if tt.updateData.Observations != nil {
                obs := *tt.updateData.Observations
                currentUpdateData.Observations = &obs
            }


            tt.setupMock(&currentUpdateData)

            gotRecord, err := service.UpdateIncomeRecord(ctx, tt.recordID, &currentUpdateData)

            if tt.wantErrMsg == "" {
                if err != nil {
                    t.Errorf("IncomeRecordService.UpdateIncomeRecord() error = %v, wantErr nil", err)
                }
                if tt.wantRecord != nil {
                    if gotRecord.Category != tt.wantRecord.Category || gotRecord.Amount != tt.wantRecord.Amount {
                         t.Errorf("IncomeRecordService.UpdateIncomeRecord() got simplified %v / %v, want %v / %v", gotRecord.Category, gotRecord.Amount, tt.wantRecord.Category, tt.wantRecord.Amount)
                    }
                }
            } else {
                if err == nil {
                    t.Errorf("IncomeRecordService.UpdateIncomeRecord() error = nil, wantErrMsg %q", tt.wantErrMsg)
                } else if !strings.Contains(err.Error(), tt.wantErrMsg) {
                    t.Errorf("IncomeRecordService.UpdateIncomeRecord() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
                }
            }
        })
    }
}


func TestIncomeRecordService_DeleteIncomeRecord(t *testing.T) {
    mockRepo := new(MockIncomeRecordRepository)
    service, _ := InitializeIncomeRecordService(mockRepo)
    userID := "user123"
    ctx := context.WithValue(context.Background(), "UserID", userID)
    recordID := "recordToDelete"

    existingRecordOwnedByUser := &entity_finance.IncomeRecord{ID: recordID, UserID: userID}
    existingRecordOtherUser := &entity_finance.IncomeRecord{ID: recordID, UserID: "user999"}

    tests := []struct {
        name       string
        recordID   string
        setupMock  func()
        wantErrMsg string
    }{
        {
            name:     "successful delete",
            recordID: recordID,
            setupMock: func() {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(existingRecordOwnedByUser, nil).Once()
                mockRepo.On("DeleteIncomeRecord", mock.AnythingOfType("*context.valueCtx"), recordID).Return(nil).Once()
            },
            wantErrMsg: "",
        },
        {
            name:     "repo GetIncomeRecordByID fails (not found)",
            recordID: recordID,
            setupMock: func() {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(nil, errors.New("record not found for delete verification")).Once()
            },
            wantErrMsg: "record not found for delete verification",
        },
        {
            name:     "user unauthorized (record belongs to another user)",
            recordID: recordID,
            setupMock: func() {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(existingRecordOtherUser, nil).Once()
            },
            wantErrMsg: "income record not found or access denied for delete",
        },
        {
            name:     "repo DeleteIncomeRecord fails",
            recordID: recordID,
            setupMock: func() {
                mockRepo.On("GetIncomeRecordByID", mock.AnythingOfType("*context.valueCtx"), recordID).Return(existingRecordOwnedByUser, nil).Once()
                mockRepo.On("DeleteIncomeRecord", mock.AnythingOfType("*context.valueCtx"), recordID).Return(errors.New("repo delete error")).Once()
            },
            wantErrMsg: "repo delete error",
        },
        {
            name: "empty record ID",
            recordID: "",
            setupMock: func() {},
            wantErrMsg: "id is empty",
        },
        {
            name: "missing userID in context",
            recordID: recordID,
            setupMock: func() {},
            wantErrMsg: "userID not found in context",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo.Mock.ExpectedCalls = nil
            tt.setupMock()

            currentCtx := ctx
            if tt.name == "missing userID in context" {
                currentCtx = context.Background()
            }

            err := service.DeleteIncomeRecord(currentCtx, tt.recordID)

            if tt.wantErrMsg == "" {
                if err != nil {
                    t.Errorf("IncomeRecordService.DeleteIncomeRecord() error = %v, wantErr nil", err)
                }
            } else {
                if err == nil {
					t.Errorf("IncomeRecordService.DeleteIncomeRecord() error = nil, wantErrMsg %q", tt.wantErrMsg)
				} else if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("IncomeRecordService.DeleteIncomeRecord() error = %q, wantErrMsg %q", err.Error(), tt.wantErrMsg)
				}
            }
        })
    }
}
