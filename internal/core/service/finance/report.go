package finance

import (
	"context"
	"encoding/json"
	"fmt"
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
	s.getMonthlyCashFlow(ctx)
	s.getExpenseByCategory(ctx)
	s.getExpenseByCategoryLast12Months(ctx)

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

	// saldo do mês corrent (CurrentMonthCashFlow)
	//	Período: Refere-se sempre ao mês calendário atual (do dia 1 até o último dia do mês corrente).
	//	Cálculo: É a diferença simples entre suas receitas e despesas do mês:
	//	    (Total de Receitas do Mês Atual) - (Total de Despesas do Mês Atual)
	if s.financialReport.SummaryCards.CurrentMonthCashFlow == 0.00 {
		_, incomeAmount, err := s.getIncomeRecordsByPeriod(ctx, utils.GetFirstDayOfCurrentMonth(), utils.GetLastDayOfCurrentMonth())
		if err != nil {
			return err
		}

		_, expenseAmount, err := s.getExpenseRecordsByPeriod(ctx, utils.GetFirstDayOfCurrentMonth(), utils.GetLastDayOfCurrentMonth())
		if err != nil {
			return err
		}
		s.financialReport.SummaryCards.CurrentMonthCashFlow = incomeAmount - expenseAmount
	}

	//	VarVariação do saldo (CurrentMonthCashFlowChangePct)
	//
	// Período: Compara o saldo do mês atual com o saldo do mês anterior completo.
	// Cálculo: A variação percentual é calculada da seguinte forma:
	//	((Saldo do Mês Atual / Saldo do Mês Anterior) - 1) * 100
	//	Se não houver dados para o mês anterior, o backend deve retornar null para este campo.
	if s.financialReport.SummaryCards.CurrentMonthCashFlowChangePct == 0.00 {
		_, incomeAmount, err := s.getIncomeRecordsByPeriod(ctx, utils.GetFirstDayOfLastMonth(), utils.GetLastDayOfLastMonth())
		if err != nil {
			return err
		}

		_, expenseAmount, err := s.getExpenseRecordsByPeriod(ctx, utils.GetFirstDayOfLastMonth(), utils.GetLastDayOfLastMonth())
		if err != nil {
			return err
		}

		lastMonthCashFlow := incomeAmount - expenseAmount
		CurrentMonthCashFlowChangePct := ((s.financialReport.SummaryCards.CurrentMonthCashFlow / lastMonthCashFlow) - 1) * 100
		s.financialReport.SummaryCards.CurrentMonthCashFlowChangePct = CurrentMonthCashFlowChangePct
	}

	// Patrimonio liquido (NetWorth)
	// Período: Este é um "snapshot", representando o valor no momento atual da consulta.
	// Cálculo: É o valor total de tudo que você possui, menos o que você deve:
	// 		(Soma dos saldos de todas as contas) + (Valor atual de todos os investimentos) - (Total de dívidas)
	if s.financialReport.SummaryCards.NetWorth == 0.00 {
		s.financialReport.SummaryCards.NetWorth = 3.75
	}

	// Crescismento do patrimônio líquido (NetWorthChangePct)
	// 	Período: Compara o seu patrimônio líquido atual com o seu patrimônio líquido de 12 meses atrás.
	// 	Cálculo: A fórmula para a variação percentual é:
	// 			((Patrimônio Atual / Patrimônio de 12 Meses Atrás) - 1) * 100
	if s.financialReport.SummaryCards.NetWorthChangePercent == 0.00 {
		s.financialReport.SummaryCards.NetWorthChangePercent = 10.87
	}

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

		if record.DueDate.After(utils.GetFirstDayOfCurrentMonth()) && record.DueDate.Before(utils.GetLastDayOfCurrentMonth()) {
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

			if startDate.After(record.DueDate) && endDate.Before(record.DueDate) {
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
