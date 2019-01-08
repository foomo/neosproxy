package model

type Status struct {
	Workspaces      []string
	ProviderReports map[string]Report `json:"providerReports"`
	ConsumerReports map[string]Report `json:"consumerReports"`
}
