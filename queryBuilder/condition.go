package queryBuilder

type Condition interface {
	assemble() string
}
