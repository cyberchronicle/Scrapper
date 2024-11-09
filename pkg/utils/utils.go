package utils

func CreateRangeSlice(ranges ...[2]int64) []int64 {
	var result []int64
	for _, r := range ranges {
		for i := r[0]; i <= r[1]; i++ {
			result = append(result, i)
		}
	}
	return result
}
