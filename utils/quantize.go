package utils

func Quantize(v float32) uint8 {
	if v < 0 {
		return 255
	}

	if v > 1 {
		v = 1
	}

	return uint8(v * 254)
}
