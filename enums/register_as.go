package enums

type RegisterAs string

const (
	RegisterAsUser   RegisterAs = "USER"
	RegisterAsAdmin  RegisterAs = "ADMIN"
	RegisterAsSensei RegisterAs = "SENSEI"
)

func (r RegisterAs) ToString() string {
	return string(r)
}
