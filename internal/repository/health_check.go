package repository

type HealthCheckRepository interface {
	PingDatabase() error
}

func (r *repository) PingDatabase() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
