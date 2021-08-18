package forwarded

import (
	"net/http"
	"strings"
	"unicode"
)

// ParseHeaders accepts RFC7239 Forwarded: or traditional ad-hoc X-Forwarded-For
func ParseHeaders(headers http.Header) (parsed []map[string]string) {
	xf := headers["Forwarded"]
	if len(xf) > 0 {
		for _, f := range xf {
			tp := ParseForwarded(f)
			if len(tp) > 0 {
				parsed = append(parsed, tp...)
			}
		}
		return
	}
	xf = headers["X-Forwarded-For"]
	for _, f := range xf {
		parts := strings.Split(f, ",")
		for _, pf := range parts {
			pf = strings.TrimSpace(pf)
			parsed = append(parsed, map[string]string{"for": pf})
		}
	}
	return
}

// ParseForwarded handles RFC7239 Forwarded: header
// https://datatracker.ietf.org/doc/html/rfc7239
// Return list of map[string]string
// The first result should be the originator out there on the net; later entries may be a chain of proxies.
func ParseForwarded(headerValue string) (parsed []map[string]string) {
	// groups of
	// a=b;c="d"
	// separated by comma and maybe whitespace

	for len(headerValue) > 0 {
		group, rem := readGroup(headerValue)
		if group != nil {
			parsed = append(parsed, group)
		}
		headerValue = rem
	}
	return
}

func FirstForwardedFor(parsed []map[string]string) (ffor string) {
	if len(parsed) == 0 {
		return
	}
	for k, v := range parsed[0] {
		if strings.ToLower(k) == "for" {
			return v
		}
	}
	return
}

func readGroup(x string) (group map[string]string, remainder string) {
	var pos int
	var c rune
	// skip any inital space
	for pos, c = range x {
		if !unicode.IsSpace(c) {
			break
		}
	}
	if pos >= len(x) {
		return // err
	}
	for {
		var dpos int
		group, dpos = readKeyValue(x[pos:], group)
		if dpos == 0 {
			panic("wat")
		}
		pos += dpos
		if pos >= len(x) {
			return
		}
		if x[pos] == ',' {
			// group break
			remainder = x[pos+1:]
			return
		}
		if x[pos] == ';' {
			// k=v;...
			pos++
			if pos >= len(x) {
				return
			}
		}
	}
}

func readKeyValue(x string, groupIn map[string]string) (group map[string]string, pos int) {
	group = groupIn
	mode := 0
	start := 0
	var key string
	var value string
	prevbs := false
	var c rune
	didBreak := false
	for pos, c = range x {
		if mode == 0 {
			if unicode.IsSpace(c) {
				continue
			}
			start = pos
			mode = 1
			// fall through
		}
		if mode == 1 {
			// key
			if c == '=' {
				key = x[start:pos]
				mode = 2
				continue
			}
		}
		if mode == 2 {
			// previous was '='
			// start value
			if c == '"' {
				// enter quoted string
				mode = 3
				continue
			} else {
				// enter token
				start = pos
				mode = 6
				// fall through
			}
		}
		if mode == 3 {
			// prev was '"'
			start = pos
			mode = 4
			// fall through
		}
		if mode == 4 {
			// quoted string body
			bs := c == '\\' && !prevbs
			if c == '"' && !prevbs {
				// done
				value = x[start:pos]
				// advance one more to discard '"'
				mode = 9
				continue
			}
			prevbs = bs
			continue
		}

		if mode == 6 {
			// token
			if c == ';' || c == ',' {
				// done
				value = x[start:pos]
				// leave pos at split char
				mode = 9 // success
				didBreak = true
				break
			}
		}

		if mode == 9 {
			// value found
			didBreak = true
			break
		}
	}
	if !didBreak {
		pos = len(x)
	}
	if mode == 6 {
		// token through to end of string
		value = x[start:]
		mode = 9 // success
	}
	if mode == 9 {
		if group == nil {
			group = make(map[string]string)
		}
		group[key] = value
	}
	return
}
