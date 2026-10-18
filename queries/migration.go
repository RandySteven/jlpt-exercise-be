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

	MigrateOnboarding TableMigration = `
		CREATE TABLE IF NOT EXISTS onboardings (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			phone_number TEXT NOT NULL,
			address TEXT NOT NULL,
			register_as TEXT NOT NULL,
			status TEXT NOT NULL,
			profile_picture_url TEXT,
			additional_info JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL
		)
	`

	MigrateOnboardingRegisterAsIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_onboardings_register_as ON onboardings (register_as)
	`

	MigrateOnboardingStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_onboardings_status ON onboardings (status)
	`

	MigrateUser TableMigration = `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			onboarding_id BIGINT NOT NULL UNIQUE,
			user_name TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL,
			additional_info JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (onboarding_id) REFERENCES onboardings(id) ON UPDATE CASCADE
		)
	`

	MigrateUserStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_users_status ON users (status)
	`

	MigrateExercise TableMigration = `
		CREATE TABLE IF NOT EXISTS exercises (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			n_level INT NOT NULL,
			exercise JSONB NOT NULL,
			score INT NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			user_id BIGINT NOT NULL,
			note TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE
		)
	`

	MigrateExerciseReferenceID TableMigration = `
		ALTER TABLE exercises
		ADD COLUMN IF NOT EXISTS reference_id TEXT NOT NULL DEFAULT ''
	`

	MigrateExerciseUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_exercises_user_id ON exercises (user_id)
	`

	MigrateExerciseStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_exercises_status ON exercises (status)
	`

	MigrateExerciseNLevelIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_exercises_n_level ON exercises (n_level)
	`

	MigrateExerciseReferenceIDIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_exercises_reference_id ON exercises (reference_id)
	`
)
