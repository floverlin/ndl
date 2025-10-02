package pkg

func ShortString(str string, length int) string {
	if len(str) > length {
		return str[:length]
	}
	return str
}

func SliceMap[T any, R any](s []T, f func(T) (R, error)) ([]R, error) {
	n := []R{}
	for _, e := range s {
		r, err := f(e)
		if err != nil {
			return nil, err
		}
		n = append(n, r)
	}
	return n, nil
}

func SliceToMapMap[
	T any,
	OK comparable,
	OV any,
](
	s []T,
	f func(v T) (OK, OV, error),
) (map[OK]OV, error) {
	m := map[OK]OV{}
	for _, e := range s {
		k, v, err := f(e)
		if err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

func MapToSliceMap[
	IK comparable,
	IV any,
	R any,
](
	m map[IK]IV,
	f func(k IK, v IV) (R, error),
) ([]R, error) {
	r := []R{}
	for ik, iv := range m {
		v, err := f(ik, iv)
		if err != nil {
			return nil, err
		}
		r = append(r, v)
	}
	return r, nil
}

func MapMap[
	IK comparable,
	IV any,
	OK comparable,
	OV any,
](
	m map[IK]IV,
	f func(k IK, v IV) (OK, OV, error),
) (map[OK]OV, error) {
	nm := map[OK]OV{}
	for k, v := range m {
		nk, nv, err := f(k, v)
		if err != nil {
			return nil, err
		}
		nm[nk] = nv
	}
	return nm, nil
}
