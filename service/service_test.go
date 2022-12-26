package service_test

import (
	"testing"

	"github.com/livingsilver94/backee/service"
)

func TestParseEmptyDocument(t *testing.T) {
	const name = "testName"
	srv, err := service.NewFromYaml(name, []byte("# This is an empty document"))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Name != name {
		t.Fatalf("expected name %s. Found %s", name, srv.Name)
	}
}
