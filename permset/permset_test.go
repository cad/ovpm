package permset

import (
	"context"
	"testing"
)

const (
	TestPerm1 Perm = iota
	TestPerm2
	TestPerm3
)

func TestNew(t *testing.T) {
	var newtests = []struct {
		perms []Perm
	}{
		{[]Perm{TestPerm1}},
		{[]Perm{TestPerm1, TestPerm2}},
		{[]Perm{TestPerm3, TestPerm2}},
		{[]Perm{TestPerm2, TestPerm2}},
		{[]Perm{TestPerm3, TestPerm2, TestPerm3}},
	}

	for _, tt := range newtests {
		permset := New(tt.perms...)

		// See if perms within the permset checks out with the ones provided to the New().
		for _, perm := range tt.perms {
			if _, ok := permset.permset[perm]; !ok {
				t.Fatalf("perm should exist in the permset: %v", perm)
			}
		}

		// See if there are any extra perms in the permset that is not provided to the New().
		for perm := range permset.permset {
			var found bool
			for _, ttPerm := range tt.perms {
				if ttPerm == perm {
					found = true
				}
			}

			if !found {
				t.Fatalf("perm should not exist in the permset: %v", perm)
			}
		}
	}
}

func TestAdd(t *testing.T) {
	permset := New(TestPerm2)
	permset.Add(TestPerm3)

	if _, ok := permset.permset[TestPerm3]; !ok {
		t.Fatal("perm TestPerm3 should exist in the permset")
	}
	if _, ok := permset.permset[TestPerm2]; !ok {
		t.Fatal("perm TestPerm2 should exist in the permset")
	}

}

func TestRemove(t *testing.T) {
	// See if remove works OK.
	permset := New(TestPerm2, TestPerm3)
	permset.Remove(TestPerm3)

	if _, ok := permset.permset[TestPerm3]; ok {
		t.Fatal("perm TestPerm3 should not exist in the permset")
	}
	if _, ok := permset.permset[TestPerm2]; !ok {
		t.Fatal("perm TestPerm2 should exist in the permset")
	}

	// See if double remove breaks it.
	permset = New(TestPerm2, TestPerm3)
	permset.Remove(TestPerm3)
	permset.Remove(TestPerm3)

	if _, ok := permset.permset[TestPerm3]; ok {
		t.Fatal("perm TestPerm3 should not exist in the permset")
	}
	if _, ok := permset.permset[TestPerm2]; !ok {
		t.Fatal("perm TestPerm2 should exist in the permset")
	}

}

func TestPerms(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)
	perms := permset.Perms()

	var found bool
	for _, perm := range perms {
		if perm == TestPerm2 {
			found = true
		}
	}

	if !found {
		t.Fatal("Perms() should return all the perms within the permset")
	}
}

func TestContains(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)

	if !permset.Contains(TestPerm2) {
		t.Fatal("permset should contain TestPerm2")
	}

	if !permset.Contains(TestPerm3) {
		t.Fatal("permset should contain TestPerm3")
	}

	if permset.Contains(TestPerm1) {
		t.Fatal("permset should  not contain TestPerm1")
	}
}

func TestContainsAll(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)

	if !permset.ContainsAll(TestPerm2) {
		t.Fatal("permset should contain TestPerm2")
	}

	if !permset.ContainsAll(TestPerm3) {
		t.Fatal("permset should contain TestPerm3")
	}

	if permset.ContainsAll(TestPerm1) {
		t.Fatal("permset should  not contain TestPerm1")
	}

	if !permset.ContainsAll(TestPerm2, TestPerm3) {
		t.Fatal("permset should contain TestPerm2 and TestPerm3")
	}

	if !permset.ContainsAll(TestPerm3, TestPerm2) {
		t.Fatal("permset should contain TestPerm2 and TestPerm3")
	}

	if permset.ContainsAll(TestPerm1, TestPerm2) {
		t.Fatal("permset should not contain TestPerm1 and TestPerm3")
	}

}

func TestContainsSome(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)

	if !permset.ContainsSome(TestPerm2) {
		t.Fatal("permset should contain TestPerm2")
	}

	if !permset.ContainsSome(TestPerm1, TestPerm3) {
		t.Fatal("permset should contain TestPerm1 and TestPerm3")
	}

	if permset.ContainsSome(TestPerm1) {
		t.Fatal("permset should contain TestPerm1")
	}

}

func TestContainsNone(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)

	if !permset.ContainsNone(TestPerm1) {
		t.Fatal("ContainsNone should return true")
	}

	if permset.ContainsNone(TestPerm2) {
		t.Fatal("ContainsNone should return false")
	}
}

func TestNewContext(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)
	ctx := NewContext(context.Background(), permset)

	permsetFromCtx, ok := ctx.Value(permsetKey).(Permset)
	if !ok {
		t.Fatal("can't extract permset from ctx")
	}

	if !permset.ContainsAll(permsetFromCtx.Perms()...) {
		t.Fatal("permsets should match")
	}
}

func TestFromContext(t *testing.T) {
	permset := New(TestPerm2, TestPerm3)
	ctx := NewContext(context.Background(), permset)

	permsetFromCtx, err := FromContext(ctx)
	if err != nil {
		t.Fatalf("error is not expected here: %v", err)
	}

	if !permset.ContainsAll(permsetFromCtx.Perms()...) {
		t.Fatal("permsets should match")
	}
}
