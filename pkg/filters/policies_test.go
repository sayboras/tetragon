// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package filters

import (
	"context"
	"testing"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/event"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyNamesFilterInvalidEvent(t *testing.T) {
	ctx := context.Background()
	filters := []*tetragon.Filter{{PolicyNames: []string{"red-policy"}}}
	filterFuncs := []OnBuildFilter{&PolicyNamesFilter{}}
	fs, err := BuildFilterList(ctx, filters, filterFuncs)
	require.NoError(t, err)

	events := eventsWithPolicyName("")
	for _, ev := range events {
		assert.False(t, fs.MatchOne(&ev))
	}
}

func TestPolicyNamesFilterCorrectValue(t *testing.T) {
	ctx := context.Background()
	filters := []*tetragon.Filter{{PolicyNames: []string{"red-policy", "blue-policy"}}}
	filterFuncs := []OnBuildFilter{&PolicyNamesFilter{}}
	fs, err := BuildFilterList(ctx, filters, filterFuncs)
	require.NoError(t, err)

	testCases := []struct {
		policyName string
		match      bool
	}{
		{"red-policy", true},
		{"blue-policy", true},
		{"yellow-policy", false},
	}

	for _, tc := range testCases {
		events := eventsWithPolicyName(tc.policyName)
		for _, ev := range events {
			assert.Equal(t, tc.match, fs.MatchOne(&ev))
		}
	}
}

func TestPolicyNamesFilterEmptyValue(t *testing.T) {
	ctx := context.Background()
	filters := []*tetragon.Filter{{PolicyNames: []string{""}}}
	filterFuncs := []OnBuildFilter{&PolicyNamesFilter{}}
	fs, err := BuildFilterList(ctx, filters, filterFuncs)
	require.NoError(t, err)
	// empty selector matches nothing
	events := eventsWithPolicyName("red-policy")
	for _, ev := range events {
		assert.False(t, fs.MatchOne(&ev))
	}
}

func TestPolicyNamesFilterNilValue(t *testing.T) {
	ctx := context.Background()
	filters := []*tetragon.Filter{{PolicyNames: nil}}
	filterFuncs := []OnBuildFilter{&PolicyNamesFilter{}}
	fs, err := BuildFilterList(ctx, filters, filterFuncs)
	require.NoError(t, err)
	// nil selector matches everything, i.e., does not filter events
	events := eventsWithPolicyName("red-policy")
	for _, ev := range events {
		assert.True(t, fs.MatchOne(&ev))
	}
}

// eventsWithPolicyName generates kprobe, tracepoint, uprobe, and lsm events
// with the specified policy name.
func eventsWithPolicyName(policyName string) []event.Event {
	return []event.Event{
		{
			Event: &tetragon.GetEventsResponse{
				Event: &tetragon.GetEventsResponse_ProcessKprobe{
					ProcessKprobe: &tetragon.ProcessKprobe{
						PolicyName: policyName,
					},
				},
			},
		},
		{
			Event: &tetragon.GetEventsResponse{
				Event: &tetragon.GetEventsResponse_ProcessTracepoint{
					ProcessTracepoint: &tetragon.ProcessTracepoint{
						PolicyName: policyName,
					},
				},
			},
		},
		{
			Event: &tetragon.GetEventsResponse{
				Event: &tetragon.GetEventsResponse_ProcessUprobe{
					ProcessUprobe: &tetragon.ProcessUprobe{
						PolicyName: policyName,
					},
				},
			},
		},
		{
			Event: &tetragon.GetEventsResponse{
				Event: &tetragon.GetEventsResponse_ProcessLsm{
					ProcessLsm: &tetragon.ProcessLsm{
						PolicyName: policyName,
					},
				},
			},
		},
	}
}
