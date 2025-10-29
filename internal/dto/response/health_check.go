package response

type HealthCheckRes struct {
	ResponseTime         float64 `json:"responseTime"`
	DatabaseConnection   bool    `json:"databaseConnection"`
	DatabaseResponseTime float64 `json:"databaseResponseTime"`
	MemoryConsumption    float64 `json:"memoryConsumption"`
}
