package main

import (
	"iter"
)

func Map[T, U any](slice []T, fn func(value T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}

	return result
}

func Filter[T any](slice []T, fn func(value T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}

	return result[:len(result):len(result)]
}

func FilterIter[T any](seq iter.Seq[T], fn func(value T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !fn(v) {
				continue
			}

			if !yield(v) {
				return
			}
		}
	}
}

func MapIter[T, U any](seq iter.Seq[T], fn func(value T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

func MapIter12[T, K, V any](seq iter.Seq[T], fn func(value T) (K, V)) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for v := range seq {
			if !yield(fn(v)) {
				return
			}
		}
	}
}

func MapIter21[K, V, T any](seq iter.Seq2[K, V], fn func(k K, v V) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for k, v := range seq {
			if !yield(fn(k, v)) {
				return
			}
		}
	}
}

func CollectOrError[T any](seq iter.Seq2[T, error]) ([]T, error) {
	result := make([]T, 0)
	for v, err := range seq {
		if err != nil {
			return nil, err
		}
		result = append(result, v)
	}

	return result, nil
}

func Zip[T, U any](seq1 iter.Seq[T], seq2 iter.Seq[U]) iter.Seq2[T, U] {
	return func(yield func(T, U) bool) {
		next1, stop1 := iter.Pull(seq1)
		defer stop1()

		next2, stop2 := iter.Pull(seq2)
		defer stop2()

		for {
			v1, ok1 := next1()
			v2, ok2 := next2()

			if !ok1 || !ok2 {
				return
			}

			if !yield(v1, v2) {
				return
			}
		}
	}
}
