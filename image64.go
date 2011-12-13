package dip

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"sync"
)

type Image64 struct {
	Pix    []Color64
	Stride int
	Rect   image.Rectangle
}

func NewImage64(r image.Rectangle) *Image64 {
	width, height := r.Dx(), r.Dy()
	return &Image64{
		Pix:    make([]Color64, width*height),
		Stride: width,
		Rect:   image.Rectangle{image.ZP, image.Point{width, height}},
	}
}

func (p *Image64) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return Color64(0)
	}
	return p.Pix[(y-p.Rect.Min.Y)*p.Stride+(x-p.Rect.Min.X)]
}

func (p *Image64) Bounds() image.Rectangle {
	return p.Rect
}

func (p *Image64) ColorModel() color.Model {
	return Color64Model
}

func (p *Image64) Opaque() bool {
	return true
}

func (p *Image64) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	p.Pix[(y-p.Rect.Min.Y)*p.Stride+(x-p.Rect.Min.X)] =
		toColor64(c).(Color64)
}

func (p *Image64) SetColor64(x, y int, c Color64) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	p.Pix[(y-p.Rect.Min.Y)*p.Stride+(x-p.Rect.Min.X)] = c
}

func (p *Image64) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	if r.Empty() {
		return &Image64{}
	}
	i := (r.Min.Y-p.Rect.Min.Y)*p.Stride + (r.Min.X - p.Rect.Min.X)
	return &Image64{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}

func ToImage64(i image.Image) *Image64 {
	b := i.Bounds()
	r := NewImage64(b)
	draw.Draw(r, r.Rect, i, b.Min, draw.Src)
	return r
}

func (img *Image64) Normalize() {
	// find min and max
	type minMax struct {
		min, max Color64
	}
	minMaxCh := make(chan minMax, goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(img.Rect.Min.Y, img.Rect.Max.Y, t)
		go func(y0, y1 int) {
			mm := minMax{Color64(math.MaxFloat64), Color64(-math.MaxFloat64)}
			for y := y0; y < y1; y++ {
				i := (y - img.Rect.Min.Y) * img.Stride
				for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
					if img.Pix[i] < mm.min {
						mm.min = img.Pix[i]
					} else if img.Pix[i] > mm.max {
						mm.max = img.Pix[i]
					}
					i++
				}
			}
			minMaxCh <- mm
		}(i0, i1)
	}
	mm := <-minMaxCh
	for t := 1; t < goRoutines; t++ {
		v := <-minMaxCh
		if v.min < mm.min {
			mm.min = v.min
		}
		if v.max > mm.max {
			mm.max = v.max
		}
	}

	// normalize
	scale := 1.0 / (mm.max - mm.min)
	var wg sync.WaitGroup
	wg.Add(goRoutines)
	for t := 0; t < goRoutines; t++ {
		i0, i1 := goRountineRanges(img.Rect.Min.Y, img.Rect.Max.Y, t)
		go func(y0, y1 int) {
			for y := y0; y < y1; y++ {
				i := (y - img.Rect.Min.Y) * img.Stride
				for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
					img.Pix[i] = (img.Pix[i] - mm.min) * scale
					i++
				}
			}
			wg.Done()
		}(i0, i1)
	}
	wg.Wait()
}
