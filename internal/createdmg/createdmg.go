package createdmg

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/gon/internal/createdmg/bindata"
)

// Cmd returns an *exec.Cmd that has the Path prepopulated to execute the
// create-dmg script. You MUST call Close on this command when you're done.
func Cmd(ctx context.Context) (*exec.Cmd, error) {
	// Create a temporary directory where we'll extract the project
	td, err := ioutil.TempDir("", "createdmg")
	if err != nil {
		return nil, err
	}

	// Extract the create-dmg project
	if err := bindata.RestoreAssets(td, ""); err != nil {
		os.RemoveAll(td)
		return nil, err
	}

	// Create a command
	return exec.CommandContext(ctx, filepath.Join(td, "create-dmg")), nil
}

// Close cleans up the temporary resources associated with the command.
// This Cmd should've been returned by Cmd otherwise we may delete unrelated
// data.
func Close(cmd *exec.Cmd) error {
	// Protect against unset commands
	if cmd == nil || cmd.Path == "" || filepath.Base(cmd.Path) == cmd.Path {
		return nil
	}

	return os.RemoveAll(filepath.Dir(cmd.Path))
}
