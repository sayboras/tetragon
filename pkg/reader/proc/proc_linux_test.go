// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package proc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cilium/tetragon/pkg/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProcStatStrings(t *testing.T) {
	stat := "206305 (zsh( )foo) S 206303 206305 206305 34821 206368 4194304 9687 4455 0 0 56 17 2 0 20 0 1 0 19321046 17514496 1866 18446744073709551615 94273300672512 94273301280581 140729040978832 0 0 0 2 3686400 134295555 1 0 0 17 3 0 0 0 0 0 94273301428976 94273301458280 94273325256704 140729040984354 140729040984358 140729040984358 140729040986095 0"
	statStrings := getProcStatStrings(stat)
	assert.Equal(t, "206305", statStrings[0], "Incorrect first field")
	assert.Equal(t, "(zsh( )foo)", statStrings[1], "Incorrect comm field")
	assert.Equal(t, "S", statStrings[2], "Incorrect third field")
	assert.Equal(t, "206303", statStrings[3], "Incorrect fourth field")
	assert.Equal(t, "140729040986095", statStrings[50], "Incorrect 51st field")
	assert.Equal(t, "0", statStrings[51], "Incorrect 52nd field")
	assert.Len(t, statStrings, 52, "Incorrect number of entries")
}

func TestGetStatus(t *testing.T) {
	self := filepath.Join(option.Config.ProcFS, "self")

	status, err := GetStatus(self)
	require.NoError(t, err)
	assert.NotEmpty(t, status.Uids)
	for i := range status.Uids {
		assert.NotEmpty(t, status.Uids[i])
	}
}

func TestGetPid1Status(t *testing.T) {
	pid1 := filepath.Join(option.Config.ProcFS, "1")

	// Is pid 1 available for reading
	file, err := os.OpenFile(filepath.Join(pid1, "status"), os.O_RDONLY, 0444)
	if err != nil {
		t.Skipf("Skipping test %s failed to open %s/status: %v", t.Name(), pid1, err)
	}
	file.Close()

	status, err := GetStatus(pid1)
	require.NoError(t, err)
	assert.NotEmpty(t, status.Uids)

	uids, err := status.GetUids()
	require.NoError(t, err)
	assert.Equal(t, uint32(0), uids[0])
	assert.Equal(t, uint32(0), uids[1])

	gids, err := status.GetGids()
	require.NoError(t, err)
	assert.Equal(t, uint32(0), gids[0])
	assert.Equal(t, uint32(0), gids[1])

	// PID 1 does not have a loginuid
	loginuid, err := status.GetLoginUid()
	require.NoError(t, err)
	assert.Equal(t, uint32(4294967295), loginuid)
}
