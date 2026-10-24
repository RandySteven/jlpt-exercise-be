package enums

type RouterPrefix string

const (
	BasePrefix     RouterPrefix = "/"
	AuthPrefix     RouterPrefix = "/auth"
	DummyPrefix    RouterPrefix = "/dummy"
	ExercisePrefix RouterPrefix = "/exercise"
	PaymentPrefix  RouterPrefix = "/payment"
	WebhookPrefix  RouterPrefix = "/webhook"
)

func (prefix RouterPrefix) ToString() string {
	return string(prefix)
}
