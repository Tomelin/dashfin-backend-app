package entity_finance

import (
	"context"
	"time"
)

type FinancialReportDataRepositoryInterface interface {
	GetFinancialReportData(ctx context.Context, id *string) (*FinancialReportData, error)
	SaveReport(ctx context.Context, userID string, report *FinancialReportData, ttl time.Duration) error
}

type FinancialReportDataServiceInterface interface {
	GetFinancialReportData(ctx context.Context) (*FinancialReportData, error)
}

// ExpenseSubCategoryItem define a estrutura para uma subcategoria de despesa.
type ExpenseSubCategoryItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// ExpenseCategoryWithSubItems define uma categoria principal com suas subcategorias.
type ExpenseCategoryWithSubItems struct {
	Name     string                   `json:"name"`
	Value    float64                  `json:"value"` // Total da categoria
	Fill     string                   `json:"fill"`
	Children []ExpenseSubCategoryItem `json:"children"`
}

// FinancialReportData define a estrutura completa para a página de relatórios.
type FinancialReportData struct {
	SummaryCards                  ReportSummaryCards            `json:"summaryCards"`
	MonthlyCashFlow               []MonthlySummaryItem          `json:"monthlyCashFlow"`               // Para o gráfico de barras Receitas vs. Despesas
	ExpenseByCategory             []CategoryExpenseItem         `json:"expenseByCategory"`             // Para o gráfico de donut
	ExpenseByCategoryLast12Months []CategoryExpenseItem         `json:"expenseByCategoryLast12Months"` // Para o gráfico Treemap (12 meses)
	NetWorthEvolution             []NetWorthHistoryItem         `json:"netWorthEvolution"`             // Para o gráfico de linha do Patrimônio
	ExpenseBreakdown              []ExpenseCategoryWithSubItems `json:"expenseBreakdown"`              // Para o gráfico Sunburst/Donut Aninhado
}

// ReportSummaryCards contém os dados para os cards de destaque.
type ReportSummaryCards struct {
	CurrentMonthCashFlow          float64 `json:"currentMonthCashFlow"`                    // Saldo do Mês: (Receitas do Mês) - (Despesas do Mês)
	CurrentMonthCashFlowChangePct float64 `json:"currentMonthCashFlowChangePct,omitempty"` // Variação % do saldo do mês em relação ao mês anterior
	NetWorth                      float64 `json:"netWorth"`                                // Valor ATUAL do Patrimônio Líquido
	NetWorthChangePercent         float64 `json:"netWorthChangePercent"`                   // Variação % do patrimônio em relação a 12 meses atrás
}

// MonthlySummaryItem representa o resumo de um mês para o gráfico de fluxo de caixa.
type MonthlySummaryItem struct {
	Month    string  `json:"month"`    // Formato "Mês/Ano", ex: "Jan/24"
	Revenue  float64 `json:"revenue"`  // Total de receitas no mês
	Expenses float64 `json:"expenses"` // Total de despesas no mês
}

// CategoryExpenseItem representa o gasto total de uma categoria.
type CategoryExpenseItem struct {
	Name  string  `json:"name"`  // Nome da categoria, ex: "Moradia"
	Value float64 `json:"value"` // Valor total gasto na categoria
	Fill  string  `json:"fill"`  // Cor para o gráfico (pode ser pré-definida no backend)
}

// NetWorthHistoryItem representa um ponto no histórico de evolução do patrimônio.
type NetWorthHistoryItem struct {
	Date  string  `json:"date"`  // Formato "Mês/Ano", ex: "Jan/24"
	Value float64 `json:"value"` // Valor do patrimônio líquido no final daquela data
}
