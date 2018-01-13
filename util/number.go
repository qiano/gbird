package util

import (
	"math"
)

func Minimum(first interface{}, rest ...interface{}) interface{} {
	minimum := first

	for _, v := range rest {
		switch v.(type) {
		case int:
			if v := v.(int); v < minimum.(int) {
				minimum = v
			}
		case float64:
			if v := v.(float64); v < minimum.(float64) {
				minimum = v
			}
		case string:
			if v := v.(string); v < minimum.(string) {
				minimum = v
			}
		}
	}
	return minimum
}

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}
