package shape

import (
	"math"

	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/svg"
	"oss.terrastruct.com/util-go/go2"
)

type shapeParallelogram struct {
	*baseShape
}

const parallelWedgeWidth = 26.

func NewParallelogram(box *geo.Box) Shape {
	shape := shapeParallelogram{
		baseShape: &baseShape{
			Type: PARALLELOGRAM_TYPE,
			Box:  box,
		},
	}
	shape.FullShape = go2.Pointer(Shape(shape))
	return shape
}

func (s shapeParallelogram) GetInnerBox() *geo.Box {
	tl := s.Box.TopLeft.Copy()
	width := s.Box.Width - 2*parallelWedgeWidth
	tl.X += parallelWedgeWidth
	return geo.NewBox(tl, width, s.Box.Height)
}

func parallelogramPath(box *geo.Box) *svg.SvgPathContext {
	wedgeWidth := parallelWedgeWidth
	// Note: box width should always be larger than parallelWedgeWidth
	// this just handles after collapsing into a line
	if box.Width <= wedgeWidth {
		wedgeWidth = box.Width / 2.0
	}
	pc := svg.NewSVGPathContext(box.TopLeft, 1, 1)
	pc.StartAt(pc.Absolute(wedgeWidth, 0))
	pc.L(false, box.Width, 0)
	pc.L(false, box.Width-wedgeWidth, box.Height)
	pc.L(false, 0, box.Height)
	pc.L(false, 0, box.Height)
	pc.Z()
	return pc
}

func (s shapeParallelogram) Perimeter() []geo.Intersectable {
	return parallelogramPath(s.Box).Path
}

func (s shapeParallelogram) GetSVGPathData() []string {
	return []string{
		parallelogramPath(s.Box).PathData(),
	}
}

func (s shapeParallelogram) GetDimensionsToFit(width, height, paddingX, paddingY float64) (float64, float64) {
	totalWidth := width + paddingX + parallelWedgeWidth*2
	return math.Ceil(totalWidth), math.Ceil(height + paddingY)
}
