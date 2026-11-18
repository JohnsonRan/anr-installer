package ini

import (
	"reflect"
	"testing"
)

const base = `; OptiScaler
[FrameGen]
; the input
FGInput=nvngxfg
FGNvngxReplacement=Apex Neural Rendering FG
DrawUIOverFG=auto

[Menu]
Scale=auto
ShortcutKey=auto
Removed=1
`

func TestParse(t *testing.T) {
	f := Parse([]byte("[A]\r\nx = 1\r\n;y=2\r\nX=3\n[b]\nz=a=b\nnoequals\n"))
	if v, ok := f.Get(Key{"a", "x"}); !ok || v != "1" {
		t.Errorf("x = %q %v; a section and a name match without regard to case, and the first value wins", v, ok)
	}
	if _, ok := f.Get(Key{"A", "y"}); ok {
		t.Error("a comment is no value")
	}
	if v, _ := f.Get(Key{"B", "z"}); v != "a=b" {
		t.Errorf("a value keeps the = in it: %q", v)
	}
	if got := len(f.Keys()); got != 2 {
		t.Errorf("%d keys, want 2", got)
	}
}

func TestMerge(t *testing.T) {
	mine := Parse(stringINI(base).withValues(map[string]string{"Scale": "1.5", "ShortcutKey": "0x2D", "FGNvngxReplacement": "nr-record", "Removed": "0"}))
	theirs := Parse([]byte(`[FrameGen]
FGInput=nvngxfg
FGNvngxReplacement=Apex Neural Rendering FG
DrawUIOverFG=true
[Menu]
Scale=auto
ShortcutKey=0x24
New=auto
`))
	pinned := func(k Key) bool { return k.Name == "FGNvngxReplacement" }
	r := Merge(Parse([]byte(base)), mine, theirs, pinned)
	want := Result{
		Kept:      []Change{{Key{"Menu", "Scale"}, "1.5", "auto"}},
		Conflicts: []Change{{Key{"Menu", "ShortcutKey"}, "0x2D", "0x24"}},
		Pinned:    []Change{{Key{"FrameGen", "FGNvngxReplacement"}, "nr-record", "Apex Neural Rendering FG"}},
	}
	if !reflect.DeepEqual(r, want) {
		t.Errorf("Merge:\n got %+v\nwant %+v", r, want)
	}
	// DrawUIOverFG: only the new file changed it, and it takes the new value. Removed: the new file lacks it.

	if got := Merge(nil, mine, theirs, pinned); len(got.Conflicts) != 3 || len(got.Kept) != 0 || len(got.Pinned) != 1 {
		t.Errorf("without a base, every value that differs is a conflict, pinned ones aside: %+v", got)
	}
}

func TestApply(t *testing.T) {
	data := []byte("[Menu]\r\nScale = auto\r\n; Scale=1\r\nKey=auto\n[Other]\nScale=auto")
	got := string(Apply(data, map[Key]string{{"menu", "scale"}: "1.5", {"Other", "Scale"}: "2", {"Menu", "Missing"}: "x"}))
	want := "[Menu]\r\nScale = 1.5\r\n; Scale=1\r\nKey=auto\n[Other]\nScale=2"
	if got != want {
		t.Errorf("Apply:\n got %q\nwant %q", got, want)
	}
}

// withValues sets values in an INI text, for the tests: the user's file is the base with some values changed.
func (s stringINI) withValues(values map[string]string) []byte {
	f := Parse([]byte(s))
	set := map[Key]string{}
	for _, k := range f.Keys() {
		if v, ok := values[k.Name]; ok {
			set[k] = v
		}
	}
	return Apply([]byte(s), set)
}

type stringINI string
