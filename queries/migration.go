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

	MigrateSensei TableMigration = `
		CREATE TABLE IF NOT EXISTS senseis (
			id BIGSERIAL PRIMARY KEY,
			onboarding_id BIGINT NOT NULL UNIQUE,
			sensei_name TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL,
			additional_info JSONB NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (onboarding_id) REFERENCES onboardings(id) ON UPDATE CASCADE
		)
	`

	MigrateSenseiStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_senseis_status ON senseis (status)
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
			feature_flags JSONB NOT NULL DEFAULT '{}'::jsonb,
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
			ADD COLUMN IF NOT EXISTS feature_flags JSONB NOT NULL DEFAULT '{}'::jsonb,
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
			subscription_id TEXT NOT NULL DEFAULT '',
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

	MigrateSubscriptionID TableMigration = `
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
					AND table_name = 'subscriptions'
					AND column_name = 'reference_id'
			) THEN
				ALTER TABLE subscriptions RENAME COLUMN reference_id TO subscription_id;
			END IF;
			IF NOT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
					AND table_name = 'subscriptions'
					AND column_name = 'subscription_id'
			) THEN
				ALTER TABLE subscriptions
				ADD COLUMN subscription_id TEXT NOT NULL DEFAULT '';
			END IF;
		END $$;
	`

	DropSubscriptionReferenceIDIndex TableMigration = `
		DROP INDEX IF EXISTS idx_subscriptions_reference_id
	`

	MigrateSubscriptionIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_subscriptions_subscription_id
		ON subscriptions (subscription_id)
		WHERE deleted_at IS NULL
			AND subscription_id <> ''
	`

	MigratePayment TableMigration = `
		CREATE TABLE IF NOT EXISTS payments (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			transaction_id TEXT NOT NULL DEFAULT '',
			transaction_type TEXT NOT NULL DEFAULT '',
			user_id BIGINT NOT NULL,
			amount INT NOT NULL,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE
		)
	`

	MigratePaymentColumns TableMigration = `
		ALTER TABLE payments
			ADD COLUMN IF NOT EXISTS transaction_id TEXT NOT NULL DEFAULT '',
			ADD COLUMN IF NOT EXISTS transaction_type TEXT NOT NULL DEFAULT ''
	`

	MigratePaymentUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id)
	`

	MigratePaymentTransactionIDIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments (transaction_id)
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

	MigrateClass TableMigration = `
		CREATE TABLE IF NOT EXISTS classes (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			sensei_id BIGINT NOT NULL,
			schedules JSONB NOT NULL DEFAULT '[]'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (sensei_id) REFERENCES senseis(id) ON UPDATE CASCADE
		)
	`

	MigrateClassSenseiIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_classes_sensei_id ON classes (sensei_id)
	`

	MigrateClassUser TableMigration = `
		CREATE TABLE IF NOT EXISTS class_users (
			id BIGSERIAL PRIMARY KEY,
			class_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (class_id) REFERENCES classes(id) ON UPDATE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE
		)
	`

	MigrateClassUserClassIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_class_users_class_id ON class_users (class_id)
	`

	MigrateClassUserUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_class_users_user_id ON class_users (user_id)
	`

	MigrateClassUserUniqueIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_class_users_class_id_user_id
		ON class_users (class_id, user_id)
		WHERE deleted_at IS NULL
	`

	MigrateClassRoom TableMigration = `
		CREATE TABLE IF NOT EXISTS class_rooms (
			id BIGSERIAL PRIMARY KEY,
			class_room_name TEXT NOT NULL,
			room_code TEXT NOT NULL,
			created_by BIGINT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL
		)
	`

	MigrateClassRoomCodeIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_class_rooms_room_code
		ON class_rooms (room_code)
		WHERE deleted_at IS NULL
	`

	MigrateClassRoomCreatedByIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_class_rooms_created_by ON class_rooms (created_by)
	`

	MigrateClassCategory TableMigration = `
		CREATE TABLE IF NOT EXISTS class_categories (
			id BIGSERIAL PRIMARY KEY,
			category TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL
		)
	`

	MigrateClassCategoryIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_class_categories_category
		ON class_categories (category)
		WHERE deleted_at IS NULL
	`

	MigrateClassCatalogue TableMigration = `
		CREATE TABLE IF NOT EXISTS class_catalogues (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			category_id BIGINT NOT NULL,
			description TEXT NOT NULL,
			additional_info JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (category_id) REFERENCES class_categories(id) ON UPDATE CASCADE
		)
	`

	MigrateClassCatalogueCategoryIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_class_catalogues_category_id ON class_catalogues (category_id)
	`

	MigrateTransactionHistory TableMigration = `
		CREATE TABLE IF NOT EXISTS transaction_histories (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			transaction_id TEXT NOT NULL DEFAULT '',
			transaction_type TEXT NOT NULL DEFAULT '',
			transaction_sub_type TEXT NOT NULL DEFAULT '',
			user_id BIGINT NOT NULL,
			amount INT NOT NULL,
			currency TEXT NOT NULL,
			status TEXT NOT NULL,
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE
		)
	`

	MigrateTransactionHistoryColumns TableMigration = `
		ALTER TABLE transaction_histories
			ADD COLUMN IF NOT EXISTS transaction_sub_type TEXT NOT NULL DEFAULT ''
	`

	MigrateTransactionHistoryUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_transaction_histories_user_id ON transaction_histories (user_id)
	`

	MigrateTransactionHistoryTransactionIDIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_transaction_histories_transaction_id ON transaction_histories (transaction_id)
	`

	MigrateTransactionHistoryTransactionTypeIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_transaction_histories_transaction_type ON transaction_histories (transaction_type)
	`

	MigrateTransactionHistoryTransactionSubTypeIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_transaction_histories_transaction_sub_type ON transaction_histories (transaction_sub_type)
	`

	MigrateTransactionHistoryStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_transaction_histories_status ON transaction_histories (status)
	`

	MigrateTransactionHistoryReferenceIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_transaction_histories_reference_id
		ON transaction_histories (reference_id)
		WHERE deleted_at IS NULL
			AND reference_id <> ''
	`

	MigrateRefund TableMigration = `
		CREATE TABLE IF NOT EXISTS refunds (
			id BIGSERIAL PRIMARY KEY,
			reference_id TEXT NOT NULL DEFAULT '',
			original_reference_id TEXT NOT NULL DEFAULT '',
			payment_id BIGINT NOT NULL,
			user_id BIGINT NOT NULL,
			amount INT NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP DEFAULT NULL,
			FOREIGN KEY (payment_id) REFERENCES payments(id) ON UPDATE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users(id) ON UPDATE CASCADE
		)
	`

	MigrateRefundPaymentIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_refunds_payment_id ON refunds (payment_id)
	`

	MigrateRefundUserIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_refunds_user_id ON refunds (user_id)
	`

	MigrateRefundStatusIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds (status)
	`

	MigrateRefundReferenceIDIndex TableMigration = `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_refunds_reference_id
		ON refunds (reference_id)
		WHERE deleted_at IS NULL
			AND reference_id <> ''
	`

	MigrateRefundOriginalReferenceIDIndex TableMigration = `
		CREATE INDEX IF NOT EXISTS idx_refunds_original_reference_id ON refunds (original_reference_id)
	`
)
