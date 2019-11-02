package createdmg

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCmd(t *testing.T) {
	require := require.New(t)

	cmd, err := Cmd(context.Background())
	defer Close(cmd)
	require.NoError(err)
	require.FileExists(cmd.Path)
	require.FileExists(filepath.Join(cmd.Path, "..", "support", "dmg-license.py"))

	require.NoError(Close(cmd))
	require.NoError(Close(cmd))
}
