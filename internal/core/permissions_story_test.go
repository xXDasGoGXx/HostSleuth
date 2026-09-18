package core

import (
	"os"
	"testing"
)

func TestPermissionDecisionUsesOwnerGroupAndOtherClasses(t *testing.T) {
	identity := &PermissionIdentity{UID: 1001, GID: 1002, SupplementaryGIDs: []uint32{2000}}

	tests := []struct {
		name   string
		node   PermissionNode
		op     string
		bit    uint32
		status string
		class  string
	}{
		{"owner read", PermissionNode{Path: "/x", UID: 1001, GID: 9, Mode: "0640"}, "read", 4, "pass", "owner"},
		{"owner execute denied", PermissionNode{Path: "/x", UID: 1001, GID: 9, Mode: "0640"}, "execute", 1, "fail", "owner"},
		{"supplementary group read", PermissionNode{Path: "/x", UID: 9, GID: 2000, Mode: "0040"}, "read", 4, "pass", "group"},
		{"other write denied", PermissionNode{Path: "/x", UID: 9, GID: 10, Mode: "0644"}, "write", 2, "fail", "other"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := permissionDecision(tc.node, identity, tc.op, tc.bit)
			if got.Status != tc.status || got.Relation != tc.class {
				t.Fatalf("got status=%s relation=%s; want %s/%s", got.Status, got.Relation, tc.status, tc.class)
			}
		})
	}
}

func TestPermissionDecisionsRequireParentTraversal(t *testing.T) {
	identity := &PermissionIdentity{UID: 1001, GID: 1001}
	chain := []PermissionNode{
		{Path: "/", Kind: "directory", UID: 0, GID: 0, Mode: "0755", Relation: "other"},
		{Path: "/srv", Kind: "directory", UID: 0, GID: 0, Mode: "0755", Relation: "other"},
		{Path: "/srv/private", Kind: "directory", UID: 0, GID: 0, Mode: "0700", Relation: "other"},
		{Path: "/srv/private/data", Kind: "file", UID: 1001, GID: 1001, Mode: "0644", Relation: "owner", IsTarget: true},
	}
	decisions := permissionDecisions(chain, identity)
	var found bool
	for _, decision := range decisions {
		if decision.Path == "/srv/private" && decision.Operation == "traverse" {
			found = true
			if decision.Status != "fail" {
				t.Fatalf("expected traversal failure, got %#v", decision)
			}
		}
	}
	if !found {
		t.Fatal("missing parent traversal decision")
	}
	story := PermissionStory{Identity: identity, ResolvedPath: "/srv/private/data", Decisions: decisions}
	status, first, _ := permissionStoryOutcome(story)
	if status != "fail" || first != "traverse:/srv/private" {
		t.Fatalf("unexpected outcome: status=%s first=%s", status, first)
	}
}

func TestSocketUsesWriteBitForConnectDecision(t *testing.T) {
	identity := &PermissionIdentity{UID: 1001, GID: 1001}
	chain := []PermissionNode{
		{Path: "/", Kind: "directory", UID: 0, GID: 0, Mode: "0755"},
		{Path: "/run", Kind: "directory", UID: 0, GID: 0, Mode: "0755"},
		{Path: "/run/app.sock", Kind: "socket", UID: 0, GID: 1001, Mode: "0660", IsTarget: true},
	}
	decisions := permissionDecisions(chain, identity)
	var connect *PermissionDecision
	for i := range decisions {
		if decisions[i].Operation == "connect" {
			connect = &decisions[i]
		}
	}
	if connect == nil || connect.Status != "pass" || connect.Required != "w" {
		t.Fatalf("unexpected socket decision: %#v", connect)
	}
}

func TestPermissionDecisionAccountsForRootDACCapabilities(t *testing.T) {
	node := PermissionNode{Path: "/srv/private", Kind: "directory", UID: 1000, GID: 1000, Mode: "0000"}

	withOverride := &PermissionIdentity{UID: 0, GID: 0, EffectiveCaps: "0000000000000002"}
	got := permissionDecision(node, withOverride, "traverse", 1)
	if got.Status != "pass" {
		t.Fatalf("CAP_DAC_OVERRIDE should permit traversal: %#v", got)
	}

	withoutCaps := &PermissionIdentity{UID: 0, GID: 0, EffectiveCaps: "0000000000000000"}
	got = permissionDecision(node, withoutCaps, "traverse", 1)
	if got.Status != "fail" {
		t.Fatalf("root without DAC capabilities should follow mode bits: %#v", got)
	}

	unknownCaps := &PermissionIdentity{UID: 0, GID: 0}
	got = permissionDecision(node, unknownCaps, "traverse", 1)
	if got.Status != "unknown" {
		t.Fatalf("missing capability evidence should be unknown: %#v", got)
	}
}

func TestPathWithinIsComponentAware(t *testing.T) {
	if !pathWithin("/srv/app/data/file", "/srv/app") {
		t.Fatal("expected descendant to match")
	}
	if pathWithin("/srv/application/file", "/srv/app") {
		t.Fatal("prefix-only path must not match")
	}
	if !pathWithin("/srv/app", "/srv/app") {
		t.Fatal("same path should match")
	}
}

func TestPermissionPathChainDoesNotReadContents(t *testing.T) {
	root := t.TempDir()
	dir := root + "/a"
	file := dir + "/payload"
	if err := os.Mkdir(dir, 0750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("secret-not-read-by-permission-story"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file, 0640); err != nil {
		t.Fatal(err)
	}
	identity := &PermissionIdentity{UID: uint32(os.Geteuid()), GID: uint32(os.Getegid())}
	resolved, chain := permissionPathChain(file, identity)
	if resolved != file {
		t.Fatalf("resolved=%q want %q", resolved, file)
	}
	if len(chain) < 3 || !chain[len(chain)-1].IsTarget {
		t.Fatalf("unexpected chain: %#v", chain)
	}
	if chain[len(chain)-1].Kind != "file" || chain[len(chain)-1].Mode != "0640" {
		t.Fatalf("unexpected target metadata: %#v", chain[len(chain)-1])
	}
}
