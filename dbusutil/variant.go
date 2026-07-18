package dbusutil

import "github.com/godbus/dbus/v5"

func As[T any](v dbus.Variant) (T, bool) {
	val, ok := v.Value().(T)
	return val, ok
}

func AsOr[T any](v dbus.Variant, def T) T {
	if val, ok := v.Value().(T); ok {
		return val
	}
	return def
}

func Get[T any](m map[string]dbus.Variant, key string) (T, bool) {
	v, ok := m[key]
	if !ok {
		var zero T
		return zero, false
	}
	return As[T](v)
}

func GetOr[T any](m map[string]dbus.Variant, key string, def T) T {
	v, ok := m[key]
	if !ok {
		return def
	}
	return AsOr(v, def)
}

func Normalize(v any) any {
	switch val := v.(type) {
	case dbus.Variant:
		return Normalize(val.Value())
	case dbus.ObjectPath:
		return string(val)
	case []dbus.ObjectPath:
		result := make([]string, len(val))
		for i, p := range val {
			result[i] = string(p)
		}
		return result
	case map[string]dbus.Variant:
		result := make(map[string]any)
		for k, vv := range val {
			result[k] = Normalize(vv.Value())
		}
		return result
	case map[string]any:
		result := make(map[string]any)
		for k, vv := range val {
			result[k] = Normalize(vv)
		}
		return result
	case map[dbus.ObjectPath]map[string]map[string]dbus.Variant:
		result := make(map[string]any)
		for path, ifaces := range val {
			ifaceMap := make(map[string]any)
			for ifaceName, props := range ifaces {
				propMap := make(map[string]any)
				for propName, propVal := range props {
					propMap[propName] = Normalize(propVal.Value())
				}
				ifaceMap[ifaceName] = propMap
			}
			result[string(path)] = ifaceMap
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = Normalize(item)
		}
		return result
	case []dbus.Variant:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = Normalize(item.Value())
		}
		return result
	default:
		return v
	}
}

func NormalizeSlice(values []any) []any {
	result := make([]any, len(values))
	for i, v := range values {
		result[i] = Normalize(v)
	}
	return result
}
