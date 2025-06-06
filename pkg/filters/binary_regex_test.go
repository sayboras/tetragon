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

func TestBinaryRegexFilterBasic(t *testing.T) {
	f := []*tetragon.Filter{{BinaryRegex: []string{"iptable", "systemd"}}}
	fl, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&BinaryRegexFilter{}})
	require.NoError(t, err)
	ev := event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{Binary: "/sbin/iptables-restore"},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/lib/systemd/systemd",
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/lib/systemd/systemd-journald",
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "kube-proxy",
					},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
}

func TestBinaryRegexFilterAdvanced(t *testing.T) {
	f := []*tetragon.Filter{{BinaryRegex: []string{"/usr/sbin/.*", "^/usr/lib/systemd/systemd$"}}}
	fl, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&BinaryRegexFilter{}})
	require.NoError(t, err)
	ev := event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/sbin/dnsmasq",
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/sbin/logrotate",
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/lib/systemd/systemd",
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{
						Binary: "/usr/lib/systemd/systemd-logind",
					},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
}

func TestBinaryRegexFilterInvalidRegex(t *testing.T) {
	f := []*tetragon.Filter{{BinaryRegex: []string{"*"}}}
	_, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&BinaryRegexFilter{}})
	require.Error(t, err)
}

func TestBinaryRegexFilterInvalidEvent(t *testing.T) {
	f := []*tetragon.Filter{{BinaryRegex: []string{".*"}}}
	fl, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&BinaryRegexFilter{}})
	require.NoError(t, err)
	assert.False(t, fl.MatchOne(nil))
	assert.False(t, fl.MatchOne(&event.Event{Event: nil}))
	assert.False(t, fl.MatchOne(&event.Event{Event: struct{}{}}))
	assert.False(t, fl.MatchOne(&event.Event{Event: &tetragon.GetEventsResponse{Event: nil}}))
	assert.False(t, fl.MatchOne(&event.Event{Event: &tetragon.GetEventsResponse{
		Event: &tetragon.GetEventsResponse_ProcessExec{ProcessExec: &tetragon.ProcessExec{Process: nil}},
	}}))
	assert.False(t, fl.MatchOne(&event.Event{Event: &tetragon.GetEventsResponse{
		Event: &tetragon.GetEventsResponse_ProcessExec{ProcessExec: &tetragon.ProcessExec{Process: nil}},
	}}))
	assert.False(t, fl.MatchOne(&event.Event{Event: &tetragon.GetEventsResponse{
		Event: &tetragon.GetEventsResponse_ProcessExec{ProcessExec: &tetragon.ProcessExec{Process: nil}},
	}}))
}

func TestParentBinaryRegexFilter(t *testing.T) {
	f := []*tetragon.Filter{{ParentBinaryRegex: []string{"bash", "zsh"}}}
	fl, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&ParentBinaryRegexFilter{}})
	require.NoError(t, err)
	ev := event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/foo"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/bash"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/zsh"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
}

func TestAncestorBinaryRegexFilter(t *testing.T) {
	f := []*tetragon.Filter{{
		EventSet:            []tetragon.EventType{tetragon.EventType_PROCESS_EXEC, tetragon.EventType_PROCESS_EXIT},
		AncestorBinaryRegex: []string{"bash", "zsh"},
	}}
	fl, err := BuildFilterList(context.Background(), f, []OnBuildFilter{&AncestorBinaryRegexFilter{}})
	require.NoError(t, err)
	ev := event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/foo"},
					Process: &tetragon.Process{Binary: "/sbin/bash"},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/bash"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/bash"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
					Ancestors: []*tetragon.Process{
						&tetragon.Process{Binary: "/bin/foo"},
						&tetragon.Process{Binary: "/bin/bar"},
					},
				},
			},
		},
	}
	assert.False(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/sh"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
					Ancestors: []*tetragon.Process{
						&tetragon.Process{Binary: "/bin/foo"},
						&tetragon.Process{Binary: "/bin/bash"},
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
	ev = event.Event{
		Event: &tetragon.GetEventsResponse{
			Event: &tetragon.GetEventsResponse_ProcessExec{
				ProcessExec: &tetragon.ProcessExec{
					Parent:  &tetragon.Process{Binary: "/bin/sh"},
					Process: &tetragon.Process{Binary: "/sbin/iptables"},
					Ancestors: []*tetragon.Process{
						&tetragon.Process{Binary: "/bin/zsh"},
						&tetragon.Process{Binary: "/bin/foo"},
					},
				},
			},
		},
	}
	assert.True(t, fl.MatchOne(&ev))
}
