package dip

import (
	"image/color"
)

// Color64 represnets a float64 grayscale color. The value does not need to be
// in the range of [0, 1] though.
type Color64 float64

func (c Color64) RGBA() (uint32, uint32, uint32, uint32) {
	gray := uint32(c * 0xffff)
	alpha := uint32(0xffff)
	return gray, gray, gray, alpha
}

func toColor64(c color.Color) color.Color {
	switch c.(type) {
	// nothing to do
	case Color64:
		return c
	// color is already gray, just scale to [0, 1]
	case color.Gray:
		return Color64(float64(c.(color.Gray).Y) / 255)
	case color.Gray16:
		return Color64(float64(c.(color.Gray16).Y) / 65535)
	// convert color to gray and scale to [0, 1]
	default:
		r, g, b, _ := c.RGBA()
		return Color64(float64(299*r+587*g+114*b) / 1000 / 0xffff)
	}
	panic("unreachable")
}

// Color64Model is the Model for Color64 colors.
var Color64Model color.Model = color.ModelFunc(toColor64)
