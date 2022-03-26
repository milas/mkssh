package sshconfig

import (
	"testing"

	sshconfig "github.com/kevinburke/ssh_config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateHost(t *testing.T) {
	h, err := GenerateHost("test", "test.local", "/home/test/.ssh/test_ed25519")
	require.NoError(t, err, "Error generating host")
	require.NotNil(t, h, "Host object was nil")
	if assert.Len(t, h.Patterns, 1) {
		assert.Equal(t, "test", h.Patterns[0].String())
	}
	assert.Equal(t, "Generated by mkssh", h.EOLComment)
	assert.Equal(
		t,
		[]sshconfig.Node{
			&sshconfig.KV{Key: "HostName", Value: "test.local"},
			&sshconfig.KV{Key: "IdentitiesOnly", Value: "yes"},
			&sshconfig.KV{
				Key:     "IdentityFile",
				Value:   "/home/test/.ssh/test_ed25519",
				Comment: "Generated by mkssh",
			},
		},
		h.Nodes,
	)
}

func TestAddOrReplaceHost_Add(t *testing.T) {
	cfg := &sshconfig.Config{
		Hosts: []*sshconfig.Host{
			{
				Patterns: mustPatterns(t, "foo.local"),
				Nodes: []sshconfig.Node{
					&sshconfig.KV{
						Key:     "FakeKey",
						Value:   "value1",
						Comment: "A comment",
					},
				},
			},
		},
	}

	h := &sshconfig.Host{
		Patterns: mustPatterns(t, "test.local"),
		Nodes: []sshconfig.Node{
			&sshconfig.KV{Key: "FakeKey", Value: "value2"},
		},
	}

	AddOrReplaceHost(cfg, h)
	require.Len(t, cfg.Hosts, 2)
	require.Equal(t, "foo.local", cfg.Hosts[0].Patterns[0].String())
	require.Equal(
		t,
		[]sshconfig.Node{
			&sshconfig.KV{Key: "FakeKey", Value: "value1", Comment: "A comment"},
		},
		cfg.Hosts[0].Nodes,
	)
	require.Equal(t, "test.local", cfg.Hosts[1].Patterns[0].String())
	require.Equal(
		t,
		[]sshconfig.Node{
			&sshconfig.KV{Key: "FakeKey", Value: "value2"},
		}, cfg.Hosts[1].Nodes,
	)
}

func TestAddOrReplaceHost_Replace(t *testing.T) {
	cfg := &sshconfig.Config{
		Hosts: []*sshconfig.Host{
			{
				Patterns: mustPatterns(t, "foo.local"),
				Nodes: []sshconfig.Node{
					&sshconfig.KV{Key: "FakeKey", Value: "value1"},
					&sshconfig.KV{Key: "FakeKey2", Value: "other"},
				},
			},
		},
	}

	h := &sshconfig.Host{
		Patterns: mustPatterns(t, "foo.local"),
		Nodes: []sshconfig.Node{
			&sshconfig.KV{Key: "FakeKey", Value: "value2"},
		},
	}

	AddOrReplaceHost(cfg, h)
	require.Len(t, cfg.Hosts, 1)
	require.Equal(t, "foo.local", cfg.Hosts[0].Patterns[0].String())
	require.Equal(t, []sshconfig.Node{&sshconfig.KV{Key: "FakeKey", Value: "value2"}}, cfg.Hosts[0].Nodes)
}

func mustPattern(t testing.TB, pattern string) *sshconfig.Pattern {
	t.Helper()
	p, err := sshconfig.NewPattern(pattern)
	require.NoError(t, err, "Invalid SSH Host pattern")
	return p
}

func mustPatterns(t testing.TB, patterns ...string) []*sshconfig.Pattern {
	t.Helper()
	out := make([]*sshconfig.Pattern, len(patterns))
	for i := range patterns {
		out[i] = mustPattern(t, patterns[i])
	}
	return out
}