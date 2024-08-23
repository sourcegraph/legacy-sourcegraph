package shape

import (
	"math"

	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/svg"
	"oss.terrastruct.com/util-go/go2"
)

type shapeDiamond struct {
	*baseShape
}

func NewDiamond(box *geo.Box) Shape {
	shape := shapeDiamond{
		baseShape: &baseShape{
			Type: DIAMOND_TYPE,
			Box:  box,
		},
	}
	shape.FullShape = go2.Pointer(Shape(shape))
	return shape
}

func (s shapeDiamond) GetInnerBox() *geo.Box {
	width := s.Box.Width
	height := s.Box.Height
	tl := s.Box.TopLeft.Copy()
	tl.X += width / 4.
	tl.Y += height / 4.
	width /= 2.
	height /= 2.
	return geo.NewBox(tl, width, height)
}

func diamondPath(box *geo.Box) *svg.SvgPathContext {
	pc := svg.NewSVGPathContext(box.TopLeft, box.Width/77, box.Height/76.9)
	pc.StartAt(pc.Absolute(38.5, 76.9))
	pc.C(true, -0.3, 0, -0.5, -0.1, -0.7, -0.3)
	pc.L(false, 0.3, 39.2)
	pc.C(true, -0.4, -0.4, -0.4, -1, 0, -1.4)
	pc.L(false, 37.8, 0.3)
	pc.C(true, 0.4, -0.4, 1, -0.4, 1.4, 0)
	pc.L(true, 37.5, 37.5)
	pc.C(true, 0.4, 0.4, 0.4, 1, 0, 1.4)
	pc.L(false, 39.2, 76.6)
	pc.C(false, 39, 76.8, 38.8, 76.9, 38.5, 76.9)
	pc.Z()
	return pc
}

func (s shapeDiamond) Perimeter() []geo.Intersectable {
	return diamondPath(s.Box).Path
}

func (s shapeDiamond) GetSVGPathData() []string {
	return []string{
		diamondPath(s.Box).PathData(),
	}
}

func (s shapeDiamond) GetDimensionsToFit(width, height, paddingX, paddingY float64) (float64, float64) {
	totalWidth := 2 * (width + paddingX)
	totalHeight := 2 * (height + paddingY)
	return math.Ceil(totalWidth), math.Ceil(totalHeight)
}

func (s shapeDiamond) GetDefaultPadding() (paddingX, paddingY float64) {
	return defaultPadding / 4, defaultPadding / 2
}
