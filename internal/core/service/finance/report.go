package finance

import (
	"context"
	"encoding/json"
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
	income          entity.IncomeRecordServiceInterface
	expense         entity.ExpenseRecordServiceInterface
	cache           cache.CacheService
	messageQueue    message_queue.MessageQueue
	incomeRecords   []entity.IncomeRecord
	expenseRecords  []entity.ExpenseRecord
	financialReport *entity.FinancialReportData
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

	// Initialize financialReport if it is nil
	if s.financialReport == nil {
		s.financialReport = &entity.FinancialReportData{}
	}

	// Initialize the date variables for the report
	s.getIncomeRecords(ctx)
	s.getExpenseRecords(ctx)

	s.getSummaryCards(ctx)
	log.Println("FinancialReportDataService: Summary cards fetched", s.financialReport.SummaryCards)
	s.getMonthlyCashFlow(ctx)
	log.Println("FinancialReportDataService: Monthly cash flow fetched", s.financialReport.MonthlyCashFlow)
	s.getExpenseByCategory(ctx)
	log.Println("FinancialReportDataService: Expense by category fetched", s.financialReport.ExpenseByCategoryLast12Months)
	s.getExpenseByCategoryLast12Months(ctx)
	log.Println("FinancialReportDataService: Expense by category last 12 months fetched", s.financialReport.ExpenseByCategoryLast12Months)

	if s.financialReport != nil {
		s.cache.Set(ctx, cacheKeyFinancialReport, *s.financialReport, serviceCacheTTL)
	}

	return s.financialReport, nil
}

func (s *FinancialReportDataService) getIncomeRecords(ctx context.Context) error {

	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return err
	}

	if userId == nil || *userId == "" {
		return fmt.Errorf("userId is nil")
	}

	cachedData, err := s.cache.Get(ctx, fmt.Sprintf("%s_%s", cacheKeyIncomeReport, *userId))
	var report []entity.IncomeRecord
	if err == nil { // Found in cache
		if jsonErr := json.Unmarshal([]byte(cachedData), &report); jsonErr == nil {
			if len(report) > 0 {
				s.incomeRecords = report
				log.Println("FinancialReportDataService: Income records found in cache", s.incomeRecords)
			}
			return nil
		}
	}

	report, err = s.income.GetIncomeRecords(ctx, &entity.GetIncomeRecordsQueryParameters{
		StartDate: nil,
		EndDate:   nil,
	})

	if err != nil {
		return err
	}

	if len(report) > 0 {
		cacheData, _ := json.Marshal(report)
		s.cache.Set(ctx, fmt.Sprintf("%s_%s", cacheKeyIncomeReport, *userId), cacheData, serviceCacheTTL)
		s.incomeRecords = report
		log.Println("FinancialReportDataService: Income records fetched from service", s.incomeRecords)
		return nil
	}

	return nil
}

func (s *FinancialReportDataService) getIncomeRecordsByPeriod(ctx context.Context, startDate, endDate time.Time) (report []entity.IncomeRecord, amount float64, err error) {

	if time.Time.Equal(time.Time{}, startDate) || time.Time.Equal(time.Time{}, endDate) {
		return nil, 0, fmt.Errorf("startDate or endDate is empty")
	}

	if len(s.incomeRecords) == 0 {
		if err := s.getIncomeRecords(ctx); err != nil {
			return nil, 0, err
		}
	}

	for _, record := range s.incomeRecords {
		if record.ReceiptDate.After(startDate) && record.ReceiptDate.Before(endDate) {
			report = append(report, record)
			amount += record.Amount
		}
	}

	return report, amount, nil
}

func (s *FinancialReportDataService) getExpenseRecords(ctx context.Context) error {

	userId, err := utils.GetUserID(ctx)
	if err != nil {
		return err
	}

	if userId == nil || *userId == "" {
		return fmt.Errorf("userId is nil")
	}

	cachedData, err := s.cache.Get(ctx, fmt.Sprintf("%s_%s", cacheKeyExpenseReport, *userId))
	var report []entity.ExpenseRecord
	if err == nil { // Found in cache
		if jsonErr := json.Unmarshal([]byte(cachedData), &report); jsonErr == nil {
			if len(report) > 0 {
				s.expenseRecords = report
			}
			return nil
		}
	}

	report, err = s.expense.GetExpenseRecords(ctx)
	if err != nil {
		return err
	}

	if len(report) > 0 {
		cacheData, _ := json.Marshal(report)
		s.cache.Set(ctx, fmt.Sprintf("%s_%s", cacheKeyExpenseReport, *userId), cacheData, serviceCacheTTL)
		s.expenseRecords = report
		return nil
	}

	return nil
}

func (s *FinancialReportDataService) getExpenseRecordsByPeriod(ctx context.Context, startDate, endDate time.Time) (report []entity.ExpenseRecord, amount float64, err error) {

	if startDate.IsZero() || endDate.IsZero() {
		return nil, 0, fmt.Errorf("startDate or endDate is empty")
	}

	if len(s.expenseRecords) == 0 {
		if err := s.getExpenseRecords(ctx); err != nil {
			return nil, 0, err
		}
	}

	for _, record := range s.expenseRecords {
		report = append(report, record)
		amount += record.Amount
	}

	return report, amount, nil
}

func (s *FinancialReportDataService) getSummaryCards(ctx context.Context) error {

	// Get primeiro dia do mês atual
	firstDayOfMonth := utils.GetFirstDayOfCurrentMonth()

	// Get último dia do mês atual
	lastDayOfMonth := utils.GetLastDayOfCurrentMonth()

	_, incomeAmount, err := s.getIncomeRecordsByPeriod(ctx, firstDayOfMonth, lastDayOfMonth)
	if err != nil {
		return err
	}

	_, expenseAmount, err := s.getExpenseRecordsByPeriod(ctx, firstDayOfMonth, lastDayOfMonth)
	if err != nil {
		return err
	}

	log.Println("FinancialReportDataService: Income amount for current month:", incomeAmount)
	log.Println("FinancialReportDataService: Expense amount for current month:", expenseAmount)
	s.financialReport.SummaryCards.CurrentMonthCashFlow = incomeAmount - expenseAmount

	return nil
}

func (s *FinancialReportDataService) getMonthlyCashFlow(ctx context.Context) error {

	_, err := utils.GetUserID(ctx)
	if err != nil {
		return err
	}

	// generate loop for 12 last months
	for i := 0; i < 12; i++ {
		startDate := time.Now().AddDate(0, -i, 0)
		endDate := startDate.AddDate(0, 1, -1)
		_, incomeAmount, err := s.getIncomeRecordsByPeriod(ctx, startDate, endDate)
		if err != nil {
			return err
		}
		_, expenseAmount, err := s.getExpenseRecordsByPeriod(ctx, startDate, endDate)
		if err != nil {
			return err
		}
		s.financialReport.MonthlyCashFlow = append(s.financialReport.MonthlyCashFlow, entity.MonthlySummaryItem{
			Month:    startDate.Format("2006-01"),
			Revenue:  incomeAmount,
			Expenses: expenseAmount,
		})
	}

	return nil

}

func (s *FinancialReportDataService) getExpenseByCategory(ctx context.Context) error {

	_, err := utils.GetUserID(ctx)
	if err != nil {
		return err
	}

	// Struct para armazenar dados de despesa com valor e mês
	type ExpenseData struct {
		Value float64 `json:"value"`
		Month string  `json:"month"`
	}

	expense := make(map[string]ExpenseData)
	currentMonth := time.Now().Format("2006-01")

	for _, record := range s.expenseRecords {
		category := "desconhecido"
		if record.Category != "" {
			category = record.Category
		}

		start, _ := utils.ConvertStringToTime("2006-01-02", record.DueDate)
		end, _ := utils.ConvertStringToTime("2006-01-02", record.DueDate)

		if start.After(utils.GetFirstDayOfCurrentMonth()) && end.Before(utils.GetLastDayOfCurrentMonth()) {
			// Se a categoria já existe, soma o valor
			if existingData, exists := expense[category]; exists {
				expense[category] = ExpenseData{
					Value: existingData.Value + record.Amount,
					Month: currentMonth,
				}
			} else {
				// Se não existe, cria novo registro
				expense[category] = ExpenseData{
					Value: record.Amount,
					Month: currentMonth,
				}
			}
		}
	}

	if len(expense) > 0 {
		for category, data := range expense {
			s.financialReport.ExpenseByCategoryLast12Months = append(s.financialReport.ExpenseByCategoryLast12Months, entity.CategoryExpenseItem{
				Name:  category,
				Value: data.Value,
			})
		}
	}

	return nil
}

func (s *FinancialReportDataService) getExpenseByCategoryLast12Months(ctx context.Context) error {

	_, err := utils.GetUserID(ctx)
	if err != nil {
		return err
	}

	type CategoryExpenseWithMonth struct {
		Category string  `json:"category"`
		Value    float64 `json:"value"`
		Month    string  `json:"month"`
	}
	expense := make(map[string]CategoryExpenseWithMonth)

	// generate loop for 12 last months
	for i := 0; i < 12; i++ {
		startDate := time.Now().AddDate(0, -i, 0)
		endDate := startDate.AddDate(0, 1, -1)
		monthKey := startDate.Format("2006-01")

		for _, record := range s.expenseRecords {
			category := "desconhecido"
			if record.Category != "" {
				category = record.Category
			}

			start, _ := utils.ConvertStringToTime("2006-01-02", record.DueDate)
			end, _ := utils.ConvertStringToTime("2006-01-02", record.DueDate)

			if startDate.After(start) && endDate.Before(end) {
				// Criar chave única para categoria + mês
				key := fmt.Sprintf("%s_%s", category, monthKey)

				// Se a categoria-mês já existe, soma o valor
				if existingData, exists := expense[key]; exists {
					expense[key] = CategoryExpenseWithMonth{
						Category: category,
						Value:    existingData.Value + record.Amount,
						Month:    monthKey,
					}
				} else {
					// Se não existe, cria novo registro
					expense[key] = CategoryExpenseWithMonth{
						Category: category,
						Value:    record.Amount,
						Month:    monthKey,
					}
				}
			}
		}
	}
	if len(expense) > 0 {
		for _, data := range expense {
			s.financialReport.ExpenseByCategoryLast12Months = append(s.financialReport.ExpenseByCategoryLast12Months, entity.CategoryExpenseItem{
				Name:  data.Category,
				Value: data.Value,
			})
		}
	}

	return nil
}

func (s *FinancialReportDataService) mqConsumer(ctx context.Context) {

	// s.messageQueue.Consumer(ctx, mq_exchange, mq_queue_income)
}

func (s *FinancialReportDataService) processMQConsumer(body []byte, traceID string) error {

	return nil
}
