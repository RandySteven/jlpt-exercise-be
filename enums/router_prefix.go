package enums

type RouterPrefix string

const (
	BasePrefix     RouterPrefix = "/"
	AuthPrefix     RouterPrefix = "/auth"
	DummyPrefix    RouterPrefix = "/dummy"
	ExercisePrefix RouterPrefix = "/exercise"
)

func (prefix RouterPrefix) ToString() string {
	return string(prefix)
}
