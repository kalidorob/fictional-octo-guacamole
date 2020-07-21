package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
)

type Entry struct {
	SharedTo   []string
	HiddenFrom []string
	Public     bool
	Hidden     bool
	AllNet     bool
}

type Perm struct {
	UserID      string
	ObjectType  string
	ObjectID    string
	Class       bool
	Permissions Entry
}

type Networks map[string][]string
type UserPerms map[string][]Perm

type System struct {
	networks Networks
	perms    UserPerms
	skills   map[string][]string
}

func NewSystem() System {
	b, err := ioutil.ReadFile("perms.json")
	if err != nil {
		panic(err)
	}

	perms := make(UserPerms)
	err = json.Unmarshal(b, &perms)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("%#v\n", perms)

	networks := map[string][]string{
		"Alice": []string{"Stonebroti", "Morfi", "Boundgrave"},
		"Bob":   []string{"Boundgrave", "Terregonje"},
		"Chip":  []string{"Mextunmo", "Terregonje"},
		"Diana": []string{"Mextunmo", "Terregonje"},
		"Frank": []string{"Toompark", "Percalcombe"},
	}

	skills := map[string][]string{
		"Alice": []string{"Alchemy", "Acrobatics", "Archery"},
		"Bob":   []string{"Brainwashing", "Boating", "Birdwatching"},
		"Chip":  []string{"Alchemy", "Cooking", "Criminology"},
		"Diana": []string{"Dancing", "Diplomacy", "Disguise"},
		"Frank": []string{"Falconry", "Forgery", "Forensics"},
	}

	return System{networks, perms, skills}
}

func main() {
	s := NewSystem()

	people := []string{}
	for v, _ := range s.networks {
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

	/*
		s.Check("Bob", "Diana", "Dancing")
		s.Check("Bob", "Diana", "Diplomacy")
		s.Check("Bob", "Diana", "Disguise")
		fmt.Println()
		s.Check("Frank", "Diana", "Dancing")
		s.Check("Frank", "Diana", "Diplomacy")
		s.Check("Frank", "Diana", "Disguise")
		fmt.Println()
		s.Check("Chip", "Diana", "Dancing")
		fmt.Println()
		s.Check("Diana", "Chip", "Alchemy")
		s.Check("Diana", "Chip", "Cooking")
		s.Check("Diana", "Chip", "Criminology")
	*/
}

func (s System) CheckAll(viewer string, viewee string) string {
	var sb strings.Builder
	none := true

	for _, skill := range s.skills[viewee] {
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

func (s System) Check(viewer string, viewee string, skill string) {
	fmt.Printf("%s looking at %s's %s is ", viewer, viewee, skill)
	v := s.Visibility(viewer, viewee, skill)
	if v {
		fmt.Printf("ALLOWED\n")
	} else {
		fmt.Printf("BANNED\n")
	}
}

func (s System) Visibility(viewer string, viewee string, skill string) bool {
	// We might have Object or Class, Object from the query
	var p Perm
	var q Perm

	for _, pt := range s.perms[viewee] {
		if pt.ObjectID == skill || pt.ObjectID == "*" {
			if pt.Class {
				q = pt
			} else {
				p = pt
			}
		}
	}

	x := []Perm{q, p}
	if q.UserID == "" {
		x = []Perm{p}
	}
	// End of making our fake SQL results.

	// We need the networks of both people.
	jnets := s.networks[viewee]
	pnets := s.networks[viewer]

	// Our summation.
	v := make(map[string]bool)

	// Set a false flag for every network the viewer is in.
	for _, n := range pnets {
		v[n] = false
	}

	// For every permission.
	for _, item := range x {
		// If the object is hidden, set everything to false.
		if item.Permissions.Hidden {
			for n := range v {
				v[n] = false
			}
		}

		// Take the "Shared To" list.
		sharedTo := item.Permissions.SharedTo
		// If it's shared to "all networks", set the "Shared To"
		// list to be all of the viewee's networks.
		if item.Permissions.AllNet {
			sharedTo = jnets
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
			v[n] = v[n] || item.Permissions.Public
		}

		// Remove all the "Hidden From" networks.
		for _, n := range item.Permissions.HiddenFrom {
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
