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

func Reduce[T any, U any](list []T, reducer func(U, T) U, initialValue U) U {
	accumulator := initialValue
	for _, item := range list {
		accumulator = reducer(accumulator, item)
	}
	return accumulator
}

func Values[T comparable, U any](dict map[T]U) []U {
	values := make([]U, 0)
	for _, value := range dict {
		values = append(values, value)
	}
	return values
}

func Prepend[T any](slice []T, val T) []T {
	return append([]T{val}, slice...)
}
