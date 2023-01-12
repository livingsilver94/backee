package installer_test

import (
	"testing"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

func TestInsertClear(t *testing.T) {
	cache := installer.NewVarCache()
	cache.Insert("service1", "key", service.VarValue{Kind: service.ClearText, Value: "value"})
	if cache.Length() != 1 {
		t.Fatalf("expected length %d. Got %d", 1, cache.Length())
	}
}

func TestInsertTwice(t *testing.T) {
	cache := installer.NewVarCache()
	cache.Insert("service1", "key", service.VarValue{Kind: service.ClearText, Value: "value"})
	cache.Insert("service1", "key", service.VarValue{Kind: service.ClearText, Value: "boo!"})
	if cache.Length() != 1 {
		t.Fatalf("expected length %d. Got %d", 1, cache.Length())
	}
	if val, _ := cache.Get("service1", "key"); val != "value" {
		t.Fatalf("expected value  %q. Got %q", "value", val)
	}
}

type testVarStore struct{}

func (testVarStore) Value(key string) (value string, err error) {
	return "testy" + key, nil
}

func TestInsertStore(t *testing.T) {
	const kind service.VarKind = "testKind"

	cache := installer.NewVarCache()
	cache.SetStore(kind, testVarStore{})
	err := cache.Insert("service1", "key", service.VarValue{Kind: kind, Value: "storeValue"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := cache.Get("service1", "key")
	if v != "testystoreValue" {
		t.Fatalf("expected value %q. Got %q", "testystoreValue", v)
	}
}

func TestGet(t *testing.T) {
	cache := installer.NewVarCache()
	cache.Insert("service1", "key", service.VarValue{Kind: service.ClearText, Value: "value"})
	value, ok := cache.Get("service1", "key")
	if !ok {
		t.Fatalf("OK value should be true")
	}
	if value != "value" {
		t.Fatalf("expected value %q. Got %q", "value", value)
	}
}
