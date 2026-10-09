package ipc

import "testing"

func TestRequestGet(t *testing.T) {
	req := Request{Params: map[string]any{"name": "test", "count": 42, "enabled": true}}

	if name, ok := req.Get[string]("name"); !ok || name != "test" {
		t.Errorf("Get[string] = %q, %v; want test, true", name, ok)
	}
	if count, ok := req.Get[int]("count"); !ok || count != 42 {
		t.Errorf("Get[int] = %d, %v; want 42, true", count, ok)
	}
	if _, ok := req.Get[string]("missing"); ok {
		t.Error("missing key should report false")
	}
	if _, ok := req.Get[int]("name"); ok {
		t.Error("wrong type should report false")
	}
}

func TestRequestGetOr(t *testing.T) {
	req := Request{Params: map[string]any{"name": "test", "enabled": true}}

	if v := req.GetOr("name", "default"); v != "test" {
		t.Errorf("GetOr existing = %q; want test", v)
	}
	if v := req.GetOr("missing", "default"); v != "default" {
		t.Errorf("GetOr missing = %q; want default", v)
	}
	if v := req.GetOr("name", 0); v != 0 {
		t.Errorf("GetOr wrong type = %d; want 0", v)
	}
}
