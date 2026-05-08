package main

func limiter(value float64, maxValue float64) float64 {
	result := value / maxValue

	if result > 1.0 {
		result = 1.0
	}

	if result < 0 {
		result = 0.0
	}

	return result
}
