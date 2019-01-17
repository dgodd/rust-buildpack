package finalize

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/pkg/errors"
)

type Stager interface {
	BuildDir() string
	CacheDir() string
}

type Manifest interface {
}

type Installer interface {
}

type Command interface {
	Execute(string, io.Writer, io.Writer, string, ...string) error
	Run(*exec.Cmd) error
	// Output(dir string, program string, args ...string) (string, error)
}

type Finalizer struct {
	Manifest Manifest
	Stager   Stager
	Command  Command
	Log      *libbuildpack.Logger
}

func (f *Finalizer) Run() error {
	f.Log.BeginStep("Cargo Build")
	cmd := exec.Command("cargo", "build", "--release")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = f.Stager.BuildDir()
	cmd.Env = append(os.Environ(), "CARGO_TARGET_DIR="+filepath.Join(f.Stager.CacheDir(), "cargo_target"))
	if err := f.Command.Run(cmd); err != nil {
		return err
	}

	// Copy target to build dir
	if err := os.RemoveAll(filepath.Join(f.Stager.BuildDir(), "target")); err != nil {
		return err
	}
	if err := f.Command.Execute(f.Stager.BuildDir(), os.Stdout, os.Stderr, "cp", "-r", filepath.Join(f.Stager.CacheDir(), "cargo_target"), filepath.Join(f.Stager.BuildDir(), "target")); err != nil {
		return err
	}

	f.Log.BeginStep("Configuring rust")
	return f.WriteReleaseYAML()
}

func (f *Finalizer) WriteReleaseYAML() error {
	var cargoToml struct {
		Package struct {
			Name string `toml:"name"`
		} `toml:"package"`
	}
	if _, err := toml.DecodeFile(filepath.Join(f.Stager.BuildDir(), "Cargo.toml"), &cargoToml); err != nil {
		return errors.Wrap(err, "Must provide package/name in Cargo.toml")
	}
	data := map[string]map[string]string{
		"default_process_types": map[string]string{
			"web": filepath.Join("./target/release", cargoToml.Package.Name),
		},
	}
	releasePath := "/tmp/rust-buildpack-release-step.yml"
	if err := libbuildpack.NewYAML().Write(releasePath, data); err != nil {
		f.Log.Error("Error writing release YAML: %v", err)
		return err
	}

	return nil
}
