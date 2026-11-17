//go:build !windows

package detect

// ExeName returns "": version resources are read on Windows only.
func ExeName(string) string { return "" }
