// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package quantizer

import (
	"cmp"
	"context"
	"math"
	"slices"

	"github.com/AvengeMedia/dankgo/material/color"
)

const (
	maxIterations       = 10
	minMovementDistance = 3.0
	wsmeansSeed         = 0x42688
)

type distanceIndex struct {
	distance float64
	index    int
}

// QuantizedMap is a map where ARGB is key and their frequencies as int
type QuantizedMap = map[color.ARGB]int

// QuantizeWsMeans is an image quantizer that improves on the speed of a
// standard K-Means algorithm by implementing several optimizations, including
// deduping identical pixels and a triangle inequality rule that reduces the
// number of comparisons needed to identify which cluster a point should be
// moved to.
//
// Wsmeans stands for Weighted Square Means.
//
// This algorithm was designed by M. Emre Celebi, and was found in their 2011
// paper, Improving the Performance of K-Means for Color Quantization.
// https://arxiv.org/abs/1101.0395
//
// It follows the Java QuantizerWsmeans, including its fixed random seed, so
// the same input always gives the same result.
func QuantizeWsMeans(input, startingClusters []color.ARGB, maxColors int) QuantizedMap {
	// ignore error because background context won't return any error
	qm, _ := QuantizeWsMeansContext(context.Background(), input, startingClusters, maxColors)
	return qm
}

// QuantizeWsMeansContext is QuantizeWsMeans with context.Context support.
func QuantizeWsMeansContext(
	ctx context.Context,
	input, startingClusters []color.ARGB,
	maxColors int,
) (QuantizedMap, error) {
	random := newJavaRandom(wsmeansSeed)

	pixelToCount := map[color.ARGB]int{}
	pixels := make([]color.ARGB, 0)
	for _, pixel := range input {
		if pixelToCount[pixel] == 0 {
			pixels = append(pixels, pixel)
		}
		pixelToCount[pixel]++
	}
	pointCount := len(pixels)
	points := make([]color.Lab, pointCount)
	counts := make([]int, pointCount)
	for i, pixel := range pixels {
		points[i] = pixel.ToLab()
		counts[i] = pixelToCount[pixel]
	}

	clusterCount := min(maxColors, pointCount)
	if len(startingClusters) > 0 {
		clusterCount = min(clusterCount, len(startingClusters))
	}
	if clusterCount <= 0 {
		return QuantizedMap{}, ctx.Err()
	}

	clusters := make([]color.Lab, clusterCount)
	for i := range clusters {
		if len(startingClusters) > 0 {
			clusters[i] = startingClusters[i].ToLab()
			continue
		}
		l := random.nextDouble() * 100.0
		a := random.nextDouble()*(100.0-(-100.0)+1) - 100
		b := random.nextDouble()*(100.0-(-100.0)+1) - 100
		clusters[i] = color.NewLab(l, a, b)
	}

	clusterIndices := make([]int, pointCount)
	for i := range clusterIndices {
		clusterIndices[i] = int(random.nextInt(int32(clusterCount)))
	}

	distanceToIndexMatrix := make([][]distanceIndex, clusterCount)
	for i := range distanceToIndexMatrix {
		distanceToIndexMatrix[i] = make([]distanceIndex, clusterCount)
		for j := range distanceToIndexMatrix[i] {
			distanceToIndexMatrix[i][j] = distanceIndex{distance: -1, index: -1}
		}
	}

	byDistance := func(a, b distanceIndex) int { return cmp.Compare(a.distance, b.distance) }
	pixelCountSums := make([]int, clusterCount)
	for iteration := range maxIterations {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		for i := range clusterCount {
			for j := i + 1; j < clusterCount; j++ {
				distance := clusters[i].DistanceSquared(clusters[j])
				distanceToIndexMatrix[j][i] = distanceIndex{distance: distance, index: i}
				distanceToIndexMatrix[i][j] = distanceIndex{distance: distance, index: j}
			}
			slices.SortStableFunc(distanceToIndexMatrix[i], byDistance)
		}

		pointsMoved := 0
		for i, point := range points {
			previousClusterIndex := clusterIndices[i]
			previousDistance := point.DistanceSquared(clusters[previousClusterIndex])
			minimumDistance := previousDistance
			newClusterIndex := -1
			for j := range clusterCount {
				// Upstream compares the jth nearest distance but moves to cluster j.
				if distanceToIndexMatrix[previousClusterIndex][j].distance >= 4*previousDistance {
					continue
				}
				distance := point.DistanceSquared(clusters[j])
				if distance < minimumDistance {
					minimumDistance = distance
					newClusterIndex = j
				}
			}
			if newClusterIndex == -1 {
				continue
			}
			if math.Abs(math.Sqrt(minimumDistance)-math.Sqrt(previousDistance)) > minMovementDistance {
				pointsMoved++
				clusterIndices[i] = newClusterIndex
			}
		}

		if pointsMoved == 0 && iteration != 0 {
			break
		}

		sumsL := make([]float64, clusterCount)
		sumsA := make([]float64, clusterCount)
		sumsB := make([]float64, clusterCount)
		clear(pixelCountSums)
		for i, point := range points {
			clusterIndex := clusterIndices[i]
			count := counts[i]
			pixelCountSums[clusterIndex] += count
			sumsL[clusterIndex] += point.L * float64(count)
			sumsA[clusterIndex] += point.A * float64(count)
			sumsB[clusterIndex] += point.B * float64(count)
		}

		for i := range clusterCount {
			count := float64(pixelCountSums[i])
			if count == 0 {
				clusters[i] = color.NewLab(0, 0, 0)
				continue
			}
			clusters[i] = color.NewLab(sumsL[i]/count, sumsA[i]/count, sumsB[i]/count)
		}
	}

	argbToPopulation := QuantizedMap{}
	for i := range clusterCount {
		count := pixelCountSums[i]
		if count == 0 {
			continue
		}
		argb := clusters[i].ToARGB()
		if _, ok := argbToPopulation[argb]; ok {
			continue
		}
		argbToPopulation[argb] = count
	}
	return argbToPopulation, ctx.Err()
}

// javaRandom is java.util.Random, the generator the Java port seeds.
type javaRandom struct {
	seed int64
}

const javaRandomMask = 1<<48 - 1

func newJavaRandom(seed int64) *javaRandom {
	return &javaRandom{seed: (seed ^ 0x5DEECE66D) & javaRandomMask}
}

func (r *javaRandom) next(bits uint) int32 {
	r.seed = (r.seed*0x5DEECE66D + 0xB) & javaRandomMask
	return int32(r.seed >> (48 - bits))
}

func (r *javaRandom) nextInt(bound int32) int32 {
	x := r.next(31)
	m := bound - 1
	if bound&m == 0 {
		return int32((int64(bound) * int64(x)) >> 31)
	}
	for u := x; ; u = r.next(31) {
		x = u % bound
		if u-x+m >= 0 {
			return x
		}
	}
}

func (r *javaRandom) nextDouble() float64 {
	return float64(int64(r.next(26))<<27+int64(r.next(27))) * 0x1.0p-53
}
