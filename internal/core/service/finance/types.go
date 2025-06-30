package finance

import "time"

// Message Queue attributes
const (
	mq_exchange            = "dashfin_finance"
	mq_queue_income        = "income_record"
	mq_queue_expense       = "expense_record"
	mq_rk_expense_create   = "expense.record.create"
	mq_rk_expense_delete   = "expense.record.delete"
	mq_rk_expense_update   = "expense.record.update"
	mq_rk_income_create    = "income.record.create"
	mq_rk_income_delete    = "income.record.delete"
	mq_rk_income_update    = "income.record.update"
	mq_queue_bank_account  = "bank_account"
	mq_queue_credit_card   = "credit_card"
	mq_queue_spending_plan = "spending_plan"
)

// Cache attributes
const (
	serviceCacheTTL = 1 * time.Minute

	cacheKeyIncomeReportByMonth      = "income_report_by_month"
	cacheKeyIncomeReportByLastMonth  = "income_report_by_last_month"
	cacheKeyIncomeReportByYear       = "income_report_by_year"
	cacheKeyExpenseReportByMonth     = "expense_report_by_month"
	cacheKeyExpenseReportByLastMonth = "expense_report_by_last_month"
	cacheKeyExpenseReportByYear      = "expense_report_by_year"
)

// Date and time
const (
	dateLayout = "2006-01-02"
)

// currentDate is a global variable holding the current date formatted as "2006-01-02".
// It's initialized once at the start of the program.
var (
	// currentDate         = time.Now().Format(dateLayout)
	firstDayOfMonth     = time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateLayout)
	lastDayOfMonth      = time.Now().AddDate(0, 1, -1).Format(dateLayout)
	firstDayOfLastMonth = time.Now().AddDate(0, -1, 0).Format(dateLayout)
	lastDayOfLastMonth  = time.Now().AddDate(0, 0, -1).Format(dateLayout)
	firstDayOfYear      = time.Date(time.Now().Year(), time.January, 1, 0, 0, 0, 0, time.UTC).Format(dateLayout)
	// lastDayOfYear       = time.Date(time.Now().Year(), time.December, 31, 0, 0, 0, 0, time.UTC).Format(dateLayout)
	// cuurentMonth        = time.Now().Month()
	// currentYear         = time.Now().Year()
)
