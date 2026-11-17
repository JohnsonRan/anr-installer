package detect

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ExeName returns the product name in an .exe's version resource, else its file description, or "".
func ExeName(exe string) string {
	size, err := windows.GetFileVersionInfoSize(exe, nil)
	if err != nil || size == 0 {
		return ""
	}
	buf := make([]byte, size)
	if err := windows.GetFileVersionInfo(exe, 0, size, unsafe.Pointer(&buf[0])); err != nil {
		return ""
	}
	// The string tables are named by language and code page, as \VarFileInfo\Translation lists them.
	var langs []string
	var p unsafe.Pointer
	var n uint32
	if windows.VerQueryValue(unsafe.Pointer(&buf[0]), `\VarFileInfo\Translation`, unsafe.Pointer(&p), &n) == nil && n >= 4 {
		for i := uint32(0); i+4 <= n; i += 4 {
			pair := (*[2]uint16)(unsafe.Add(p, i))
			langs = append(langs, fmt.Sprintf("%04x%04x", pair[0], pair[1]))
		}
	}
	langs = append(langs, "040904b0", "040904e4", "000004b0")
	for _, key := range []string{"ProductName", "FileDescription"} {
		for _, l := range langs {
			if windows.VerQueryValue(unsafe.Pointer(&buf[0]), `\StringFileInfo\`+l+`\`+key, unsafe.Pointer(&p), &n) != nil || n == 0 {
				continue
			}
			s := strings.TrimSpace(windows.UTF16PtrToString((*uint16)(p)))
			if s != "" {
				return s
			}
		}
	}
	return ""
}
