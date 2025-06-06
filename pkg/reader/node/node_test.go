// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package node

import (
	"os"
	"testing"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNodeNameForExport(t *testing.T) {
	assert.NotEmpty(t, GetNodeNameForExport()) // we should get the hostname here
	require.NoError(t, os.Setenv("NODE_NAME", "from-node-name"))
	SetExportNodeName()
	assert.Equal(t, "from-node-name", GetNodeNameForExport())
	require.NoError(t, os.Setenv("HUBBLE_NODE_NAME", "from-hubble-node-name"))
	SetExportNodeName()
	assert.Equal(t, "from-hubble-node-name", GetNodeNameForExport())
	require.NoError(t, os.Unsetenv("NODE_NAME"))
	require.NoError(t, os.Unsetenv("HUBBLE_NODE_NAME"))
}

func TestSetCommonFields(t *testing.T) {
	ev := tetragon.GetEventsResponse{}
	assert.Empty(t, ev.NodeName)
	assert.Empty(t, ev.ClusterName)
	nodeName := "my-node-name"
	require.NoError(t, os.Setenv("NODE_NAME", nodeName))
	SetExportNodeName()
	option.Config.ClusterName = "my-cluster-name"
	SetCommonFields(&ev)
	assert.Equal(t, nodeName, ev.GetNodeName())
	assert.Equal(t, option.Config.ClusterName, ev.GetClusterName())
	require.NoError(t, os.Unsetenv("NODE_NAME"))
}

func TestGetKubernetesNodeName(t *testing.T) {
	assert.NotEmpty(t, GetNodeName()) // we should get the hostname here
	require.NoError(t, os.Setenv("NODE_NAME", "from-node-name"))
	SetNodeName()
	assert.Equal(t, "from-node-name", GetNodeName())
	require.NoError(t, os.Setenv("HUBBLE_NODE_NAME", "from-hubble-node-name"))
	SetNodeName()
	assert.Equal(t, "from-node-name", GetNodeName())
	require.NoError(t, os.Unsetenv("NODE_NAME"))
	require.NoError(t, os.Unsetenv("HUBBLE_NODE_NAME"))
}
