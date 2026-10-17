package jobs

type (
	Job struct {
		UserJob IUserJob
	}
)

func NewJob(userJob IUserJob) *Job {
	return &Job{
		UserJob: userJob,
	}
}
