package enums

type (
	ClassUserStatus string
)

const (
	ClassUserStatusActive   ClassUserStatus = "ACTIVE"
	ClassUserStatusInactive ClassUserStatus = "INACTIVE"
	ClassUserStatusPending  ClassUserStatus = "PENDING"
)

func (c ClassUserStatus) String() string {
	return string(c)
}
