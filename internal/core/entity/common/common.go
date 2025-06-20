package entity_common

type ActionEvent string

const (
	ActionCreate ActionEvent = "create"
	ActionUpdate ActionEvent = "update"
	ActionDelete ActionEvent = "delete"
)

const (
	MessageQueueExchangeName = "dashfin_finance"
)
