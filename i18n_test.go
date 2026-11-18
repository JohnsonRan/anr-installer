package main

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// verbRE matches a printf verb: an optional argument index, flags, width and precision, and the verb.
var verbRE = regexp.MustCompile(`%(?:\[(\d+)\])?[-+# 0]*\d*(?:\.\d+)?([a-zA-Z%])`)

// verbs returns the verbs of a format, each as the argument it formats and its letter, sorted: a translation may
// reorder the arguments by index, but must format the same ones the same way.
func verbs(format string) []string {
	var out []string
	next := 1
	for _, m := range verbRE.FindAllStringSubmatch(format, -1) {
		if m[2] == "%" {
			continue
		}
		arg := next
		if m[1] != "" {
			arg, _ = strconv.Atoi(m[1])
		}
		next = arg + 1
		out = append(out, strconv.Itoa(arg)+m[2])
	}
	slices.Sort(out)
	return out
}

func TestVerbs(t *testing.T) {
	if got := verbs("%[2]s, %[1]d of %d %% done"); !slices.Equal(got, []string{"1d", "2d", "2s"}) {
		t.Errorf("verbs = %q", got)
	}
}

// Every translation has every text, no other, and formats the same arguments as the English.
func TestTranslationsComplete(t *testing.T) {
	langs := map[string]map[string]string{"zh-Hans": {}, "zh-Hant": messagesZhHant, "ja": messagesJa}
	for key, m := range messages {
		langs["zh-Hans"][key] = m[1]
	}
	for name, tr := range langs {
		for key, m := range messages {
			s, ok := tr[key]
			if !ok {
				t.Errorf("%s lacks %q", name, key)
				continue
			}
			if want, got := verbs(m[0]), verbs(s); !slices.Equal(want, got) {
				t.Errorf("%s %q formats %q, English %q", name, key, got, want)
			}
		}
		for key := range tr {
			if _, ok := messages[key]; !ok {
				t.Errorf("%s has %q, which messages lacks", name, key)
			}
		}
	}
}

// simplifiedOnly are characters of Simplified Chinese that Traditional writes otherwise: a sample of the common
// ones, to catch a text left in, or converted to, the wrong script.
const simplifiedOnly = "设应显无请选择夹载录们这个为与对开关闭发现动认证资讯视频网络运戏统级装档机说时间进项务"

func TestTraditionalChineseScript(t *testing.T) {
	for key, s := range messagesZhHant {
		if key == "settings.language.zh-Hans" { // the language's own name
			continue
		}
		if i := strings.IndexAny(s, simplifiedOnly); i >= 0 {
			t.Errorf("zh-Hant %q has the Simplified %q: %s", key, []rune(s[i:])[0], s)
		}
	}
}
