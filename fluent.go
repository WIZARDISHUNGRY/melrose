package melrose

import "sync"

type Fluent struct{}

func (f Fluent) Sequence(notation string) Sequenceable {
	return MustParseSequence(notation)
}

func (f Fluent) Repeat(times int, s Sequenceable) Sequenceable {
	return Repeat{Target: s, Times: times}
}

func (f Fluent) Pitch(semitones int, s Sequenceable) Sequenceable {
	return Pitch{Target: s, Semitones: semitones}
}

func (f Fluent) Parallel(s Sequenceable) Sequenceable {
	return Parallel{Target: s}
}

func (f Fluent) Note(s string) Note { return MustParseNote(s) }

func (f Fluent) Join(s ...Sequenceable) Sequenceable {
	return Join{List: s}
}

func (f Fluent) Go(a AudioDevice, s ...Sequenceable) {
	wg := new(sync.WaitGroup)
	for _, each := range s {
		wg.Add(1)
		go func(p Sequenceable) {
			a.Play(p, true)
			wg.Done()
		}(each)
	}
	wg.Wait()
}

// IndexMap creates a IndexMapper from indices.
// Example of indices: "1 (2 3 4) 5 (6 7)". One-based indexes.
func (f Fluent) IndexMap(s Sequenceable, indices string) Sequenceable {
	return IndexMapper{Target: s, Indices: parseIndices(indices)}
}

// Chord creates a new Chord by parsing the input. See Chord for the syntax.
func (f Fluent) Chord(s string) Chord {
	return MustParseChord(s)
}

// Serial returns a new object that serialises all the notes of the argument.
func (f Fluent) Serial(s Sequenceable) Sequenceable {
	return Serial{Target: s}
}
