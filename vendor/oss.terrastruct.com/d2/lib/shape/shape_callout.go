package shape

import (
	"math"

	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/svg"
	"oss.terrastruct.com/util-go/go2"
)

type shapeCallout struct {
	*baseShape
}

const (
	defaultTipWidth  = 30.
	defaultTipHeight = 45.
)

func NewCallout(box *geo.Box) Shape {
	shape := shapeCallout{
		baseShape: &baseShape{
			Type: CALLOUT_TYPE,
			Box:  box,
		},
	}
	shape.FullShape = go2.Pointer(Shape(shape))
	return shape
}

func getTipWidth(box *geo.Box) float64 {
	tipWidth := defaultTipWidth
	if box.Width < tipWidth*2 {
		tipWidth = box.Width / 2.0
	}
	return tipWidth
}

func getTipHeight(box *geo.Box) float64 {
	tipHeight := defaultTipHeight
	if box.Height < tipHeight*2 {
		tipHeight = box.Height / 2.0
	}
	return tipHeight
}

func (s shapeCallout) GetInnerBox() *geo.Box {
	tipHeight := getTipHeight(s.Box)
	height := s.Box.Height - tipHeight
	return geo.NewBox(s.Box.TopLeft.Copy(), s.Box.Width, height)
}

func calloutPath(box *geo.Box) *svg.SvgPathContext {
	tipWidth := getTipWidth(box)
	tipHeight := getTipHeight(box)
	pc := svg.NewSVGPathContext(box.TopLeft, 1, 1)
	pc.StartAt(pc.Absolute(0, 0))
	pc.V(true, box.Height-tipHeight)
	pc.H(true, box.Width/2.0)
	pc.V(true, tipHeight)
	pc.L(true, tipWidth, -tipHeight)
	pc.H(true, box.Width/2.0-tipWidth)
	pc.V(true, -(box.Height - tipHeight))
	pc.H(true, -box.Width)
	pc.Z()
	return pc
}

func (s shapeCallout) Perimeter() []geo.Intersectable {
	return calloutPath(s.Box).Path
}

func (s shapeCallout) GetSVGPathData() []string {
	return []string{
		calloutPath(s.Box).PathData(),
	}
}

func (s shapeCallout) GetDimensionsToFit(width, height, paddingX, paddingY float64) (float64, float64) {
	// return the minimum shape dimensions needed to fit content (width x height)
	// in the shape's innerBox with padding
	baseHeight := height + paddingY
	if baseHeight < defaultTipHeight {
		baseHeight *= 2
	} else {
		baseHeight += defaultTipHeight
	}
	return math.Ceil(width + paddingX), math.Ceil(baseHeight)
}

func (s shapeCallout) GetDefaultPadding() (paddingX, paddingY float64) {
	return defaultPadding, defaultPadding / 2
}
