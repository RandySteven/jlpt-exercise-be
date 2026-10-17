package enums

type ContextKey string

const (
	UserID     ContextKey = `user_id`
	RoleID     ContextKey = `role_id`
	RequestID  ContextKey = `request_id`
	Env        ContextKey = `env`
	ClientIP   ContextKey = `client_ip`
	FileHeader ContextKey = `file_header`
	FileObject ContextKey = `file_object`
)

func (c ContextKey) ToString() string {
	return string(c)
}
