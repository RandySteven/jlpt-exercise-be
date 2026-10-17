package queries

type (
	GoQuery        string
	DropTable      string
	TableMigration string
)

func (q GoQuery) ToString() string {
	return string(q)
}

func (d DropTable) ToString() string {
	return string(d)
}

func (t TableMigration) ToString() string {
	return string(t)
}
