// Package ini reads OptiScaler-style INI files, merges a user's values into a newer file three ways, and writes
// values back into a file's own text, so that its comments, order and line endings stay as they are.
package ini

import (
	"bytes"
	"strings"
)

// Key names a value: its section and its name, as a file writes them.
type Key struct {
	Section, Name string
}

func (k Key) id() string { return strings.ToLower(k.Section) + "\x00" + strings.ToLower(k.Name) }

// String writes the key as "[Section] Name".
func (k Key) String() string { return "[" + k.Section + "] " + k.Name }

// File is an INI file's values, in the order it writes them. Sections and names match without regard to case, as
// OptiScaler reads them; a name written twice in a section takes its first value.
type File struct {
	keys   []Key
	values map[string]string
}

// Parse reads the values of an INI file: "[Section]" lines, and "name=value" lines under them. Lines starting with
// ";" or "#" are comments.
func Parse(data []byte) *File {
	f := &File{values: map[string]string{}}
	section := ""
	for _, line := range lines(data) {
		s := strings.TrimSpace(line.text)
		switch {
		case s == "" || s[0] == ';' || s[0] == '#':
		case s[0] == '[' && strings.HasSuffix(s, "]"):
			section = strings.TrimSpace(s[1 : len(s)-1])
		default:
			name, value, ok := strings.Cut(s, "=")
			if !ok {
				continue
			}
			k := Key{section, strings.TrimSpace(name)}
			if _, seen := f.values[k.id()]; !seen {
				f.keys = append(f.keys, k)
				f.values[k.id()] = strings.TrimSpace(value)
			}
		}
	}
	return f
}

// Get returns the value of k and whether the file has it.
func (f *File) Get(k Key) (string, bool) {
	v, ok := f.values[k.id()]
	return v, ok
}

// Keys returns the file's keys in its order.
func (f *File) Keys() []Key { return f.keys }

func same(a, b string) bool { return strings.EqualFold(a, b) }

// Change is a key whose value the user changed, with the user's value and the new file's.
type Change struct {
	Key          Key
	Mine, Theirs string
}

// Result is what Merge found. Every other key takes the new file's value.
type Result struct {
	Kept      []Change // only the user changed them: the user's values stay
	Conflicts []Change // the user and the new file changed them: the user chooses
	Pinned    []Change // the user changed them, but the new file's values are needed
}

// Merge compares the user's file (mine) with the new one (theirs), against the file the user's started from (base;
// nil when it is not known). A value the user did not change takes the new file's. A value only the user changed
// is kept. A value both changed is a conflict, and so is every value the user changed when base is nil. Keys for
// which pinned is true take the new file's value whatever the user did. Keys the new file lacks are dropped: the
// new version no longer reads them.
func Merge(base, mine, theirs *File, pinned func(Key) bool) Result {
	var r Result
	for _, k := range theirs.keys {
		t := theirs.values[k.id()]
		m, ok := mine.values[k.id()]
		if !ok || same(m, t) {
			continue
		}
		c := Change{k, m, t}
		if pinned != nil && pinned(k) {
			r.Pinned = append(r.Pinned, c)
			continue
		}
		if base == nil {
			r.Conflicts = append(r.Conflicts, c)
			continue
		}
		b, inBase := base.values[k.id()]
		switch {
		case inBase && same(b, t): // only the user changed it
			r.Kept = append(r.Kept, c)
		case inBase && same(b, m): // only the new file changed it
		default:
			r.Conflicts = append(r.Conflicts, c)
		}
	}
	return r
}

// Apply writes values into data, an INI file's text, in place: each key's line keeps its name, its spacing around
// "=" and its line ending, with the value replaced. Keys data lacks are left out.
func Apply(data []byte, values map[Key]string) []byte {
	want := map[string]string{}
	for k, v := range values {
		want[k.id()] = v
	}
	var out bytes.Buffer
	section := ""
	done := map[string]bool{}
	for _, line := range lines(data) {
		text := line.text
		s := strings.TrimSpace(text)
		switch {
		case s == "" || s[0] == ';' || s[0] == '#':
		case s[0] == '[' && strings.HasSuffix(s, "]"):
			section = strings.TrimSpace(s[1 : len(s)-1])
		default:
			name, _, ok := strings.Cut(s, "=")
			id := Key{section, strings.TrimSpace(name)}.id()
			if v, set := want[id]; ok && set && !done[id] {
				done[id] = true
				eq := strings.Index(text, "=")
				rest := text[eq+1:]
				lead := rest[:len(rest)-len(strings.TrimLeft(rest, " \t"))]
				text = text[:eq+1] + lead + v
			}
		}
		out.WriteString(text)
		out.WriteString(line.end)
	}
	return out.Bytes()
}

type line struct{ text, end string }

// lines splits data into lines, each with its own ending: "\r\n", "\n", or none for the last.
func lines(data []byte) []line {
	var ls []line
	s := string(data)
	for s != "" {
		i := strings.IndexByte(s, '\n')
		if i < 0 {
			ls = append(ls, line{s, ""})
			break
		}
		text, end := s[:i], "\n"
		if strings.HasSuffix(text, "\r") {
			text, end = text[:len(text)-1], "\r\n"
		}
		ls = append(ls, line{text, end})
		s = s[i+1:]
	}
	return ls
}
