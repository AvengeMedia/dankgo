// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package quantizer

import (
	"context"
	"math"

	"github.com/AvengeMedia/dankgo/material/color"
)

const (
	indexBits  = 5
	sideLength = 33    // (1 << indexBits) + 1
	totalSize  = 35937 // sideLength^3
)

type direction int

const (
	directionRed direction = iota
	directionGreen
	directionBlue
)

type box struct {
	r0, r1 int
	g0, g1 int
	b0, b1 int
	vol    int
}

type maximizeResult struct {
	cutLocation int
	maximum     float64
}

// Moments are float64 like the JS numbers in QuantizerWu, so large images
// neither overflow nor truncate.
type quantizerWu struct {
	weights  []float64
	momentsR []float64
	momentsG []float64
	momentsB []float64
	moments  []float64
	cubes    []box
}

// QuantizeWu is an image quantizer that divides the image's pixels into
// clusters by recursively cutting an RGB cube, based on the weight of pixels in
// each area of the cube.
//
// The algorithm was described by Xiaolin Wu in Graphic Gems II, published in
// 1991.
func QuantizeWu(input []color.ARGB, maxColors int) []color.ARGB {
	// ignore error because background context won't return any error
	qw, _ := QuantizeWuContext(context.Background(), input, maxColors)
	return qw
}

// QuantizeWuContext is QuantizeWu with context.Context support.
func QuantizeWuContext(
	ctx context.Context,
	input []color.ARGB,
	maxColors int,
) ([]color.ARGB, error) {
	if maxColors <= 0 {
		return nil, nil
	}
	q := &quantizerWu{}
	q.constructHistogram(input)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	q.computeMoments()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	count := q.createBoxes(maxColors)
	return q.createResult(count), ctx.Err()
}

func getIndex(r, g, b int) int {
	return (r << (indexBits * 2)) + (r << (indexBits + 1)) + r + (g << indexBits) + g + b
}

func (q *quantizerWu) constructHistogram(pixels []color.ARGB) {
	q.weights = make([]float64, totalSize)
	q.momentsR = make([]float64, totalSize)
	q.momentsG = make([]float64, totalSize)
	q.momentsB = make([]float64, totalSize)
	q.moments = make([]float64, totalSize)

	const bitsToRemove = 8 - indexBits
	for _, pixel := range pixels {
		if pixel.Alpha() < 0xFF {
			continue
		}
		red := int(pixel.Red())
		green := int(pixel.Green())
		blue := int(pixel.Blue())
		index := getIndex((red>>bitsToRemove)+1, (green>>bitsToRemove)+1, (blue>>bitsToRemove)+1)
		q.weights[index]++
		q.momentsR[index] += float64(red)
		q.momentsG[index] += float64(green)
		q.momentsB[index] += float64(blue)
		q.moments[index] += float64(red*red + green*green + blue*blue)
	}
}

func (q *quantizerWu) computeMoments() {
	for r := 1; r < sideLength; r++ {
		var area, areaR, areaG, areaB, area2 [sideLength]float64
		for g := 1; g < sideLength; g++ {
			var line, lineR, lineG, lineB, line2 float64
			for b := 1; b < sideLength; b++ {
				index := getIndex(r, g, b)
				line += q.weights[index]
				lineR += q.momentsR[index]
				lineG += q.momentsG[index]
				lineB += q.momentsB[index]
				line2 += q.moments[index]

				area[b] += line
				areaR[b] += lineR
				areaG[b] += lineG
				areaB[b] += lineB
				area2[b] += line2

				previousIndex := getIndex(r-1, g, b)
				q.weights[index] = q.weights[previousIndex] + area[b]
				q.momentsR[index] = q.momentsR[previousIndex] + areaR[b]
				q.momentsG[index] = q.momentsG[previousIndex] + areaG[b]
				q.momentsB[index] = q.momentsB[previousIndex] + areaB[b]
				q.moments[index] = q.moments[previousIndex] + area2[b]
			}
		}
	}
}

func (q *quantizerWu) createBoxes(maxColors int) int {
	q.cubes = make([]box, maxColors)
	volumeVariance := make([]float64, maxColors)
	q.cubes[0].r1 = sideLength - 1
	q.cubes[0].g1 = sideLength - 1
	q.cubes[0].b1 = sideLength - 1

	generatedColorCount := maxColors
	next := 0
	for i := 1; i < maxColors; i++ {
		if q.cut(&q.cubes[next], &q.cubes[i]) {
			volumeVariance[next] = q.boxVariance(&q.cubes[next])
			volumeVariance[i] = q.boxVariance(&q.cubes[i])
		} else {
			volumeVariance[next] = 0
			i--
		}

		next = 0
		temp := volumeVariance[0]
		for j := 1; j <= i; j++ {
			if volumeVariance[j] > temp {
				temp = volumeVariance[j]
				next = j
			}
		}
		if temp <= 0 {
			generatedColorCount = i + 1
			break
		}
	}
	return generatedColorCount
}

func (q *quantizerWu) boxVariance(cube *box) float64 {
	if cube.vol <= 1 {
		return 0
	}
	return q.variance(cube)
}

func (q *quantizerWu) createResult(colorCount int) []color.ARGB {
	colors := make([]color.ARGB, 0, colorCount)
	for i := range colorCount {
		cube := &q.cubes[i]
		weight := q.volume(cube, q.weights)
		if weight <= 0 {
			continue
		}
		r := uint8(int(math.Round(q.volume(cube, q.momentsR)/weight)) & 0xFF)
		g := uint8(int(math.Round(q.volume(cube, q.momentsG)/weight)) & 0xFF)
		b := uint8(int(math.Round(q.volume(cube, q.momentsB)/weight)) & 0xFF)
		colors = append(colors, color.ARGBFromRGB(r, g, b))
	}
	return colors
}

func (q *quantizerWu) variance(cube *box) float64 {
	dr := q.volume(cube, q.momentsR)
	dg := q.volume(cube, q.momentsG)
	db := q.volume(cube, q.momentsB)
	xx := q.volume(cube, q.moments)
	hypotenuse := dr*dr + dg*dg + db*db
	return xx - hypotenuse/q.volume(cube, q.weights)
}

func (q *quantizerWu) cut(one, two *box) bool {
	wholeR := q.volume(one, q.momentsR)
	wholeG := q.volume(one, q.momentsG)
	wholeB := q.volume(one, q.momentsB)
	wholeW := q.volume(one, q.weights)

	maxR := q.maximize(one, directionRed, one.r0+1, one.r1, wholeR, wholeG, wholeB, wholeW)
	maxG := q.maximize(one, directionGreen, one.g0+1, one.g1, wholeR, wholeG, wholeB, wholeW)
	maxB := q.maximize(one, directionBlue, one.b0+1, one.b1, wholeR, wholeG, wholeB, wholeW)

	var dir direction
	switch {
	case maxR.maximum >= maxG.maximum && maxR.maximum >= maxB.maximum:
		if maxR.cutLocation < 0 {
			return false
		}
		dir = directionRed
	case maxG.maximum >= maxR.maximum && maxG.maximum >= maxB.maximum:
		dir = directionGreen
	default:
		dir = directionBlue
	}

	two.r1 = one.r1
	two.g1 = one.g1
	two.b1 = one.b1

	switch dir {
	case directionRed:
		one.r1 = maxR.cutLocation
		two.r0 = one.r1
		two.g0 = one.g0
		two.b0 = one.b0
	case directionGreen:
		one.g1 = maxG.cutLocation
		two.r0 = one.r0
		two.g0 = one.g1
		two.b0 = one.b0
	case directionBlue:
		one.b1 = maxB.cutLocation
		two.r0 = one.r0
		two.g0 = one.g0
		two.b0 = one.b1
	}

	one.vol = (one.r1 - one.r0) * (one.g1 - one.g0) * (one.b1 - one.b0)
	two.vol = (two.r1 - two.r0) * (two.g1 - two.g0) * (two.b1 - two.b0)
	return true
}

func (q *quantizerWu) maximize(
	cube *box,
	dir direction,
	first, last int,
	wholeR, wholeG, wholeB, wholeW float64,
) maximizeResult {
	bottomR := bottom(cube, dir, q.momentsR)
	bottomG := bottom(cube, dir, q.momentsG)
	bottomB := bottom(cube, dir, q.momentsB)
	bottomW := bottom(cube, dir, q.weights)

	result := maximizeResult{cutLocation: -1}
	for i := first; i < last; i++ {
		halfR := bottomR + top(cube, dir, i, q.momentsR)
		halfG := bottomG + top(cube, dir, i, q.momentsG)
		halfB := bottomB + top(cube, dir, i, q.momentsB)
		halfW := bottomW + top(cube, dir, i, q.weights)
		if halfW == 0 {
			continue
		}
		temp := (halfR*halfR + halfG*halfG + halfB*halfB) / halfW

		halfR = wholeR - halfR
		halfG = wholeG - halfG
		halfB = wholeB - halfB
		halfW = wholeW - halfW
		if halfW == 0 {
			continue
		}
		temp += (halfR*halfR + halfG*halfG + halfB*halfB) / halfW

		if temp > result.maximum {
			result = maximizeResult{cutLocation: i, maximum: temp}
		}
	}
	return result
}

func (q *quantizerWu) volume(cube *box, moment []float64) float64 {
	return moment[getIndex(cube.r1, cube.g1, cube.b1)] -
		moment[getIndex(cube.r1, cube.g1, cube.b0)] -
		moment[getIndex(cube.r1, cube.g0, cube.b1)] +
		moment[getIndex(cube.r1, cube.g0, cube.b0)] -
		moment[getIndex(cube.r0, cube.g1, cube.b1)] +
		moment[getIndex(cube.r0, cube.g1, cube.b0)] +
		moment[getIndex(cube.r0, cube.g0, cube.b1)] -
		moment[getIndex(cube.r0, cube.g0, cube.b0)]
}

func bottom(cube *box, dir direction, moment []float64) float64 {
	switch dir {
	case directionRed:
		return -moment[getIndex(cube.r0, cube.g1, cube.b1)] +
			moment[getIndex(cube.r0, cube.g1, cube.b0)] +
			moment[getIndex(cube.r0, cube.g0, cube.b1)] -
			moment[getIndex(cube.r0, cube.g0, cube.b0)]
	case directionGreen:
		return -moment[getIndex(cube.r1, cube.g0, cube.b1)] +
			moment[getIndex(cube.r1, cube.g0, cube.b0)] +
			moment[getIndex(cube.r0, cube.g0, cube.b1)] -
			moment[getIndex(cube.r0, cube.g0, cube.b0)]
	default:
		return -moment[getIndex(cube.r1, cube.g1, cube.b0)] +
			moment[getIndex(cube.r1, cube.g0, cube.b0)] +
			moment[getIndex(cube.r0, cube.g1, cube.b0)] -
			moment[getIndex(cube.r0, cube.g0, cube.b0)]
	}
}

func top(cube *box, dir direction, position int, moment []float64) float64 {
	switch dir {
	case directionRed:
		return moment[getIndex(position, cube.g1, cube.b1)] -
			moment[getIndex(position, cube.g1, cube.b0)] -
			moment[getIndex(position, cube.g0, cube.b1)] +
			moment[getIndex(position, cube.g0, cube.b0)]
	case directionGreen:
		return moment[getIndex(cube.r1, position, cube.b1)] -
			moment[getIndex(cube.r1, position, cube.b0)] -
			moment[getIndex(cube.r0, position, cube.b1)] +
			moment[getIndex(cube.r0, position, cube.b0)]
	default:
		return moment[getIndex(cube.r1, cube.g1, position)] -
			moment[getIndex(cube.r1, cube.g0, position)] -
			moment[getIndex(cube.r0, cube.g1, position)] +
			moment[getIndex(cube.r0, cube.g0, position)]
	}
}
