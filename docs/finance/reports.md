
# Backend Go: API de Relatórios Financeiros

Este documento descreve a API para fornecer os dados para a página de Relatórios Financeiros.

## Visão Geral

A API de Relatórios Financeiros é responsável por coletar, agregar e formatar dados de receitas, despesas e patrimônio do usuário para alimentar os diversos gráficos e seções da página de relatórios.

**Path Base da API:** `/api/finance/reports`

**Autenticação:** A rota requer autenticação via token Firebase ID no header `X-AUTHORIZATION`. O `X-USERID` também é esperado e será usado para filtrar os dados do usuário correto.

**Criptografia:** O payload de resposta para este endpoint GET deve ser criptografado usando o padrão AES-CBC definido no projeto. O backend deve retornar o JSON no formato `{ "payload": "base64_encrypted_string" }`.

## Estrutura de Dados da Resposta (Go Struct)

A resposta completa da API deve ser uma única struct `FinancialReportData`.

```go
package main // ou seu pacote de models

import "time"

// ExpenseSubCategoryItem define a estrutura para uma subcategoria de despesa.
type ExpenseSubCategoryItem struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

// ExpenseCategoryWithSubItems define uma categoria principal com suas subcategorias.
type ExpenseCategoryWithSubItems struct {
	Name     string                 `json:"name"`
	Value    float64                `json:"value"` // Total da categoria
	Fill     string                 `json:"fill"`
	Children []ExpenseSubCategoryItem `json:"children"`
}

// FinancialReportData define a estrutura completa para a página de relatórios.
type FinancialReportData struct {
	SummaryCards                  ReportSummaryCards              `json:"summaryCards"`
	MonthlyCashFlow               []MonthlySummaryItem            `json:"monthlyCashFlow"`                // Para o gráfico de barras Receitas vs. Despesas
	ExpenseByCategory             []CategoryExpenseItem           `json:"expenseByCategory"`              // Para o gráfico de donut
	ExpenseByCategoryLast12Months []CategoryExpenseItem           `json:"expenseByCategoryLast12Months"` // Para o gráfico Treemap (12 meses)
	NetWorthEvolution             []NetWorthHistoryItem           `json:"netWorthEvolution"`              // Para o gráfico de linha do Patrimônio
	ExpenseBreakdown              []ExpenseCategoryWithSubItems   `json:"expenseBreakdown"`               // Para o gráfico Sunburst/Donut Aninhado
}

// ReportSummaryCards contém os dados para os cards de destaque.
type ReportSummaryCards struct {
    CurrentMonthCashFlow          float64  `json:"currentMonthCashFlow"`          // Saldo do Mês: (Receitas do Mês) - (Despesas do Mês)
    CurrentMonthCashFlowChangePct *float64 `json:"currentMonthCashFlowChangePct,omitempty"` // Variação % do saldo do mês em relação ao mês anterior
    NetWorth                      float64  `json:"netWorth"`                      // Valor ATUAL do Patrimônio Líquido
    NetWorthChangePercent         float64  `json:"netWorthChangePercent"`         // Variação % do patrimônio em relação a 12 meses atrás
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
```

## Endpoint da API

### 1. Obter Dados dos Relatórios Financeiros

*   **Método:** `GET`
*   **Path:** `/api/finance/reports`
*   **Autenticação:** Obrigatória.
*   **Descrição:** Retorna todos os dados necessários para a página de Relatórios Financeiros.
*   **Lógica do Backend:**
    1.  Verificar autenticação do usuário.
    2.  **Calcular `SummaryCards`:**
        *   `CurrentMonthCashFlow`: (Total de Receitas do Mês Atual) - (Total de Despesas do Mês Atual). Usado no card "Saldo do Mês".
        *   `CurrentMonthCashFlowChangePct`: ((Saldo Mês Atual / Saldo Mês Anterior) - 1) * 100. Retornar `null` se o mês anterior não tiver dados.
        *   `NetWorth`: Valor ATUAL do Patrimônio Líquido. (Soma dos Saldos de Contas) + (Valor Atual dos Ativos de Investimento) - (Dívidas).
        *   `NetWorthChangePercent`: ((Patrimônio Atual / Patrimônio 12 meses atrás) - 1) * 100. Usado no card "Crescimento (12M)".
    3.  **Calcular `MonthlyCashFlow`:**
        *   Para cada um dos últimos 6 ou 12 meses, agregar a soma de todas as receitas e a soma de todas as despesas.
        *   Formatar o mês como "Mmm/aa" (ex: "Jul/24").
    4.  **Calcular `ExpenseByCategory`:**
        *   Para o **mês atual**, agregar a soma de todas as despesas por categoria.
        *   Associar uma cor a cada categoria para consistência no gráfico (ex: `hsl(var(--chart-1))`).
    5.  **Calcular `ExpenseByCategoryLast12Months`:**
        *   Para os **últimos 12 meses**, agregar a soma de todas as despesas por categoria.
        *   Associar a mesma cor do item correspondente em `ExpenseByCategory` para consistência.
    6.  **Calcular `NetWorthEvolution`:**
        *   Para cada um dos últimos 6 ou 12 meses, calcular o patrimônio líquido no final daquele mês.
        *   Formatar o mês como "Mmm/aa".
    7.  **Calcular `ExpenseBreakdown` (NOVO):**
        *   Para o **mês atual**, primeiro agregue os gastos por categoria (como em `ExpenseByCategory`).
        *   Depois, para cada categoria, agregue os gastos por subcategoria (ou por descrição, se não houver subcategoria).
        *   Monte a estrutura aninhada `[]ExpenseCategoryWithSubItems`. A soma dos `children` de uma categoria deve ser igual ao `value` da categoria pai.
    8.  Montar a struct `FinancialReportData` completa.
    9.  Criptografar o JSON da struct e envolvê-lo em `PayloadWrapper`.
    10. Retornar os dados.

*   **Resposta de Sucesso (200 OK):**
    JSON contendo o payload criptografado. Exemplo do JSON **antes** da criptografia:
    ```json
    {
      "summaryCards": {
        "currentMonthCashFlow": 2700.00,
        "currentMonthCashFlowChangePct": 5.8,
        "netWorth": 38200.00,
        "netWorthChangePercent": 12.5
      },
      "monthlyCashFlow": [
        { "month": "Jan/24", "revenue": 7200, "expenses": 4500 }
      ],
      "expenseByCategory": [
        { "name": "Moradia", "value": 1800, "fill": "hsl(var(--chart-1))" }
      ],
      "expenseByCategoryLast12Months": [
        { "name": "Moradia", "value": 21600, "fill": "hsl(var(--chart-1))" }
      ],
      "netWorthEvolution": [
        { "date": "Jan/24", "value": 25000 }
      ],
      "expenseBreakdown": [
        {
          "name": "Moradia",
          "value": 1800,
          "fill": "hsl(var(--chart-1))",
          "children": [
            { "name": "Aluguel", "value": 1500 },
            { "name": "Energia", "value": 200 },
            { "name": "Internet", "value": 100 }
          ]
        },
        {
          "name": "Alimentação",
          "value": 1200,
          "fill": "hsl(var(--chart-2))",
          "children": [
            { "name": "Supermercado", "value": 800 },
            { "name": "Restaurantes", "value": 400 }
          ]
        }
      ]
    }
    ```

*   **Respostas de Erro:**
    *   `401 Unauthorized`: Token inválido ou ausente.
    *   `500 Internal Server Error`: Erro ao buscar ou agregar os dados.

## Análises Derivadas no Frontend

É importante notar que o frontend pode realizar análises adicionais com base nos dados fornecidos por esta API. O backend **não precisa** pré-calcular ou fornecer dados para os seguintes componentes, pois eles são derivados dos dados acima no cliente:

*   **Card "Receitas (Últimos 12M)":** O frontend calcula o total somando todos os valores `revenue` do array `monthlyCashFlow`.
*   **Card "Despesas (Últimos 12M)":** O frontend calcula o total somando todos os valores `expenses` do array `monthlyCashFlow`.
*   **Gráfico "Top 5 Maiores Despesas (Mês)":** O frontend ordena o array `expenseByCategory` por `value` em ordem decrescente e pega os 5 primeiros itens para exibir.
*   **Gráfico "Visão Geral de Despesas (Mês)":** O frontend utiliza o array completo `expenseByCategory` para montar este gráfico.

Manter essa separação de responsabilidades permite que o backend se concentre em fornecer os dados brutos agregados, enquanto o frontend lida com a apresentação e derivações visuais.

## Análise de Tendência e Projeções com Média Móvel Exponencial (EMA)

Para implementar a análise de tendências que dá mais peso aos meses recentes, a **Média Móvel Exponencial (EMA)** é uma excelente abordagem. Ela reage mais rapidamente a mudanças recentes nos dados do que uma média móvel simples.

### O Cálculo

A EMA é calculada iterativamente com a seguinte fórmula:

1.  **Fator de Suavização (α - Alpha):** Define o peso dado aos valores mais recentes.
    `α = 2 / (N + 1)`
    Onde `N` é o número de períodos (meses) que você quer considerar na média (ex: 3, 6, 12). Um `N` menor torna a média mais sensível a mudanças recentes.

2.  **Cálculo da EMA:**
    `EMA_atual = (Valor_atual * α) + (EMA_anterior * (1 - α))`

3.  **Primeiro Valor da EMA:** O primeiro valor da série de EMA geralmente é uma Média Móvel Simples (SMA) dos primeiros `N` períodos, ou pode ser simplesmente o primeiro valor de dados da série para simplificar.

### Exemplo de Implementação em Go

Abaixo está um exemplo de como você pode implementar o cálculo da EMA no backend Go para uma série de despesas mensais.

```go
package main

import (
	"fmt"
	"time"
)

// MonthlyData representa um dado financeiro para um mês específico.
type MonthlyData struct {
	Date  time.Time
	Value float64
}

// CalculateEMA calcula a Média Móvel Exponencial para uma série de dados.
// Retorna um slice de float64 com os valores da EMA para cada ponto de dados a partir do período N.
func CalculateEMA(data []MonthlyData, period int) ([]float64, error) {
	if len(data) < period {
		return nil, fmt.Errorf("dados insuficientes para o período de %d meses", period)
	}

	// 1. Calcular o fator de suavização (alpha)
	alpha := 2.0 / (float64(period) + 1.0)

	var emas []float64

	// 2. Calcular a primeira EMA como uma Média Móvel Simples (SMA)
	var sum float64
	for i := 0; i < period; i++ {
		sum += data[i].Value
	}
	firstEma := sum / float64(period)
	emas = append(emas, firstEma)

	// 3. Calcular as EMAs subsequentes
	for i := period; i < len(data); i++ {
		currentValue := data[i].Value
		previousEma := emas[len(emas)-1]
		
		currentEma := (currentValue * alpha) + (previousEma * (1 - alpha))
		emas = append(emas, currentEma)
	}

	return emas, nil
}

func main() {
	// Exemplo de uso com dados de despesas mensais
	// Supondo que você buscou estes dados do seu banco de dados
	monthlyExpenses := []MonthlyData{
		{Date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), Value: 4500}, // Jan
		{Date: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC), Value: 4800}, // Fev
		{Date: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC), Value: 4600}, // Mar
		{Date: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), Value: 5100}, // Abr
		{Date: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), Value: 4900}, // Mai
		{Date: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), Value: 5300}, // Jun
	}

	// Calcular a EMA de 3 meses
	ema3, err := CalculateEMA(monthlyExpenses, 3)
	if err != nil {
		fmt.Println("Erro:", err)
		return
	}

	// A série EMA começa após o primeiro período de cálculo da média simples.
	// O primeiro valor da EMA corresponde ao terceiro mês (Março).
	fmt.Println("Tendência de Despesas (EMA de 3 meses):")
	for i, emaValue := range ema3 {
		// O índice do dado original é i + (period - 1)
		monthIndex := i + (3 - 1)
		fmt.Printf("Mês: %s, EMA: %.2f\n", monthlyExpenses[monthIndex].Date.Format("Jan/06"), emaValue)
	}

    // Para determinar a tendência, compare as duas últimas EMAs
    if len(ema3) >= 2 {
        lastEma := ema3[len(ema3)-1]
        penultimateEma := ema3[len(ema3)-2]
        
        if lastEma > penultimateEma {
            fmt.Println("\nConclusão: A tendência das despesas é de ALTA.")
        } else if lastEma < penultimateEma {
            fmt.Println("\nConclusão: A tendência das despesas é de BAIXA.")
        } else {
            fmt.Println("\nConclusão: A tendência das despesas está ESTÁVEL.")
        }
    }
}
```

### Como Utilizar no Endpoint

O backend pode usar a função `CalculateEMA` para analisar os dados de `monthlyCashFlow` (receitas e despesas) e determinar se a tendência é de alta, baixa ou estável. Essa informação pode ser retornada em um novo campo na resposta da API, por exemplo:

```go
type FinancialReportData struct {
    // ... outros campos ...
    ExpensesTrend string `json:"expensesTrend,omitempty"` // "alta", "baixa", "estavel"
    RevenueTrend  string `json:"revenueTrend,omitempty"`
}
```

Isso permite que o frontend exiba um ícone de tendência ou uma mensagem informativa, agregando muito valor à análise do usuário.

---

## Endpoint Separado: Planejado vs. Gasto (Mês Anterior)

Para alimentar o gráfico **"Orçamento: Planejado vs. Gasto"**, é necessário um endpoint separado que retorna a comparação para o mês fechado anterior.

*   **Método:** `GET`
*   **Path:** `/api/finance/dashboard/planned-vs-actual`
*   **Autenticação:** Obrigatória.
*   **Descrição:** Retorna uma comparação de valores planejados versus gastos reais por categoria para o mês fechado anterior.
*   **Estrutura da Resposta (Go Struct):** O endpoint deve retornar um array de `PlannedVsActualCategory`.
    ```go
    // PlannedVsActualCategory representa os dados de comparação por categoria
    type PlannedVsActualCategory struct {
        Category        string  `json:"category"`        // ID da categoria (ex: "food")
        Label           string  `json:"label"`           // Nome para exibição (ex: "Alimentação")
        PlannedAmount   float64 `json:"plannedAmount"`   // Valor orçado para a categoria
        ActualAmount    float64 `json:"actualAmount"`    // Valor real gasto na categoria
        SpentPercentage float64 `json:"spentPercentage"` // (actualAmount / plannedAmount) * 100
        Fill            string  `json:"fill"`            // Cor HSL para o gráfico
    }
    ```
*   **Exemplo de Resposta JSON:**
    ```json
    [
      {
        "category": "food",
        "label": "Alimentação", 
        "plannedAmount": 1200.00,
        "actualAmount": 980.50,
        "spentPercentage": 81.71,
        "fill": "hsl(var(--chart-2))"
      },
      {
        "category": "transport",
        "label": "Transporte",
        "plannedAmount": 800.00, 
        "actualAmount": 1050.00,
        "spentPercentage": 131.25,
        "fill": "hsl(var(--chart-3))"
      }
    ]
    ```
*   **Lógica do Backend:**
    1.  Determinar o período do **mês fechado anterior**.
    2.  Buscar o plano de gastos (`spending_plans`) do usuário.
    3.  Agregar as despesas reais (`expenses`) do usuário para o período, agrupadas por categoria.
    4.  Para cada categoria no plano de gastos, montar o objeto `PlannedVsActualCategory` comparando com os gastos reais.
    5.  Retornar o array resultante.
*   **Referência:** Para uma guia de implementação detalhada, consulte `src/data/dashboard/prompt.md`.

```
