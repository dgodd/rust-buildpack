package supply

import (
	"io"

	"github.com/cloudfoundry/libbuildpack"
)

type Stager interface {
	AddBinDependencyLink(string, string) error
	DepDir() string
}

type Manifest interface {
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	InstallDependency(libbuildpack.Dependency, string) error
}

type Command interface{}

type Supplier struct {
	Manifest  Manifest
	Installer Installer
	Stager    Stager
	Command   Command
	Log       *libbuildpack.Logger
}

func (s *Supplier) Run() error {
	s.Log.BeginStep("Supplying rust")

	dep, err := s.Manifest.DefaultVersion("rust")
	if err != nil {
		return err
	}
	if err := s.Installer.InstallDependency(dep, s.Stager.DepDir()); err != nil {
		return err
	}

	// Create bin and lib links
	for _, dirType := range []string{"bin", "lib"} {
		files, err := filepath.Glob(filepath.Join(s.Stager.DepDir(), "rust-*", "*", dirType, "*")
		if err != nil {
			return err
		}
		for _, file := range files {
			if fi, err := os.Stat(file); err != nil {
				return err
			} else if fi.IsDir() {
				continue
			}

			fmt.Println("DG: DEBUG: addDependencyLink:", file, filepath.Join(s.Stager.DepDir(), dirType, filepath.Base(file)))
			if err := addDependencyLink(file, filepath.Join(s.Stager.DepDir(), dirType, filepath.Base(file))); err != nil {
				return err
			}
		}
	}

	return nil
}

func addDependencyLink(dest, src string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	relPath, err := filepath.Rel(filepath.Dir(src), dest)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		return os.Link(relPath, filepath.Join(binDir, sourceName))
	} else {
		return os.Symlink(relPath, filepath.Join(binDir, sourceName))
	}
}
