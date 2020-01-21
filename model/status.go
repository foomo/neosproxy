package model

type Status struct {
	Workspace       string
	ProviderReports map[string]Report `json:"providerReports"`
	ConsumerReports map[string]Report `json:"consumerReports"`
}
