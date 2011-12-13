package dip

import (
	"image/color"
)

type Color64 float64

func (c Color64) RGBA() (r, g, b, a uint32) {
	r = uint32(c * 0xffff)
	a = 0xffff
	return r, r, r, a
}

func toColor64(c color.Color) color.Color {
	switch c.(type) {
	case Color64:
		return c
	case color.Gray:
		return Color64(float64(c.(color.Gray).Y) / 255)
	case color.Gray16:
		return Color64(float64(c.(color.Gray16).Y) / 65535)
	default:
		r, g, b, _ := c.RGBA()
		return Color64(float64(299*r+587*g+114*b) / 1000 / 0xffff)
	}
	panic("unreachable")
}

var Color64Model color.Model = color.ModelFunc(toColor64)
