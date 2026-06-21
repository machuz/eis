package git

import "testing"

func TestBuildIdentityMap_CollapsesSameEmail(t *testing.T) {
	commits := []Commit{
		{Author: "Sebastian Markbåge", Email: "seb@meta.com"},
		{Author: "Sebastian Markbåge", Email: "seb@meta.com"},
		{Author: "sebmarkbage", Email: "seb@meta.com"},
		{Author: "Joe Savona", Email: "joe@meta.com"},
		{Author: "Joseph Savona", Email: "joe@meta.com"},
		{Author: "Joe Savona", Email: "joe@meta.com"},
		{Author: "Solo Dev", Email: "solo@x.com"},
	}
	m := BuildIdentityMap(commits)
	// minority spelling under each email collapses to the majority name
	if m["sebmarkbage"] != "Sebastian Markbåge" {
		t.Errorf("sebmarkbage -> %q, want Sebastian Markbåge", m["sebmarkbage"])
	}
	if m["Joseph Savona"] != "Joe Savona" {
		t.Errorf("Joseph Savona -> %q, want Joe Savona", m["Joseph Savona"])
	}
	// the canonical names themselves and a solo author must not appear (no-op)
	if _, ok := m["Sebastian Markbåge"]; ok {
		t.Errorf("canonical name should not remap to itself")
	}
	if _, ok := m["Solo Dev"]; ok {
		t.Errorf("solo author should not be in the map")
	}
}

func TestBuildIdentityMap_CrossEmailSameNameUnifies(t *testing.T) {
	// Jan Kassens commits under two emails; the "kassens" handle appears under
	// one of them. The handle should collapse to "Jan Kassens".
	commits := []Commit{
		{Author: "Jan Kassens", Email: "jan@meta.com"},
		{Author: "Jan Kassens", Email: "jan@meta.com"},
		{Author: "kassens", Email: "jan@noreply.github.com"},
		{Author: "Jan Kassens", Email: "jan@noreply.github.com"},
	}
	m := BuildIdentityMap(commits)
	if m["kassens"] != "Jan Kassens" {
		t.Errorf("kassens -> %q, want Jan Kassens", m["kassens"])
	}
}

func TestBuildIdentityMap_IgnoresEmptyEmail(t *testing.T) {
	commits := []Commit{
		{Author: "A", Email: ""},
		{Author: "B", Email: ""},
	}
	if len(BuildIdentityMap(commits)) != 0 {
		t.Error("empty emails must not produce mappings")
	}
}

func TestCanonicalizeAuthors_RewritesInPlace(t *testing.T) {
	idmap := map[string]string{"sebmarkbage": "Sebastian Markbåge"}
	commits := []Commit{{Author: "sebmarkbage"}, {Author: "Other"}}
	blame := []BlameLine{{Author: "sebmarkbage"}, {Author: "Other"}}
	CanonicalizeAuthors(commits, blame, idmap)
	if commits[0].Author != "Sebastian Markbåge" || commits[1].Author != "Other" {
		t.Errorf("commit remap wrong: %+v", commits)
	}
	if blame[0].Author != "Sebastian Markbåge" || blame[1].Author != "Other" {
		t.Errorf("blame remap wrong: %+v", blame)
	}
}
