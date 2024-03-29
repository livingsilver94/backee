// SPDX-FileCopyrightText: Fabio Forni <development@redaril.me>
// SPDX-License-Identifier: MPL-2.0

package repo_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/livingsilver94/backee/repo"
	"github.com/livingsilver94/backee/service"
)

func TestAddParentNoService(t *testing.T) {
	cache := createVariables("key", "val")
	cache.Insert("parent", "parentKey", service.VarValue{Kind: service.ClearText, Value: "parentValue"})
	err := cache.AddParent("not a service", "parent")
	if !errors.Is(err, repo.ErrNoService) {
		t.Fatalf("expected %v error. Got %v", repo.ErrNoService, err)
	}
}

func TestAddParentNoParent(t *testing.T) {
	cache := createVariables("key", "val")
	cache.Insert("parent", "parentKey", service.VarValue{Kind: service.ClearText, Value: "parentValue"})
	err := cache.AddParent(serviceName, "not a parent")
	if !errors.Is(err, repo.ErrNoService) {
		t.Fatalf("expected %v error. Got %v", repo.ErrNoService, err)
	}
}

func TestInsertClearText(t *testing.T) {
	cache := createVariables("key", "val")
	if cache.Length() != 1 {
		t.Fatalf("expected length %d. Got %d", 1, cache.Length())
	}
}

func TestInsertClearTextTwice(t *testing.T) {
	cache := createVariables("key", "value", "key", "boo!")
	if cache.Length() != 1 {
		t.Fatalf("expected length %d. Got %d", 1, cache.Length())
	}
	if val, _ := cache.Get(serviceName, "key"); val != "value" {
		t.Fatalf("expected value  %q. Got %q", "value", val)
	}
}

type testVarStore struct{}

func (testVarStore) Value(key string) (value string, err error) {
	return "testy" + key, nil
}

func TestInsertVarStore(t *testing.T) {
	const kind service.VarKind = "testKind"

	cache := repo.NewVariables()
	cache.RegisterSolver(kind, testVarStore{})
	err := cache.Insert(serviceName, "key", service.VarValue{Kind: kind, Value: "storeValue"})
	if err != nil {
		t.Fatal(err)
	}
	v, _ := cache.Get(serviceName, "key")
	if v != "testystoreValue" {
		t.Fatalf("expected value %q. Got %q", "testystoreValue", v)
	}
}

func TestGet(t *testing.T) {
	cache := createVariables("key", "value")
	value, ok := cache.Get(serviceName, "key")
	if ok != nil {
		t.Fatalf("OK value should be nil")
	}
	if value != "value" {
		t.Fatalf("expected value %q. Got %q", "value", value)
	}
}

func TestGetNoService(t *testing.T) {
	cache := repo.NewVariables()
	_, err := cache.Get(serviceName, "key")
	if !errors.Is(err, repo.ErrNoService) {
		t.Fatalf("expected error %v. Got %v", repo.ErrNoService, err)
	}
}

func TestGetNoVariable(t *testing.T) {
	cache := createVariables("key", "value")
	_, err := cache.Get(serviceName, "absent key")
	if !errors.Is(err, repo.ErrNoVariable) {
		t.Fatalf("expected error %v. Got %v", repo.ErrNoVariable, err)
	}
}

func TestParents(t *testing.T) {
	const parentName = "parentName"

	cache := createVariables("key", "val")
	cache.Insert(parentName, "parentKey", service.VarValue{Kind: service.ClearText, Value: "parentValue"})
	err := cache.AddParent(serviceName, parentName)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	obtained, err := cache.Parents(serviceName)
	if err != nil {
		t.Fatalf("expected nil error. Got %v", err)
	}
	expected := []string{parentName}
	if !reflect.DeepEqual(obtained, expected) {
		t.Fatalf("expected parent list %v. Got %v", expected, obtained)
	}
}

func TestGobCodec(t *testing.T) {
	expected := createVariables("key", "val")
	expected.Common = map[string]string{
		"common1:": "value1",
	}

	data, err := expected.GobEncode()
	if err != nil {
		t.Fatal(err)
	}
	var result repo.Variables
	err = result.GobDecode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("expected decoded variables %#v. Got %#v", expected, result)
	}
}

const serviceName = "service1"

func createVariables(keyVal ...string) repo.Variables {
	if len(keyVal)%2 != 0 {
		panic("keys and values must be pairs")
	}
	v := repo.NewVariables()
	for i := 0; i < len(keyVal)-1; i += 2 {
		v.Insert(serviceName, keyVal[i], service.VarValue{Kind: service.ClearText, Value: keyVal[i+1]})
	}
	return v
}
