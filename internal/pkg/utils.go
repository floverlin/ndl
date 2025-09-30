package pkg

func ShortString(str string, length int) string {
	if len(str) > length {
		return str[:length]
	}
	return str
}

func SliceMap[T any, R any](s []T, f func(T) R) []R {
	n := []R{}
	for _, e := range s {
		n = append(n, f(e))
	}
	return n
}