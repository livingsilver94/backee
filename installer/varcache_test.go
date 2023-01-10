package installer_test

import (
	"testing"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

func TestNew(t *testing.T) {
	cache := installer.NewVarCache()
	if cache == nil {
		t.Fatal("returned value is nil")
	}
}

func TestInsertClear(t *testing.T) {
	cache := installer.NewVarCache()
	cache.Insert("service1", "key", service.VarValue{Kind: service.ClearText, Value: "value"})
	if len(cache) != 1 {
		t.Fatalf("expected length %d. Got %d", 1, len(cache))
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
