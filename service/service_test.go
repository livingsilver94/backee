package service_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/livingsilver94/backee/service"
)

const name = "testName"

func TestParseEmptyDocument(t *testing.T) {
	srv, err := service.NewFromYAML(name, []byte("# This is an empty document"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv, service.New(name)) {
		t.Fatalf("expected empty service %#v. Got %#v", service.New(name), srv)
	}
}

func TestParseEmptyDocumentReader(t *testing.T) {
	srv, err := service.NewFromYAMLReader(name, strings.NewReader("# This is an empty document"))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Name != name {
		t.Fatalf("expected name %s. Found %s", name, srv.Name)
	}
}

func TestParseDepends(t *testing.T) {
	expect := service.NewDepSetFrom([]string{"service1", "service2"})
	const doc = `
depends:
  - service1
  - service2
  - service1`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !srv.Depends.Equal(expect) {
		t.Fatalf("expected dependencies %v. Found %v", expect, srv.Depends)
	}
}

func TestParseSetup(t *testing.T) {
	const expect = "echo \"Test!\"\n# Another line."
	const doc = `
setup: |
  echo "Test!"
  # Another line.`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Setup == nil {
		t.Fatal("nil value")
	}
	if *srv.Setup != expect {
		t.Fatalf("expected setup %q. Found %q", expect, *srv.Setup)
	}
}

func TestParsePkgManager(t *testing.T) {
	expect := []string{"sudo", "apt-get", "install", "-y"}
	const doc = `
pkgmanager: ["sudo", "apt-get", "install", "-y"]`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.PkgManager, expect) {
		t.Fatalf("expected pkgmanager %v. Found %v", expect, srv.PkgManager)
	}
}

func TestParsePackages(t *testing.T) {
	expect := []string{"nano", "micro", "zsh"}
	const doc = `
packages:
  - nano
  - micro
  - zsh`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Packages, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Packages)
	}
}

func TestParseLinks(t *testing.T) {
	expect := map[string]service.FilePath{
		"/my/path/file1": {Path: "/tmp/alias1", Mode: 0o000},
		"my/path/file2":  {Path: "/tmp/alias2", Mode: 0o755},
	}
	const doc = `
links:
  /my/path/file1:
    path: /tmp/alias1
    mode: 0o000
  my/path/file2:
    path: /tmp/alias2
    mode: 0o755`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Links, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Links)
	}
}

func TestParseLinksString(t *testing.T) {
	expect := map[string]service.FilePath{
		"/my/path/file1": {Path: "/tmp/alias1", Mode: 0o000},
		"my/path/file2":  {Path: "/tmp/alias2", Mode: 0o000},
	}
	const doc = `
links:
  /my/path/file1: /tmp/alias1
  my/path/file2: /tmp/alias2`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Links, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Links)
	}
}

func TestParseVariables(t *testing.T) {
	expect := map[string]service.VarValue{
		"username":         {Kind: service.ClearText, Value: "value1"},
		"password":         {Kind: service.VarKind("keepassxc"), Value: "dbKey"},
		"implicitKind":     {Kind: service.ClearText, Value: "value2"},
		"scalar":           {Kind: service.ClearText, Value: "value3"},
		service.VarDatadir: {Kind: service.Datadir, Value: name},
	}
	const doc = `
variables:
  username:
    kind: cleartext
    value: value1
  password:
    kind: keepassxc
    value: dbKey
  implicitKind:
    value: value2
  scalar: value3`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Variables, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Variables)
	}
}

func TestParseVariablesString(t *testing.T) {
	expect := map[string]service.VarValue{
		"username":         {Kind: service.ClearText, Value: "value1"},
		service.VarDatadir: {Kind: service.Datadir, Value: name},
	}
	const doc = `
variables:
  username: value1`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Variables, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Variables)
	}
}

func TestParseCopies(t *testing.T) {
	expect := map[string]service.FilePath{
		"nginx.conf": {Path: "/etc/nginx/nginx.conf", Mode: 0o000},
		"config":     {Path: "${HOME}/.ssh/config", Mode: 0o600},
	}
	const doc = `
copies:
  nginx.conf:
    path: /etc/nginx/nginx.conf
    mode: 0o000
  config:
    path: ${HOME}/.ssh/config
    mode: 0o600`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Copies, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Copies)
	}
}

func TestParseCopiesString(t *testing.T) {
	expect := map[string]service.FilePath{
		"/my/path/file1": {Path: "/tmp/alias1", Mode: 0o000},
		"my/path/file2":  {Path: "/tmp/alias2", Mode: 0o000},
	}
	const doc = `
copies:
  /my/path/file1: /tmp/alias1
  my/path/file2: /tmp/alias2`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(srv.Copies, expect) {
		t.Fatalf("expected packages %v. Found %v", expect, srv.Copies)
	}
}

func TestParseFinalize(t *testing.T) {
	const expect = "echo \"Test!\"\n# Another line."
	const doc = `
finalize: |
  echo "Test!"
  # Another line.`
	srv, err := service.NewFromYAML(name, []byte(doc))
	if err != nil {
		t.Fatal(err)
	}
	if srv.Finalize == nil {
		t.Fatal("nil value")
	}
	if *srv.Finalize != expect {
		t.Fatalf("expected finalize script %q. Found %q", expect, *srv.Finalize)
	}
}

func TestServiceHash(t *testing.T) {
	srv := service.Service{Name: "myName"}
	expected := srv.Name
	obtained := srv.Hash()
	if obtained != expected {
		t.Fatalf("expected %s. Got %s", expected, obtained)
	}
}
