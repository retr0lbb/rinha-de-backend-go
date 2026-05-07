package main

func limiter(value float32, maxValue float32) float32 {
	result := value / maxValue

	if result > 1.0 {
		result = 1.0
	}

	if result < 0 {
		result = 0.0
	}

	return result
}
