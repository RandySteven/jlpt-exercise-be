package enums

type RegisterAs string

const (
	RegisterAsUser RegisterAs = "USER"
)

func (r RegisterAs) ToString() string {
	return string(r)
}
