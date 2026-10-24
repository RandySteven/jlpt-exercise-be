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
			exercise JSONB NOT NULL DEFAULT '{}'::jsonb,
			answer JSONB NOT NULL DEFAULT '{}'::jsonb,
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

	MigrateExerciseAnswer TableMigration = `
		ALTER TABLE exercises
		ADD COLUMN IF NOT EXISTS answer JSONB NOT NULL DEFAULT '{}'::jsonb
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

	DropExerciseReferenceIDIndex TableMigration = `
		DROP INDEX IF EXISTS idx_exercises_reference_id
	`

	MigrateExerciseReferenceIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_exercises_reference_id
		ON exercises (reference_id)
		WHERE deleted_at IS NULL
			AND reference_id <> ''
	`

	MigrateExerciseOneOpenPerUserIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_exercises_one_open_per_user
		ON exercises (user_id)
		WHERE deleted_at IS NULL
			AND status IN ('IN_GENERATED', 'GENERATED')
	`

	MigrateSubscriptionPlan TableMigration = `
		CREATE TABLE IF NOT EXISTS subscription_plans (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			price INT NOT NULL,
			currency TEXT NOT NULL,
			interval TEXT NOT NULL,
			interval_count INT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL
		)
	`

	MigrateSubscriptionPlanColumns TableMigration = `
		ALTER TABLE subscription_plans
			ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS price INT NOT NULL DEFAULT 0,
			ADD COLUMN IF NOT EXISTS currency TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS interval TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS interval_count INT NOT NULL DEFAULT 1,
			ADD COLUMN IF NOT EXISTS created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP DEFAULT NULL
	`

	MigrateSubscriptionPlanNameIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_subscription_plans_name
		ON subscription_plans (name)
		WHERE deleted_at IS NULL
	`

	SeedSubscriptionPlans TableMigration = `
		INSERT INTO subscription_plans (name, price, currency, interval, interval_count)
		VALUES
			('N5', 15000, 'IDR', 'month', 1),
			('N4', 15000, 'IDR', 'month', 1),
			('N3', 15000, 'IDR', 'month', 1),
			('N2', 15000, 'IDR', 'month', 1),
			('N1', 15000, 'IDR', 'month', 1)
		ON CONFLICT (name) WHERE deleted_at IS NULL DO NOTHING
	`

	MigrateSubscription TableMigration = `
		CREATE TABLE IF NOT EXISTS subscriptions (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			user_id BIGINT NOT NULL,
			plan_id BIGINT NOT NULL,
			start_date TIMESTAMP NOT NULL,
			end_date TIMESTAMP NOT NULL,
			status TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE,
			FOREIGN KEY (plan_id) REFERENCES subscription_plans(id) ON UPDATE CASCADE
		)
	`

	MigrateSubscriptionUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions (user_id)
	`

	MigrateSubscriptionPlanIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_id ON subscriptions (plan_id)
	`

	MigrateSubscriptionStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions (status)
	`

	MigrateSubscriptionReferenceID TableMigration = `
		ALTER TABLE subscriptions
		ADD COLUMN IF NOT EXISTS reference_id TEXT NOT NULL DEFAULT ''
	`

	MigrateSubscriptionReferenceIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_reference_id
		ON subscriptions (reference_id)
		WHERE deleted_at IS NULL
			AND reference_id <> ''
	`

	MigratePayment TableMigration = `
		CREATE TABLE IF NOT EXISTS payments (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			user_id BIGINT NOT NULL,
			subscription_id BIGINT NOT NULL,
			amount INT NOT NULL,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE,
			FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON UPDATE CASCADE
		)
	`

	MigratePaymentUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id)
	`

	MigratePaymentSubscriptionIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_payments_subscription_id ON payments (subscription_id)
	`

	MigratePaymentStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status)
	`

	MigratePaymentReferenceIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_reference_id
		ON payments (reference_id)
		WHERE deleted_at IS NULL
			AND reference_id <> ''
	`
)
