// Copyright (c) 2011, The pg Authors. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// +build darwin freebsd linux netbsd openbsd solaris

package pg

import "os/user"

func userCurrent() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}
