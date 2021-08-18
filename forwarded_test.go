package forwarded

import (
	"encoding/json"
	"testing"
)

func TestParseForwarded(t *testing.T) {
	// example strings from rfc https://datatracker.ietf.org/doc/html/rfc7239
	// map from header to json equivalent
	examples := map[string]string{
		`for="_gazonk"`: `[{"for":"_gazonk"}]`,

		`For="[2001:db8:cafe::17]:4711"`: `[{"For":"[2001:db8:cafe::17]:4711"}]`,

		`for=192.0.2.60;proto=http;by=203.0.113.43`: `[{"for":"192.0.2.60","proto":"http","by":"203.0.113.43"}]`,

		`for=192.0.2.43, for=198.51.100.17`: `[{"for":"192.0.2.43"}, {"for":"198.51.100.17"}]`,

		`for=192.0.2.43,for="[2001:db8:cafe::17]",for=unknown`: `[{"for":"192.0.2.43"},{"for":"[2001:db8:cafe::17]"},{"for":"unknown"}]`,
	}

	for h, j := range examples {
		t.Run(h, func(t *testing.T) {
			hp := ParseForwarded(h)
			var jp []map[string]string
			err := json.Unmarshal([]byte(j), &jp)
			if err != nil {
				t.Fatal(err)
			}
			if len(hp) != len(jp) {
				t.Errorf("expected [%d] but got %d", len(jp), len(hp))
				return
			}
			for i, jv := range jp {
				hv := hp[i]
				mapeq(t, jv, hv)
			}
		})
	}
}

func mapeq(t *testing.T, a, b map[string]string) {
	for ak, av := range a {
		bv, ok := b[ak]
		if !ok {
			t.Errorf("in a but not b: %#v", ak)
		} else if av != bv {
			t.Errorf("a[%s] is %#v but b has %#v", ak, av, bv)
		}
	}
	for bk := range b {
		_, ok := a[bk]
		if !ok {
			t.Errorf("in b but not a: %#v", bk)
		}
	}
}
