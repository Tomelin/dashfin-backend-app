package finance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

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
	_, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	report := entity.FinancialReportData{
		SummaryCards: entity.ReportSummaryCards{
			CurrentMonthCashFlow:          99.99,
			CurrentMonthCashFlowChangePct: 99.99,
			NetWorth:                      99.99,
			NetWorthChangePercent:         99.99,
		},
		MonthlyCashFlow: []entity.MonthlySummaryItem{
			entity.MonthlySummaryItem{
				Month:    "2024-12",
				Revenue:  99.99,
				Expenses: 99.99,
			},
			entity.MonthlySummaryItem{
				Month:    "2025-01",
				Revenue:  199.99,
				Expenses: 99.99,
			},
			entity.MonthlySummaryItem{
				Month:    "2025-04",
				Revenue:  2199.99,
				Expenses: 199.99,
			},
			entity.MonthlySummaryItem{
				Month:    "2025-05",
				Revenue:  12199.99,
				Expenses: 2199.99,
			},
		},
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
						Value: 11169.99,
					},
					entity.ExpenseSubCategoryItem{
						Name:  "Gás",
						Value: 15189.99,
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

	incomeReportMonth, incomeAmountMonth, err := s.getMonthIncomeRecords(ctx, firstDayOfMonth, lastDayOfMonth, cacheKeyIncomeReportByMonth)
	if err != nil {
		return nil, err
	}

	incomeReportYear, incomeAmountYear, err := s.getMonthIncomeRecords(ctx, firstDayOfYear, lastDayOfMonth, cacheKeyIncomeReportByYear)
	if err != nil {
		return nil, err
	}

	incomeReportLastMonth, incomeAmountLastMonth, err := s.getMonthIncomeRecords(ctx, firstDayOfLastMonth, lastDayOfLastMonth, cacheKeyIncomeReportByLastMonth)
	if err != nil {
		return nil, err
	}

	expenseReportMonth, expenseAmountMonth, err := s.getExpenseRecords(ctx, firstDayOfMonth, lastDayOfMonth, cacheKeyExpenseReportByMonth)
	if err != nil {
		return nil, err
	}

	expenseReportYear, expenseAmountYear, err := s.getExpenseRecords(ctx, firstDayOfYear, lastDayOfMonth, cacheKeyExpenseReportByYear)
	if err != nil {
		return nil, err
	}

	expenseReportLastMonth, expenseAmountLastMonth, err := s.getExpenseRecords(ctx, firstDayOfLastMonth, lastDayOfLastMonth, cacheKeyExpenseReportByLastMonth)
	if err != nil {
		return nil, err
	}

	log.Println(">>>>> <<<<<")
	log.Println(incomeReportMonth, incomeReportYear, expenseReportMonth, expenseReportYear, incomeAmountYear, expenseAmountYear, incomeAmountMonth, incomeAmountLastMonth, incomeReportLastMonth, expenseReportLastMonth, expenseAmountLastMonth)
	log.Println(">>>>> <<<<<")

	// cacheFlowPct represents the percentage change in cash flow for the current month compared to the previous month's cash flow.
	lastMonthCashFlow := incomeAmountLastMonth - expenseAmountLastMonth
	currentMonthCashFlow := incomeAmountMonth - expenseAmountMonth
	last12Months := incomeAmountYear - expenseAmountYear
	var cacheFlowPct float64
	if lastMonthCashFlow != 0 {
		cacheFlowPct = (currentMonthCashFlow - lastMonthCashFlow) / lastMonthCashFlow * 100
	}

	report.SummaryCards.CurrentMonthCashFlow = currentMonthCashFlow
	report.SummaryCards.CurrentMonthCashFlowChangePct = cacheFlowPct

	// Calculate the Net Worth change percentage over the last 12 months.
	var netWorthChangePct float64
	if last12Months != 0 { // Avoid division by zero
		netWorthChangePct = ((currentMonthCashFlow - last12Months) / last12Months) * 100
	}
	report.SummaryCards.NetWorth = last12Months
	report.SummaryCards.NetWorthChangePercent = netWorthChangePct

	log.Println("Report > ", report)
	log.Println("SummaryCards > ", report.SummaryCards)
	log.Println(fmt.Sprintf("SummaryCards: %s", report.SummaryCards))
	log.Println("CurrentMonthCashFlow > ", report.SummaryCards.CurrentMonthCashFlow)
	log.Println("CurrentMonthCashFlowChangePct > ", report.SummaryCards.CurrentMonthCashFlowChangePct)
	log.Println("NetWorth > ", report.SummaryCards.NetWorth)
	log.Println("NetWorthChangePercent > ", report.SummaryCards.NetWorthChangePercent)

	return &report, nil
}

func (s *FinancialReportDataService) getMonthIncomeRecords(ctx context.Context, startDate, endDate, cacheKey string) ([]entity.IncomeRecord, float64, error) {
	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, 0, err
	}

	if userId == nil || *userId == "" {
		return nil, 0, fmt.Errorf("userId is nil")
	}

	cachedData, err := s.cache.Get(ctx, fmt.Sprintf("%s_%s", cacheKey, *userId))
	var report []entity.IncomeRecord
	var amount float64
	if err == nil { // Found in cache
		if jsonErr := json.Unmarshal([]byte(cachedData), &report); jsonErr == nil {
			return nil, amount, nil
		}
		if len(report) > 0 {
			for _, v := range report {
				amount += v.Amount
			}
			return report, amount, nil
		}
	}

	report, err = s.income.GetIncomeRecords(ctx, &entity.GetIncomeRecordsQueryParameters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})

	if err != nil {
		return nil, amount, err
	}

	if len(report) > 0 {
		cacheData, _ := json.Marshal(report)
		s.cache.Set(ctx, fmt.Sprintf("%s_%s", cacheKey, *userId), cacheData, serviceCacheTTL)
		for _, v := range report {
			amount += v.Amount
		}
		return report, amount, nil
	}

	return report, amount, nil
}

func (s *FinancialReportDataService) getExpenseRecords(ctx context.Context, startDate, endDate, cacheKey string) ([]entity.ExpenseRecord, float64, error) {

	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, 0, err
	}

	if userId == nil || *userId == "" {
		return nil, 0, fmt.Errorf("userId is nil")
	}

	cachedData, err := s.cache.Get(ctx, fmt.Sprintf("%s_%s", cacheKey, *userId))
	var report []entity.ExpenseRecord
	var amount float64
	if err == nil { // Found in cache
		if jsonErr := json.Unmarshal([]byte(cachedData), &report); jsonErr == nil {
			return nil, amount, nil
		}
		if len(report) > 0 {
			for _, v := range report {
				amount += v.Amount
			}
			return report, amount, nil
		}
	}
	report, err = s.expense.GetExpenseRecordsByDate(ctx, &entity.ExpenseRecordQueryByDate{
		StartDate: startDate,
		EndDate:   endDate,
	})

	if err != nil {
		log.Println("error getting expense records", err)
		return nil, amount, err
	}

	if len(report) > 0 {
		cacheData, _ := json.Marshal(report)
		s.cache.Set(ctx, fmt.Sprintf("%s_%s", cacheKey, *userId), cacheData, serviceCacheTTL)
		for _, v := range report {
			amount += v.Amount
		}
		return report, amount, nil
	}

	return report, amount, nil
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
