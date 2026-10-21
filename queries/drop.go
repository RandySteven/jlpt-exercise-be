package queries

const (
	DropSubscriptionTable     DropTable = `DROP TABLE IF EXISTS subscriptions`
	DropPaymentTable          DropTable = `DROP TABLE IF EXISTS payments`
	DropSubscriptionPlanTable DropTable = `DROP TABLE IF EXISTS subscription_plans`
	DropExerciseTable         DropTable = `DROP TABLE IF EXISTS exercises`
	DropUserTable             DropTable = `DROP TABLE IF EXISTS users`
	DropOnboardingTable       DropTable = `DROP TABLE IF EXISTS onboardings`
	DropWorklowTable          DropTable = `DROP TABLE IF EXISTS workflows`
)
