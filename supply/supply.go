package supply

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	// AddBinDependencyLink(string, string) error
	BuildDir() string
	DepDir() string
}

type Manifest interface {
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	InstallOnlyVersion(string, string) error
}

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying rust")

	if err := s.Installer.InstallOnlyVersion("rust", s.Stager.DepDir()); err != nil {
		return err
	}

	depDir := s.Stager.DepDir()
	srcBaseDir, err := singleDirGlob(filepath.Join(depDir, "rust-*"))
	if err != nil {
		return err
	}
	defer os.RemoveAll(srcBaseDir)

	if err := os.RemoveAll(filepath.Join(depDir, "bin")); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(depDir, "lib")); err != nil {
		return err
	}

	if err := mvGlobContents(filepath.Join(srcBaseDir, "rustc"), depDir); err != nil {
		return err
	}
	if err := mvGlobContents(filepath.Join(srcBaseDir, "cargo", "bin"), filepath.Join(depDir, "bin")); err != nil {
		return err
	}
	if err := mvGlobContents(filepath.Join(srcBaseDir, "rust-std-*/lib/rustlib/*/lib"), filepath.Join(depDir, "lib/rustlib/x86_64-*/lib/")); err != nil {
		return err
	}

	return nil
}

func mvGlobContents(src, dest string) error {
	src, err := singleDirGlob(src)
	if err != nil {
		return err
	}
	dest, err = singleDirGlob(dest)
	if err != nil {
		return err
	}

	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := os.Rename(filepath.Join(src, file.Name()), filepath.Join(dest, file.Name())); err != nil {
			return err
		}
	}
	return nil
}

func singleDirGlob(dir string) (string, error) {
	files, err := filepath.Glob(dir)
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("expected 1 match for '%s', found %d", dir, len(files))
	}
	return files[0], nil
}
