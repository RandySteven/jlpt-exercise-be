package enums

type RouterPrefix string

const (
	BasePrefix  RouterPrefix = "/"
	AuthPrefix  RouterPrefix = "/auth"
	DummyPrefix RouterPrefix = "/dummy"
)

func (prefix RouterPrefix) ToString() string {
	return string(prefix)
}
