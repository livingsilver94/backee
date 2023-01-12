package installer

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/go-logr/logr"
	"github.com/livingsilver94/backee/service"
)

type Performer func(logr.Logger, *service.Service) error

var (
	Setup Performer = func(log logr.Logger, srv *service.Service) error {
		if srv.Setup == nil || *srv.Setup == "" {
			return nil
		}
		log.Info("Running setup script")
		return runScript(*srv.Setup)
	}

	PackageInstaller Performer = func(log logr.Logger, srv *service.Service) error {
		if len(srv.Packages) == 0 {
			return nil
		}
		log.Info("Installing OS packages")
		args := make([]string, 0, len(srv.PkgManager[1:])+len(srv.Packages))
		args = append(args, srv.PkgManager[1:]...)
		args = append(args, srv.Packages...)
		return runProcess(srv.PkgManager[0], args...)
	}
)

func SymlinkPerformer(repo Repository) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if len(srv.Links) == 0 {
			return nil
		}
		log.Info("Symlinking files")
		linkDir, err := repo.LinkDir(srv.Name)
		if err != nil {
			return err
		}
		wr := symlinkWriter{log: log}
		return writeFilePaths(srv.Links, linkDir, wr)
	}
}

func CopyPerformer(repo Repository, vars VarCache) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if len(srv.Copies) == 0 {
			return nil
		}
		log.Info("Copying files")
		dataDir, err := repo.DataDir(srv.Name)
		if err != nil {
			return err
		}
		vars.Insert(srv.Name, service.VarDatadir, service.VarValue{Kind: service.ClearText, Value: dataDir})
		vars.InsertMany(srv.Name, srv.Variables)
		wr := copyWriter{variables: vars.GetAll(srv.Name)}
		return writeFilePaths(srv.Copies, dataDir, wr)
	}
}

func Finalizer(repo Repository, vars VarCache) Performer {
	return func(log logr.Logger, srv *service.Service) error {
		if srv.Finalize == nil || *srv.Finalize == "" {
			return nil
		}
		log.Info("Running finalizer script")
		if _, ok := vars.Get(srv.Name, service.VarDatadir); !ok {
			datadir, err := repo.DataDir(srv.Name)
			if err != nil {
				return err
			}
			vars.Insert(srv.Name, service.VarDatadir, service.VarValue{Kind: service.ClearText, Value: datadir})
		}
		vars.InsertMany(srv.Name, srv.Variables)
		tmpl, err := template.New("finalizer").Parse(*srv.Finalize)
		if err != nil {
			return err
		}
		var script strings.Builder
		err = tmpl.Execute(&script, vars.GetAll(srv.Name))
		if err != nil {
			return err
		}
		return runScript(script.String())
	}
}

type fileWriter interface {
	writeFile(dstPath, srcPath string) error
}

func writeFilePaths(paths map[string]service.FilePath, srcBase string, writer fileWriter) error {
	for srcFile, param := range paths {
		srcPath := filepath.Join(srcBase, srcFile)
		dstPath := ReplaceEnvVars(param.Path)

		dstDir := filepath.Dir(dstPath)
		id, err := parentPathOwner(dstDir)
		if err != nil {
			return err
		}
		err = RunAsUnixID(func() error {
			err := os.MkdirAll(dstDir, 0755)
			if err != nil {
				return err
			}
			err = writer.writeFile(dstPath, srcPath)
			if err != nil {
				return err
			}
			if param.Mode != 0 {
				err := os.Chmod(dstPath, fs.FileMode(param.Mode))
				if err != nil {
					return err
				}
			}
			return nil
		}, id)
		if err != nil {
			return err
		}
	}
	return nil
}

type symlinkWriter struct {
	log logr.Logger
}

func (w symlinkWriter) writeFile(dstPath, srcPath string) error {
	err := os.Symlink(srcPath, dstPath)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}
		w.log.Info("already exists", dstPath)
	}
	return nil
}

type copyWriter struct {
	variables map[string]string
}

func (w copyWriter) writeFile(dstPath, srcPath string) error {
	tmp, err := template.ParseFiles(srcPath)
	if err != nil {
		return err
	}
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	buff := bufio.NewWriter(dstFile)
	err = tmp.Execute(buff, w.variables)
	if err != nil {
		return err
	}
	return buff.Flush()
}

const (
	envVarPrefix  = "${"
	envVarPattern = "[a-zA-Z_]\\w+"
	envVarSuffix  = "}"
)

var envVarRegex *regexp.Regexp

func init() {
	envVarRegex = regexp.MustCompile(regexp.QuoteMeta(envVarPrefix) + envVarPattern + regexp.QuoteMeta(envVarSuffix))
}

func ReplaceEnvVars(src string) string {
	replacer := func(match string) string {
		varName := strings.TrimPrefix(match, envVarPrefix)
		varName = strings.TrimSuffix(varName, envVarSuffix)
		return os.Getenv(varName)
	}
	return envVarRegex.ReplaceAllStringFunc(src, replacer)
}

func runProcess(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = nil
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type UnixID struct {
	UID uint32
	GID uint32
}

func PathOwner(path string) (UnixID, error) {
	return PathOwnerFS(os.DirFS(path), ".")
}

func parentPathOwner(path string) (UnixID, error) {
	var id *UnixID
	for id == nil {
		if len(path) == 1 {
			return UnixID{}, fmt.Errorf("parent directory of %s: %w", path, fs.ErrNotExist)
		}
		var err error
		*id, err = PathOwner(path)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return UnixID{}, err
			}
			path = filepath.Dir(path)
		}
	}
	return *id, nil
}
