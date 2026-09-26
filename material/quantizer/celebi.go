// Changed by Avenge Media LLC from github.com/Nadim147c/material v3.1.2.

package quantizer

import (
	"context"

	"github.com/AvengeMedia/dankgo/material/color"
)

// QuantizeCelebi seeds QuantizeWsMeans with the result of QuantizeWu, which
// gives better clusters than either alone.
func QuantizeCelebi(input []color.ARGB, maxColors int) QuantizedMap {
	// ignore error because background context won't return any error
	qm, _ := QuantizeCelebiContext(context.Background(), input, maxColors)
	return qm
}

// QuantizeCelebiContext is QuantizeCelebi with context.Context support.
func QuantizeCelebiContext(
	ctx context.Context,
	input []color.ARGB,
	maxColors int,
) (QuantizedMap, error) {
	wu, err := QuantizeWuContext(ctx, input, maxColors)
	if err != nil {
		return nil, err
	}
	return QuantizeWsMeansContext(ctx, input, wu, maxColors)
}
