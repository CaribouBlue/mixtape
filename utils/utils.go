package utils

import "strings"

func Map[T any, U any](list []T, mapper func(T) U) []U {
	mappedList := make([]U, len(list))
	for i, item := range list {
		mappedList[i] = mapper(item)
	}
	return mappedList
}

func MapJoin[T any](list []T, separator string, mapper func(T) string) string {
	return strings.Join(Map(list, mapper), separator)
}
