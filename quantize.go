package main

func quantize(v float32) uint8 {
	if v < 0 {
		return 255
	}

	if v > 1 {
		v = 1
	}

	return uint8(v * 254)
}

func dequantize(v uint8) float32 {
	if v == 255 {
		return -1
	}

	return float32(v) / 254.0
}
