package installer_test

import (
	"testing"

	"github.com/livingsilver94/backee/installer"
	"github.com/livingsilver94/backee/service"
)

func TestTemplateServiceVar(t *testing.T) {
	vars := installer.NewVariables()
	vars.InsertMany("service", map[string]service.VarValue{
		"var1": service.VarValue{Kind: service.ClearText, Value: "value1"},
		"var2": service.VarValue{Kind: service.ClearText, Value: "value2"},
	})
}
