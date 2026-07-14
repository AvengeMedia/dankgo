package httpapi

import (
	"net/url"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gorilla/schema"
)

var formDecoder = schema.NewDecoder()

var urlEncodedFormat = huma.Format{
	Marshal: nil,
	Unmarshal: func(data []byte, v any) error {
		values, err := url.ParseQuery(string(data))
		if err != nil {
			return err
		}

		// Huma validates bodies by first parsing into *any before decoding
		// into the target struct, but gorilla/schema needs a struct — so map
		// url.Values into a map[string]any for that pass.
		// See: https://github.com/danielgtaylor/huma/blob/main/huma.go#L1264
		if vPtr, ok := v.(*any); ok {
			m := map[string]any{}
			for k, vals := range values {
				switch len(vals) {
				case 0:
				case 1:
					m[k] = vals[0]
				default:
					m[k] = vals
				}
			}
			*vPtr = m
			return nil
		}

		return formDecoder.Decode(v, values)
	},
}
