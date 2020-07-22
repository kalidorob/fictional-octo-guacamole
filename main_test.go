package main

import (
	"strings"
	"testing"
)

const data = `Alice : Bob : Boating | Birdwatching
Alice : Chip : <none>
Alice : Diana : Diplomacy
Alice : Frank : <none>
Bob : Alice : Alchemy | Acrobatics
Bob : Chip : Alchemy | Cooking | Criminology
Bob : Diana : Diplomacy
Bob : Frank : <none>
Chip : Alice : Alchemy
Chip : Bob : Boating | Birdwatching
Chip : Diana : Dancing | Diplomacy
Chip : Frank : <none>
Diana : Alice : Alchemy
Diana : Bob : Boating | Birdwatching
Diana : Chip : Alchemy | Cooking | Criminology
Diana : Frank : <none>
Frank : Alice : Alchemy
Frank : Bob : Birdwatching
Frank : Chip : <none>
Frank : Diana : Diplomacy
Alice : Gerald : <none>
Bob : Gerald : <none>
Chip : Gerald : <none>
Diana : Gerald : <none>
Frank : Gerald : Gambling | Geology | Generosity
Gerald : Alice : Alchemy
Gerald : Bob : Birdwatching
Gerald : Chip : <none>
Gerald : Diana : Diplomacy
Gerald : Frank : Falconry | Forgery | Forensics`

func TestAll(t *testing.T) {
	s := NewSystem()
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		parts := strings.Split(line, " : ")
		viewer := parts[0]
		viewee := parts[1]
		expected := parts[2]

		result := s.CheckAll(viewer, viewee)

		if result != expected {
			t.Errorf("[%s] <> [%s]", expected, result)
		}
	}
}

func BenchmarkAll(b *testing.B) {
	s := NewSystem()
	for n := 0; n < b.N; n++ {
		s.CheckAll("Diana", "Chip")
	}
}

func BenchmarkOne(b *testing.B) {
	s := NewSystem()
	for n := 0; n < b.N; n++ {
		s.Visibility("Diana", "Chip", "Cooking")
	}
}
