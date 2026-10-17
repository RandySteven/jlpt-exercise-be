package jobs

import "context"

type IUserJob interface {
	UpdateUserStatus(ctx context.Context) error
}
