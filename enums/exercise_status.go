package enums

const (
	InGeneratedStatus = `IN_GENERATED`
	GeneratedStatus   = `GENERATED`
	SubmittedStatus   = `SUBMITTED`
	CompletedStatus   = `COMPLETED`
	FailedStatus      = `FAILED`
)

// ExerciseInProgress reports whether the user still owns this exercise.
// A new generate is allowed after submit (SUBMITTED or COMPLETED) or FAILED.
func ExerciseInProgress(status string) bool {
	switch status {
	case InGeneratedStatus, GeneratedStatus:
		return true
	default:
		return false
	}
}
