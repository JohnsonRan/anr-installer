// Package steam finds the Steam libraries on this machine and the games installed in them.
package steam

import (
	"fmt"
	"strings"
)

// Node is a value of Valve's KeyValues text format (VDF/ACF): either a string or a block of named children.
type Node struct {
	Value    string
	Children map[string]*Node
	Order    []string // keys in file order, lower-cased
}

// Get returns the child named key (case-insensitive), or nil.
func (n *Node) Get(key string) *Node {
	if n == nil || n.Children == nil {
		return nil
	}
	return n.Children[strings.ToLower(key)]
}

// Str returns the string value of the child named key, or "".
func (n *Node) Str(key string) string {
	if c := n.Get(key); c != nil {
		return c.Value
	}
	return ""
}

// ParseVDF parses a KeyValues text document into a root block.
func ParseVDF(src string) (*Node, error) {
	p := &vdfParser{src: src}
	root := &Node{Children: map[string]*Node{}}
	if err := p.block(root, true); err != nil {
		return nil, err
	}
	return root, nil
}

type vdfParser struct {
	src string
	pos int
}

func (p *vdfParser) skip() {
	for p.pos < len(p.src) {
		ch := p.src[p.pos]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == 0xEF || ch == 0xBB || ch == 0xBF:
			p.pos++
		case ch == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '/':
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		case ch == '[': // platform conditionals such as [$WIN32]
			for p.pos < len(p.src) && p.src[p.pos] != ']' {
				p.pos++
			}
			p.pos++
		default:
			return
		}
	}
}

// token returns the next string token, or "{" / "}" as structural tokens (quoted reports whether it was a string).
func (p *vdfParser) token() (tok string, quoted bool, err error) {
	p.skip()
	if p.pos >= len(p.src) {
		return "", false, nil
	}
	ch := p.src[p.pos]
	if ch == '{' || ch == '}' {
		p.pos++
		return string(ch), false, nil
	}
	var b strings.Builder
	if ch == '"' {
		p.pos++
		for p.pos < len(p.src) {
			c := p.src[p.pos]
			if c == '\\' && p.pos+1 < len(p.src) {
				p.pos++
				switch e := p.src[p.pos]; e {
				case 'n':
					b.WriteByte('\n')
				case 't':
					b.WriteByte('\t')
				default:
					b.WriteByte(e)
				}
				p.pos++
				continue
			}
			if c == '"' {
				p.pos++
				return b.String(), true, nil
			}
			b.WriteByte(c)
			p.pos++
		}
		return "", false, fmt.Errorf("vdf: unterminated string")
	}
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '{' || c == '}' || c == '"' {
			break
		}
		b.WriteByte(c)
		p.pos++
	}
	return b.String(), true, nil
}

func (p *vdfParser) block(n *Node, top bool) error {
	for {
		key, quoted, err := p.token()
		if err != nil {
			return err
		}
		if !quoted {
			switch key {
			case "":
				if top {
					return nil
				}
				return fmt.Errorf("vdf: unexpected end of file")
			case "}":
				if top {
					return fmt.Errorf("vdf: unexpected }")
				}
				return nil
			default:
				return fmt.Errorf("vdf: unexpected %q", key)
			}
		}
		val, vquoted, err := p.token()
		if err != nil {
			return err
		}
		child := &Node{}
		switch {
		case vquoted:
			child.Value = val
		case val == "{":
			child.Children = map[string]*Node{}
			if err := p.block(child, false); err != nil {
				return err
			}
		default:
			return fmt.Errorf("vdf: no value for %q", key)
		}
		lk := strings.ToLower(key)
		if _, dup := n.Children[lk]; !dup {
			n.Order = append(n.Order, lk)
		}
		n.Children[lk] = child
	}
}
