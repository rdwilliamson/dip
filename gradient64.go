package dip

func SobelX64() Kernel64 {
	return &separableKernel64{[]float64{-1, 0, 1}, []float64{1, 2, 1}}
}

func SobelY64() Kernel64 {
	return &separableKernel64{[]float64{1, 2, 1}, []float64{-1, 0, 1}}
}
