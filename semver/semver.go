// By Daniel Theophanes 2014

// semver holds standard version structure.
package semver

import (
	"bytes"
	"fmt"
)

type Version struct {
	Product    string
	Major      uint16
	Minor      uint16
	Patch      uint16
	PreRelease string
	InHex      bool
}

func (v *Version) String() string {
	if v == nil {
		return "0.0.0"
	}
	bb := &bytes.Buffer{}
	if len(v.Product) != 0 {
		fmt.Fprintf(bb, "%s ", v.Product)
	}
	if v.InHex {
		fmt.Fprintf(bb, "%X.%X.%X", v.Major, v.Minor, v.Patch)
	} else {
		fmt.Fprintf(bb, "%d.%d.%d", v.Major, v.Minor, v.Patch)
	}
	if len(v.PreRelease) != 0 {
		fmt.Fprintf(bb, "-%s", v.PreRelease)
	}
	return bb.String()
}

// Compare v1 to v2.
//   -1 if v1 <  v2
//    0 if v1 == v2
//   +1 if v1 >  v2
func (v1 *Version) Comp(v2 *Version) (r int) {
	if v1.Major != v2.Major {
		if v1.Major < v2.Major {
			return -1
		}
		return 1
	}
	if v1.Minor != v2.Minor {
		if v1.Minor < v2.Minor {
			return -1
		}
		return 1
	}
	if v1.Patch != v2.Patch {
		if v1.Patch < v2.Patch {
			return -1
		}
		return 1
	}
	return 0
}
