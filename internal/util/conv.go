package util

import (
	"strconv"
)

func StringsToInts(ss []string) ([]int, error) {
	result := make([]int, 0, len(ss))
	for _, s := range ss {
		n, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, nil
}
