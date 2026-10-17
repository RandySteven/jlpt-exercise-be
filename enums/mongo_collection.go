package enums

type MongoCollection string

const (
	UserCollection MongoCollection = "user"
)

func (coll MongoCollection) ToString() string {
	return string(coll)
}
