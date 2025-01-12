package middleware

type ContextKey struct {
	name string
}

func (c ContextKey) String() string {
	return c.name
}

var (
	UserCtxKey            ContextKey = ContextKey{"user"}
	RequestMetaDataCtxKey ContextKey = ContextKey{"request_meta_data"}
)
