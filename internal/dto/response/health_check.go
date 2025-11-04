package response

type HealthCheckRes struct {
	ResponseTime         float64 `json:"response_time"`
	DatabaseConnection   bool    `json:"database_connection"`
	DatabaseResponseTime float64 `json:"database_responseTime"`
	MemoryConsumption    float64 `json:"memory_consumption"`
}
