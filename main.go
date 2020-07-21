package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

type Permission struct {
	SharedTo   []string
	HiddenFrom []string
	All        bool
}

type ObjectPermission struct {
	UserID     string
	ObjectType string
	ObjectID   string
	Class      bool
	Public     bool
	Hidden     bool
	Networks   Permission
}

type Networks map[string][]string
type UserPerms map[string][]ObjectPermission

type System struct {
	Networks Networks
	Perms    UserPerms
	Skills   map[string][]string
}

func NewSystem() System {
	// Easier to initialise from a JSON file than make a big
	// literal map of structs of structs of ...
	b, err := ioutil.ReadFile("system.json")
	if err != nil {
		panic(err)
	}

	s := System{}
	err = json.Unmarshal(b, &s)
	if err != nil {
		panic(err)
	}

	return s
}

func main() {
	s := NewSystem()

	people := []string{}
	for v, _ := range s.Networks {
		people = append(people, v)
	}
	sort.Strings(people)

	for _, viewer := range people {
		for _, viewee := range people {
			if viewer == viewee {
				continue
			}
			r := s.CheckAll(viewer, viewee)
			fmt.Printf("%s : %s : %s\n", viewer, viewee, r)
		}
		fmt.Println()
	}
}

// `CheckAll` checks the visibility of all of `viewee`'s skills to `viewer`.
func (s System) CheckAll(viewer string, viewee string) string {
	var sb strings.Builder
	none := true

	for _, skill := range s.Skills[viewee] {
		if s.Visibility(viewer, viewee, skill) {
			if !none {
				sb.WriteString(" | ")
			}
			none = false
			sb.WriteString(skill)
		}
	}
	if none {
		return "<none>"
	}
	return sb.String()
}

// `Visibility` returns the visibility of `viewer`'s `skill` to `viewee`.
func (s System) Visibility(viewer string, viewee string, skill string) bool {
	// We might have Object or Class, Object from the SQL query.  But we don't
	// currently have an SQL query which means we need to fake up the result.
	var object, class ObjectPermission

	for _, pt := range s.Perms[viewee] {
		// We either match exactly or with the wildcard.
		if pt.ObjectID == skill || pt.ObjectID == "*" {
			// Technically `Class` should only be set on wildcard entries.
			if pt.Class {
				class = pt
			} else {
				object = pt
			}
		}
	}

	// Assume we got both a Class and Object permission entry but if the
	// Class one is blank, we actually only got an Object entry.
	entries := []ObjectPermission{class, object}
	if class.UserID == "" {
		entries = []ObjectPermission{object}
	}

	// We need the networks of our viewer.
	pnets := s.Networks[viewer]

	// Our summation.
	v := make(map[string]bool)

	// Set a false flag for every network the viewer is in.
	for _, n := range pnets {
		v[n] = false
	}

	// For every permission.
	for _, item := range entries {
		// If the object is hidden, set everything to false.
		if item.Hidden {
			for n := range v {
				v[n] = false
			}
		}

		// Take the "Shared To" list.
		sharedTo := item.Networks.SharedTo
		// If it's shared to "all networks", set the "Shared To"
		// list to be all of the viewee's networks.
		if item.Networks.All {
			sharedTo = s.Networks[viewee]
		}

		for _, n := range sharedTo {
			// If this network has a flag, set it to be true.
			// Otherwise we ignore it.
			if _, ok := v[n]; ok {
				v[n] = true
			}
		}

		// Add in the "Public" flag.
		for n := range v {
			v[n] = v[n] || item.Public
		}

		// Remove all the "Hidden From" networks.
		for _, n := range item.Networks.HiddenFrom {
			v[n] = false
		}
	}

	// Start from "invisible", add up all the flags we've set.
	f := false
	for _, k := range v {
		f = f || k
	}

	return f
}
