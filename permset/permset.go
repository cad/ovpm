// Package permset provides primitives for permission management.
package permset

import (
	"context"
	"fmt"
)

// Perm is a permission to do some action.
type Perm int

// Permset represents a set of permissions.
type Permset struct {
	permset map[Perm]bool
}

// New receives permissions to contain and returns a permset from it.
func New(perms ...Perm) Permset {
	permset := Permset{permset: make(map[Perm]bool)}
	permset.Add(perms...)
	return permset
}

// Add adds the received perms to the permset.
func (ps *Permset) Add(perms ...Perm) {
	for _, perm := range perms {
		ps.permset[perm] = true
	}
}

// Remove removes the received perms from the permset.
func (ps *Permset) Remove(perms ...Perm) {
	for _, perm := range perms {
		if _, ok := ps.permset[perm]; ok {
			delete(ps.permset, perm)
		}
	}
}

// Perms returns the permissions contained within the permset.
func (ps *Permset) Perms() []Perm {
	var perms []Perm
	for k := range ps.permset {
		perms = append(perms, k)
	}
	return perms
}

// Contains receives single Perm and returns true if the permset contains it.
func (ps *Permset) Contains(perm Perm) bool {
	if _, ok := ps.permset[perm]; !ok {
		return false
	}
	return true
}

// ContainsAll returns true if the permset contains all received Perms.
func (ps *Permset) ContainsAll(perms ...Perm) bool {
	for _, perm := range perms {
		if _, ok := ps.permset[perm]; !ok {
			return false
		}
	}
	return true
}

// ContainsSome returns true if the permset contains any one or more of the received Perms.
func (ps *Permset) ContainsSome(perms ...Perm) bool {
	for _, perm := range perms {
		if _, ok := ps.permset[perm]; ok {
			return true
		}
	}
	return false
}

// ContainsNone returns true if the permset contains none one of the received Perms.
func (ps *Permset) ContainsNone(perms ...Perm) bool {
	for _, perm := range perms {
		if _, ok := ps.permset[perm]; ok {
			return false
		}
	}
	return true
}

type permsetKeyType int

const permsetKey permsetKeyType = iota

// NewContext receives perms and returns a context with the received perms are the value of the context.
func NewContext(ctx context.Context, permset Permset) context.Context {
	return context.WithValue(ctx, permsetKey, permset)
}

// FromContext receives a context and returns the permset in it.
func FromContext(ctx context.Context) (Permset, error) {
	permset, ok := ctx.Value(permsetKey).(Permset)
	if !ok {
		return Permset{}, fmt.Errorf("cannot get context value")
	}
	return permset, nil
}
