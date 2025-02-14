package utils

import "net/http"

type RequestBasedFactory[T any] func(r *http.Request) T
