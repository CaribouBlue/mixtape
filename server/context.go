package server

type ContextKey struct {
	name string
}

func (c ContextKey) String() string {
	return c.name
}

var (
	UserCtxKey ContextKey = ContextKey{"user"}
)
