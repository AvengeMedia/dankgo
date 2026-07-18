package dbusutil

import (
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func TestAs(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		v := dbus.MakeVariant("hello")
		val, ok := As[string](v)
		assert.True(t, ok)
		assert.Equal(t, "hello", val)
	})

	t.Run("bool", func(t *testing.T) {
		v := dbus.MakeVariant(true)
		val, ok := As[bool](v)
		assert.True(t, ok)
		assert.True(t, val)
	})

	t.Run("int32", func(t *testing.T) {
		v := dbus.MakeVariant(int32(42))
		val, ok := As[int32](v)
		assert.True(t, ok)
		assert.Equal(t, int32(42), val)
	})

	t.Run("wrong type", func(t *testing.T) {
		v := dbus.MakeVariant("hello")
		_, ok := As[int](v)
		assert.False(t, ok)
	})
}

func TestAsOr(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		v := dbus.MakeVariant("hello")
		val := AsOr(v, "default")
		assert.Equal(t, "hello", val)
	})

	t.Run("wrong type uses default", func(t *testing.T) {
		v := dbus.MakeVariant(123)
		val := AsOr(v, "default")
		assert.Equal(t, "default", val)
	})
}

func TestGet(t *testing.T) {
	m := map[string]dbus.Variant{
		"name":    dbus.MakeVariant("test"),
		"enabled": dbus.MakeVariant(true),
		"count":   dbus.MakeVariant(int32(5)),
	}

	t.Run("exists", func(t *testing.T) {
		val, ok := Get[string](m, "name")
		assert.True(t, ok)
		assert.Equal(t, "test", val)
	})

	t.Run("missing key", func(t *testing.T) {
		_, ok := Get[string](m, "missing")
		assert.False(t, ok)
	})

	t.Run("wrong type", func(t *testing.T) {
		_, ok := Get[int](m, "name")
		assert.False(t, ok)
	})
}

func TestGetOr(t *testing.T) {
	m := map[string]dbus.Variant{
		"name": dbus.MakeVariant("test"),
	}

	t.Run("exists", func(t *testing.T) {
		val := GetOr(m, "name", "default")
		assert.Equal(t, "test", val)
	})

	t.Run("missing uses default", func(t *testing.T) {
		val := GetOr(m, "missing", "default")
		assert.Equal(t, "default", val)
	})

	t.Run("wrong type uses default", func(t *testing.T) {
		val := GetOr(m, "name", 42)
		assert.Equal(t, 42, val)
	})
}

func TestNormalize(t *testing.T) {
	t.Run("variant unwrap", func(t *testing.T) {
		v := dbus.MakeVariant("hello")
		result := Normalize(v)
		assert.Equal(t, "hello", result)
	})

	t.Run("nested variant", func(t *testing.T) {
		v := dbus.MakeVariant(dbus.MakeVariant("nested"))
		result := Normalize(v)
		assert.Equal(t, "nested", result)
	})

	t.Run("object path", func(t *testing.T) {
		v := dbus.ObjectPath("/org/test")
		result := Normalize(v)
		assert.Equal(t, "/org/test", result)
	})

	t.Run("object path slice", func(t *testing.T) {
		v := []dbus.ObjectPath{"/org/a", "/org/b"}
		result := Normalize(v)
		assert.Equal(t, []string{"/org/a", "/org/b"}, result)
	})

	t.Run("variant map", func(t *testing.T) {
		v := map[string]dbus.Variant{
			"key": dbus.MakeVariant("value"),
		}
		result := Normalize(v)
		expected := map[string]any{"key": "value"}
		assert.Equal(t, expected, result)
	})

	t.Run("any slice", func(t *testing.T) {
		v := []any{dbus.MakeVariant("a"), dbus.ObjectPath("/b")}
		result := Normalize(v)
		expected := []any{"a", "/b"}
		assert.Equal(t, expected, result)
	})

	t.Run("passthrough primitives", func(t *testing.T) {
		assert.Equal(t, "hello", Normalize("hello"))
		assert.Equal(t, 42, Normalize(42))
		assert.Equal(t, true, Normalize(true))
	})
}

func TestNormalizeSlice(t *testing.T) {
	input := []any{
		dbus.MakeVariant("a"),
		dbus.ObjectPath("/b"),
		"c",
	}
	result := NormalizeSlice(input)
	expected := []any{"a", "/b", "c"}
	assert.Equal(t, expected, result)
}
