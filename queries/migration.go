package queries

const (
	MigrateWorkflow TableMigration = `CREATE TABLE IF NOT EXISTS workflows (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		run_id VARCHAR(255) NOT NULL,
		workflow_id VARCHAR(255) NOT NULL,
		activity_name VARCHAR(255) NOT NULL,
		metadata JSON NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		deleted_at TIMESTAMP NULL,

		INDEX idx_run_id (run_id),
		INDEX idx_workflow_id (workflow_id),
		INDEX idx_activity_name (activity_name)
	)`

	MigrateUser TableMigration = `CREATE TABLE IF NOT EXISTS users ()`
)
