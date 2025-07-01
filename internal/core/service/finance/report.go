package finance

import (
	"context"
	"fmt"
	"log"
	"time"

	entity "github.com/Tomelin/dashfin-backend-app/internal/core/entity/finance"
	"github.com/Tomelin/dashfin-backend-app/pkg/cache"
	"github.com/Tomelin/dashfin-backend-app/pkg/message_queue"
	"github.com/Tomelin/dashfin-backend-app/pkg/utils"
)

type FinancialReportDataService struct {
	// repo         entity.FinancialReportDataRepositoryInterface
	income       entity.IncomeRecordServiceInterface
	expense      entity.ExpenseRecordServiceInterface
	cache        cache.CacheService
	messageQueue message_queue.MessageQueue
}

func InitializeFinancialReportDataService(
	// repo entity.FinancialReportDataRepositoryInterface,
	income entity.IncomeRecordServiceInterface,
	expense entity.ExpenseRecordServiceInterface,
	cacheService cache.CacheService,
	messageQueue message_queue.MessageQueue,
) (entity.FinancialReportDataServiceInterface, error) {

	if income == nil {
		return nil, fmt.Errorf("income cannot be nil")
	}

	if expense == nil {
		return nil, fmt.Errorf("expense cannot be nil")
	}

	if cacheService == nil {
		return nil, fmt.Errorf("cacheService cannot be nil")
	}

	if messageQueue == nil {
		return nil, fmt.Errorf("message queue cannot be nil")
	}

	report := FinancialReportDataService{
		income:       income,
		expense:      expense,
		cache:        cacheService,
		messageQueue: messageQueue,
	}

	go report.mqConsumer(context.Background())

	return &report, nil
}

func (s *FinancialReportDataService) GetFinancialReportData(ctx context.Context) (*entity.FinancialReportData, error) {
	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	if userId == nil || *userId == "" {
		return nil, fmt.Errorf("userId is nil")
	}

	report := entity.FinancialReportData{
		MonthlyCashFlow: s.CalculateMonthlyCashFlow(ctx),
		ExpenseByCategory: []entity.CategoryExpenseItem{
			entity.CategoryExpenseItem{
				Name:  "Moradia",
				Value: 99.99,
			},
			entity.CategoryExpenseItem{
				Name:  "Transporte",
				Value: 99.99,
			},
			entity.CategoryExpenseItem{
				Name:  "Saúde",
				Value: 199.99,
			},
		},
		ExpenseByCategoryLast12Months: []entity.CategoryExpenseItem{
			entity.CategoryExpenseItem{
				Name:  "Moradia",
				Value: 99.99,
			},
			entity.CategoryExpenseItem{
				Name:  "Transporte",
				Value: 99.99,
			},
			entity.CategoryExpenseItem{
				Name:  "Saúde",
				Value: 79.99,
			},
		},
		NetWorthEvolution: []entity.NetWorthHistoryItem{
			entity.NetWorthHistoryItem{
				Date:  "2025-01",
				Value: 99.99,
			},
			entity.NetWorthHistoryItem{
				Date:  "2025-02",
				Value: 99.99,
			},
			entity.NetWorthHistoryItem{
				Date:  "2025-03",
				Value: 199.99,
			},
			entity.NetWorthHistoryItem{
				Date:  "2025-05",
				Value: 59.99,
			},
		},
		ExpenseBreakdown: []entity.ExpenseCategoryWithSubItems{
			entity.ExpenseCategoryWithSubItems{
				Name:  "Moradia",
				Value: 99.99,
				Children: []entity.ExpenseSubCategoryItem{
					entity.ExpenseSubCategoryItem{
						Name:  "Aluguel",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Condomínio",
						Value: 59.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Luz",
						Value: 0.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Água",
						Value: 1169.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Gás",
						Value: 1589.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Outros",
						Value: 99.99,
					},
				},
			},
			entity.ExpenseCategoryWithSubItems{
				Name:  "Transporte",
				Value: 99.99,
				Children: []entity.ExpenseSubCategoryItem{
					entity.ExpenseSubCategoryItem{
						Name:  "Combustível",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Manutenção",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Uber",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Ônibus",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Outros",
						Value: 99.99,
					},
				},
			},

			entity.ExpenseCategoryWithSubItems{
				Name:  "Saúde",
				Value: 99.99,
				Children: []entity.ExpenseSubCategoryItem{
					entity.ExpenseSubCategoryItem{
						Name:  "Dentista",
						Value: 99.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Farmácia",
						Value: 299.99,
					},
				},
			},
			entity.ExpenseCategoryWithSubItems{
				Name:  "Educação",
				Value: 99.99,
				Children: []entity.ExpenseSubCategoryItem{
					entity.ExpenseSubCategoryItem{
						Name:  "Udemy",
						Value: 139.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Curso AI",
						Value: 299.99,
					},
				},
			},
		},
	}

	log.Println("################ CalculateMonthlyCashFlow ###############")
	log.Println(s.CalculateMonthlyCashFlow(ctx))

	// cacheFlowPct represents the percentage change in cash flow for the current month compared to the previous month's cash flow.
	log.Println("Report > ", report)
	log.Println("SummaryCards > ", report.SummaryCards)
	log.Println(fmt.Sprintf("SummaryCards: %s", report.SummaryCards))
	log.Println("CurrentMonthCashFlow > ", report.SummaryCards.CurrentMonthCashFlow)
	log.Println("CurrentMonthCashFlowChangePct > ", report.SummaryCards.CurrentMonthCashFlowChangePct)
	log.Println("NetWorth > ", report.SummaryCards.NetWorth)
	log.Println("NetWorthChangePercent > ", report.SummaryCards.NetWorthChangePercent)

	return &report, nil
}

func (s *FinancialReportDataService) getIncomeRecords(ctx context.Context, startDate, endDate string) ([]entity.IncomeRecord, float64, error) {

	var report []entity.IncomeRecord
	var amount float64
	report, err := s.income.GetIncomeRecords(ctx, &entity.GetIncomeRecordsQueryParameters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		return nil, amount, err
	}

	log.Printf(">> startDate %s, amount %v itens %v", startDate, amount, len(report))
	count := 0
	for _, v := range report {
		amount += v.Amount
		count += 1
		log.Printf("[IncomeRecords] > month %v item %v amount %v total %v ", v.ReceiptDate, count, v.Amount, amount)
	}
	log.Printf(">> total amount  %v", amount)
	return report, amount, nil
}

func (s *FinancialReportDataService) getExpenseRecords(ctx context.Context, startDate, endDate, cacheKey string) ([]entity.ExpenseRecord, float64, error) {

	var report []entity.ExpenseRecord
	var amount float64

	report, err := s.expense.GetExpenseRecordsByDate(ctx, &entity.ExpenseRecordQueryByDate{
		StartDate: startDate,
		EndDate:   endDate,
	})

	if err != nil {
		log.Println("error getting expense records", err)
		return nil, amount, err
	}

	for _, v := range report {
		amount += v.Amount
	}

	return report, amount, nil
}

func (s *FinancialReportDataService) CalculateMonthlyCashFlow(ctx context.Context) []entity.MonthlySummaryItem {

	dataPoints := make(map[string]*entity.MonthlySummaryItem)
	// Convert map to slice and sort by month
	var monthlySummary []entity.MonthlySummaryItem

	// Get records for the last 12 months
	now := time.Now()
	for i := 0; i < 12; i++ {
		month := now.AddDate(0, -i, 0)
		firstDayOfMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location()).Format("2006-01-02")
		lastDayOfMonth := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, month.Location()).Format("2006-01-02")
		monthYearFormat := month.Format("2006-01")

		_, incomeAmount, err := s.getIncomeRecords(ctx, firstDayOfMonth, lastDayOfMonth)
		if err != nil {
			log.Printf("Error getting income records for %s: %v", monthYearFormat, err)
			continue
		}

		_, expenseAmount, err := s.getExpenseRecords(ctx, firstDayOfMonth, lastDayOfMonth, cacheKeyExpenseReportByLastMonth)
		if err != nil {
			log.Printf("Error getting expense records for %s: %v", monthYearFormat, err)
			continue
		}

		dataPoints[monthYearFormat] = &entity.MonthlySummaryItem{
			Month:    monthYearFormat,
			Revenue:  incomeAmount,
			Expenses: expenseAmount,
		}

		monthlySummary = append(monthlySummary, entity.MonthlySummaryItem{
			Month:    monthYearFormat,
			Revenue:  0,
			Expenses: 123,
		})
		log.Printf("[MONTH] start %v end %v income %v expense %v", firstDayOfMonth, lastDayOfMonth, incomeAmount, expenseAmount)

		// Although we fetched records, we only need the total amount for this summary item.
		// The individual records are not stored in the MonthlySummaryItem.
		incomeAmount = 0
		expenseAmount = 0
	}

	return monthlySummary
}

func (s *FinancialReportDataService) setCache(ctx context.Context, cacheKey string, planData *entity.SpendingPlan) {

	s.cache.Set(ctx, cacheKey, *planData, serviceCacheTTL)
}

func (s *FinancialReportDataService) mqConsumer(ctx context.Context) {

	// s.messageQueue.Consumer(ctx, mq_exchange, mq_queue_income)
}

func (s *FinancialReportDataService) processMQConsumer(body []byte, traceID string) error {

	return nil
}
