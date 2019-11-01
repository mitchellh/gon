package sign

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// childEnv is the env var that must be set to trigger a child command.
const childEnv = "GON_TEST_CHILD"

// childCommands is the list of commands we support
var childCommands = map[string]func() int{
	"success": childSuccess,
}

// childCmd is used to create a command that executes a command in the
// childCommands map in a new process.
func childCmd(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()

	// Get the path to our executable
	selfPath, err := filepath.Abs(os.Args[0])
	if err != nil {
		t.Fatalf("error creating child command: %s", err)
		return nil
	}

	cmd := exec.Command(selfPath, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, childEnv+"="+name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

func childSuccess() int {
	println("success")
	return 0
}
