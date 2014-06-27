package semver

import (
	"reflect"
	"testing"
)

var list = map[string]*Version{
	"12.21.3":      &Version{Major: 12, Minor: 21, Patch: 3},
	"1.2.33-":      &Version{Major: 1, Minor: 2, Patch: 33},
	"1.2.3-rc2":    &Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "rc2"},
	"1.2.3-rc2-xy": &Version{Major: 1, Minor: 2, Patch: 3, PreRelease: "rc2-xy"},
}

func TestParse(t *testing.T) {
	for s, ver := range list {
		tryVer, err := Parse(s)
		if err != nil {
			t.Errorf("Failed to parse: %v", err)
		}
		if reflect.DeepEqual(ver, tryVer) == false {
			t.Errorf("Different: want <%v> got <%v>.", ver, tryVer)
		}
	}
}
