package dip

import (
	"errors"
	"math"
	"sync"
)

// A 2n+1 x 2n+1 kernel for convolution/filtering.
type Kernel64 interface {
	// The weights of the kernel in row major order.
	Kernel() []float64
}

var (
	ErrKernelNotSquare = errors.New("kernel is not square")
	ErrKernelNotOdd    = errors.New("kernel size is not odd")
)

type fullKernel64 []float64

// Creates a new kernel from weights in row major order.
func NewKernel64(weights []float64) (Kernel64, error) {
	size := int(math.Sqrt(float64(len(weights))))
	if size*size != len(weights) {
		return nil, ErrKernelNotSquare
	}
	if size%2 == 0 {
		return nil, ErrKernelNotOdd
	}
	return fullKernel64(weights), nil
}

func (k fullKernel64) Kernel() []float64 {
	return k
}

type separableKernel64 struct {
	x, y []float64
}

// Creates a new kernel from horizontal and vertical weights.
func NewSeparableKernel64(x, y []float64) (Kernel64, error) {
	if len(x) != len(y) {
		return nil, ErrKernelNotSquare
	}
	if len(x)%2 == 0 {
		return nil, ErrKernelNotOdd
	}
	return &separableKernel64{x, y}, nil
}

func (sk *separableKernel64) Kernel() []float64 {
	size := len(sk.x)
	weights := make([]float64, size*size)
	for y := range sk.y {
		r := y * size
		for x := range sk.x {
			weights[r+x] = sk.y[y] * sk.x[x]
		}
	}
	return weights
}

func kernelSize(k Kernel64) int {
	switch k.(type) {
	case fullKernel64:
		return int(math.Sqrt(float64(len(k.(fullKernel64))))) / 2
	case *separableKernel64:
		return len(k.(*separableKernel64).x) / 2
	default:
		return int(math.Sqrt(float64(len(k.Kernel())))) / 2
	}
	panic("unreachable")
}

// Returns a Image64 that has been convolved with the kernel.
func (p *Image64) Convolved(k Kernel64) *Image64 {
	switch k.(type) {
	case *separableKernel64:
		return p.separableConvolution(k.(*separableKernel64))
	default:
		return p.fullConvolution(k)
	}
	panic("unreachable")
}

func (p *Image64) fullConvolution(k Kernel64) *Image64 {
	result := NewImage64(p.Rect)
	halfKernel := kernelSize(k)
	kernelStride := halfKernel*2 + 1
	weights := k.Kernel()
	width := p.Rect.Dx()
	height := p.Rect.Dy()
	var wg sync.WaitGroup

	// divide rows between go routines
	wg.Add(goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(0, height, t)
		go func(y0, y1 int) {

			// for each pixel
			for y := y0; y < y1; y++ {
				resultIndex := y * result.Stride
				for x := 0; x < width; x++ {

					// convolve
					v := float64(0)
					for yk := -halfKernel; yk <= halfKernel; yk++ {
						imageRowIndex := y + yk
						if imageRowIndex < 0 {
							imageRowIndex = 0
						} else if imageRowIndex >= height {
							imageRowIndex = height - 1
						}
						imageRowIndex *= p.Stride
						kernelRowIndex := (yk + halfKernel) * kernelStride
						for xk := -halfKernel; xk <= halfKernel; xk++ {
							imageColumnIndex := x + xk
							if imageColumnIndex < 0 {
								imageColumnIndex = 0
							} else if imageColumnIndex >= width {
								imageColumnIndex = width - 1
							}

							iv := float64(p.Pix[imageRowIndex+imageColumnIndex])
							kv := weights[kernelRowIndex+xk+halfKernel]
							v += iv * kv
						}
					}
					result.Pix[resultIndex] = Color64(v)

					resultIndex++
				}
			}
			wg.Done()
		}(i0, i1)
	}
	wg.Wait()

	return result
}

func (image *Image64) separableConvolution(k *separableKernel64) *Image64 {
	result := NewImage64(image.Rect)
	buffer := NewImage64(image.Rect)
	halfKernel := kernelSize(k)
	width := image.Rect.Dx()
	height := image.Rect.Dy()
	var wg sync.WaitGroup

	// divide rows between go routines
	wg.Add(goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(0, height, t)
		go func(y0, y1 int) {

			// for each pixel
			for y := y0; y < y1; y++ {
				imageRow := y * image.Stride
				bufferIndex := y * buffer.Stride
				for x := 0; x < width; x++ {

					// horizontal convolution
					v := float64(0)
					for xk := -halfKernel; xk <= halfKernel; xk++ {
						imageIndex := x + xk
						if imageIndex < 0 {
							imageIndex = 0
						} else if imageIndex >= width {
							imageIndex = width - 1
						}
						imageIndex += imageRow
						iv := float64(image.Pix[imageIndex])
						kv := k.x[xk+halfKernel]
						v += iv * kv
					}
					buffer.Pix[bufferIndex] = Color64(v)

					bufferIndex++
				}
			}
			wg.Done()
		}(i0, i1)
	}
	wg.Wait()

	// divide rows between go routines
	wg.Add(goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(0, height, t)
		go func(y0, y1 int) {

			// for each pixel
			for y := y0; y < y1; y++ {
				resultIndex := y * result.Stride
				for x := 0; x < width; x++ {

					// vertical convolution
					v := float64(0)
					for yk := -halfKernel; yk <= halfKernel; yk++ {
						bufferIndex := y + yk
						if bufferIndex < 0 {
							bufferIndex = 0
						} else if bufferIndex >= height {
							bufferIndex = height - 1
						}
						bufferIndex = bufferIndex*buffer.Stride + x
						iv := float64(buffer.Pix[bufferIndex])
						kv := k.y[yk+halfKernel]
						v += iv * kv
					}
					result.Pix[resultIndex] = Color64(v)

					resultIndex++
				}
			}
			wg.Done()
		}(i0, i1)
	}
	wg.Wait()

	return result
}
