package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	p := map[string]any{"key": "value"}
	val, err := Get[string](p, "key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	_, err = Get[string](p, "missing")
	assert.Error(t, err)

	_, err = Get[int](p, "key")
	assert.Error(t, err)
}

func TestGetOpt(t *testing.T) {
	p := map[string]any{"key": "value"}
	assert.Equal(t, "value", GetOpt(p, "key", "default"))
	assert.Equal(t, "default", GetOpt(p, "missing", "default"))
}

func TestString(t *testing.T) {
	p := map[string]any{"s": "hello", "n": 123}
	val, err := String(p, "s")
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)

	_, err = String(p, "n")
	assert.Error(t, err)
}

func TestStringNonEmpty(t *testing.T) {
	p := map[string]any{"s": "hello", "empty": ""}
	val, err := StringNonEmpty(p, "s")
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)

	_, err = StringNonEmpty(p, "empty")
	assert.Error(t, err)

	_, err = StringNonEmpty(p, "missing")
	assert.Error(t, err)
}

func TestStringOpt(t *testing.T) {
	p := map[string]any{"s": "hello"}
	assert.Equal(t, "hello", StringOpt(p, "s", "default"))
	assert.Equal(t, "default", StringOpt(p, "missing", "default"))
}

func TestInt(t *testing.T) {
	p := map[string]any{"n": float64(42), "s": "str"}
	val, err := Int(p, "n")
	assert.NoError(t, err)
	assert.Equal(t, 42, val)

	_, err = Int(p, "s")
	assert.Error(t, err)
}

func TestIntOpt(t *testing.T) {
	p := map[string]any{"n": float64(42), "i": 7}
	assert.Equal(t, 42, IntOpt(p, "n", 0))
	assert.Equal(t, 7, IntOpt(p, "i", 0))
	assert.Equal(t, 99, IntOpt(p, "missing", 99))
}

func TestFloat(t *testing.T) {
	p := map[string]any{"f": 3.14, "s": "str"}
	val, err := Float(p, "f")
	assert.NoError(t, err)
	assert.Equal(t, 3.14, val)

	_, err = Float(p, "s")
	assert.Error(t, err)
}

func TestFloatOpt(t *testing.T) {
	p := map[string]any{"f": 3.14}
	assert.Equal(t, 3.14, FloatOpt(p, "f", 0))
	assert.Equal(t, 1.0, FloatOpt(p, "missing", 1.0))
}

func TestBool(t *testing.T) {
	p := map[string]any{"b": true, "s": "str"}
	val, err := Bool(p, "b")
	assert.NoError(t, err)
	assert.True(t, val)

	_, err = Bool(p, "s")
	assert.Error(t, err)
}

func TestBoolOpt(t *testing.T) {
	p := map[string]any{"b": true}
	assert.True(t, BoolOpt(p, "b", false))
	assert.True(t, BoolOpt(p, "missing", true))
}

func TestBoolLoose(t *testing.T) {
	p := map[string]any{"b": true, "s": "true", "bad": "nope", "n": 1}
	assert.True(t, BoolLoose(p, "b"))
	assert.True(t, BoolLoose(p, "s"))
	assert.False(t, BoolLoose(p, "bad"))
	assert.False(t, BoolLoose(p, "n"))
	assert.False(t, BoolLoose(p, "missing"))
}

func TestStringSlice(t *testing.T) {
	p := map[string]any{
		"topics":   []any{"a", "b", 3},
		"notSlice": "x",
	}
	assert.Equal(t, []string{"a", "b"}, StringSlice(p, "topics"))
	assert.Nil(t, StringSlice(p, "notSlice"))
	assert.Nil(t, StringSlice(p, "missing"))
}

func TestStringMap(t *testing.T) {
	p := map[string]any{
		"m": map[string]any{"a": "1", "b": "2", "c": 3},
	}
	val, err := StringMap(p, "m")
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"a": "1", "b": "2"}, val)

	_, err = StringMap(p, "missing")
	assert.Error(t, err)
}

func TestStringMapOpt(t *testing.T) {
	p := map[string]any{
		"m": map[string]any{"a": "1"},
	}
	assert.Equal(t, map[string]string{"a": "1"}, StringMapOpt(p, "m"))
	assert.Nil(t, StringMapOpt(p, "missing"))
}

func TestAny(t *testing.T) {
	p := map[string]any{"k": 123}
	val, ok := Any(p, "k")
	assert.True(t, ok)
	assert.Equal(t, 123, val)

	_, ok = Any(p, "missing")
	assert.False(t, ok)
}

func TestAnyMap(t *testing.T) {
	inner := map[string]any{"nested": true}
	p := map[string]any{"m": inner}
	val, ok := AnyMap(p, "m")
	assert.True(t, ok)
	assert.Equal(t, inner, val)

	_, ok = AnyMap(p, "missing")
	assert.False(t, ok)
}

func TestStringAlt(t *testing.T) {
	p := map[string]any{"b": "found"}
	val, ok := StringAlt(p, "a", "b", "c")
	assert.True(t, ok)
	assert.Equal(t, "found", val)

	_, ok = StringAlt(p, "x", "y")
	assert.False(t, ok)
}
