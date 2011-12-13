package dip

import (
	"errors"
	"math"
	"sync"
)

type Kernel64 interface {
	Size() int
	Kernel() []float64
}

var (
	ErrKernelNotSquare = errors.New("kernel is not square")
	ErrKernelNotOdd    = errors.New("kernel size is not odd")
)

type fullKernel64 []float64

func NewKernel64(w []float64) (Kernel64, error) {
	s := int(math.Sqrt(float64(len(w))))
	if s*s != len(w) {
		return nil, ErrKernelNotSquare
	}
	if s%2 == 0 {
		return nil, ErrKernelNotOdd
	}
	return fullKernel64(w), nil
}

func (k fullKernel64) Kernel() []float64 {
	return k
}

func (k fullKernel64) Size() int {
	return int(math.Sqrt(float64(len(k))))/2 - 1
}

type separableKernel64 struct {
	x, y []float64
}

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
	n := len(sk.x)
	k := make([]float64, n*n)
	for y := range sk.y {
		r := y * n
		for x := range sk.x {
			k[r+x] = sk.y[y] * sk.x[x]
		}
	}
	return k
}

func (sk *separableKernel64) Size() int {
	return len(sk.x) / 2
}

func (p *Image64) Convolved(k Kernel64) *Image64 {
	switch k.(type) {
	// case sepKernel64:
	// 	return p.separableConvolution(k.(sepKernel64))
	default:
		return p.fullConvolution(k)
	}
	panic("unreachable")
}

func (p *Image64) fullConvolution(k Kernel64) *Image64 {
	r := NewImage64(p.Rect)
	hk := k.Size()
	ks := hk*2 + 1
	w := k.Kernel()
	width := p.Rect.Dx()
	height := p.Rect.Dy()
	var wg sync.WaitGroup

	wg.Add(goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(0, height, t)
		go func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				resultIndex := y * r.Stride
				for x := 0; x < width; x++ {
					v := float64(0.0)
					for yk := -hk; yk <= hk; yk++ {
						imageRowIndex := y + yk
						if imageRowIndex < 0 {
							imageRowIndex = 0
						} else if imageRowIndex >= height {
							imageRowIndex = height - 1
						}
						imageRowIndex *= p.Stride
						kernelRowIndex := (yk + hk) * ks
						for xk := -hk; xk <= hk; xk++ {
							imageColumnIndex := x + xk
							if imageColumnIndex < 0 {
								imageColumnIndex = 0
							} else if imageColumnIndex >= width {
								imageColumnIndex = width - 1
							}

							iv := float64(p.Pix[imageRowIndex+imageColumnIndex])
							kv := w[kernelRowIndex+xk+hk]
							v += iv * kv
						}
					}
					r.Pix[resultIndex] = Color64(v)
					resultIndex++
				}
			}
			wg.Done()
		}(i0, i1)
	}
	wg.Wait()

	return r
}

func (p *Image64) separableConvolution(k separableKernel64) *Image64 {
	return p
}
