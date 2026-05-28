package provider

import (
	"strings"
	"testing"
)

// TestJoinGroupIDs verifies the composite-ID format used as the resource
// ID and the import-ID protocol.
func TestJoinGroupIDs(t *testing.T) {
	cases := []struct {
		name string
		in   []groupPriorityBlockModel
		want string
	}{
		{"empty", nil, ""},
		{"one", []groupPriorityBlockModel{{GroupID: "g1", Priority: 1}}, "g1"},
		{
			"multiple",
			[]groupPriorityBlockModel{
				{GroupID: "g1", Priority: 1},
				{GroupID: "g2", Priority: 2},
				{GroupID: "g3", Priority: 3},
			},
			"g1,g2,g3",
		},
	}
	for _, c := range cases {
		got := joinGroupIDs(c.in)
		if got != c.want {
			t.Errorf("%s: want %q, got %q", c.name, c.want, got)
		}
	}
}

// TestJoinGroupIDs_PreservesOrder verifies a reorder in the plan changes
// the ID — important so Terraform sees re-ordering as a real change
// rather than a no-op.
func TestJoinGroupIDs_PreservesOrder(t *testing.T) {
	a := []groupPriorityBlockModel{
		{GroupID: "g1", Priority: 1},
		{GroupID: "g2", Priority: 2},
	}
	b := []groupPriorityBlockModel{
		{GroupID: "g2", Priority: 1},
		{GroupID: "g1", Priority: 2},
	}
	if joinGroupIDs(a) == joinGroupIDs(b) {
		t.Error("expected different IDs for re-ordered priorities")
	}
}

// TestSplitGroupIDs verifies import-ID parsing. Spaces around commas and
// empty entries are tolerated so users can hand-edit without ceremony.
func TestSplitGroupIDs(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", []string{}},
		{"g1", []string{"g1"}},
		{"g1,g2,g3", []string{"g1", "g2", "g3"}},
		{" g1 , g2 , g3 ", []string{"g1", "g2", "g3"}},
		{"g1,,g2", []string{"g1", "g2"}},
		{",,,", []string{}},
	}
	for _, c := range cases {
		got := splitGroupIDs(c.in)
		if len(got) != len(c.want) {
			t.Errorf("%q: want %v, got %v", c.in, c.want, got)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q[%d]: want %q, got %q", c.in, i, c.want[i], got[i])
			}
		}
	}
}

// TestJoinSplitRoundTrip verifies the composite ID survives an
// import-style round-trip — what we write into state can be parsed back.
func TestJoinSplitRoundTrip(t *testing.T) {
	original := []groupPriorityBlockModel{
		{GroupID: "abc123", Priority: 1},
		{GroupID: "def456", Priority: 2},
	}
	id := joinGroupIDs(original)
	got := splitGroupIDs(id)
	if len(got) != len(original) {
		t.Fatalf("length mismatch: want %d, got %d", len(original), len(got))
	}
	for i := range got {
		if got[i] != original[i].GroupID {
			t.Errorf("[%d]: want %q, got %q", i, original[i].GroupID, got[i])
		}
	}
	// And the ID is a clean comma-join.
	if strings.Contains(id, " ") {
		t.Errorf("composite ID should not contain spaces: %q", id)
	}
}
