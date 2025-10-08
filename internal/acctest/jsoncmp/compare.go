// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package jsoncmp

// Inspired by https://github.com/hgsgtk/jsoncmp.

import (
	"encoding/json"
	"fmt"

	"github.com/google/go-cmp/cmp"
)

func Diff(x, y string) string {
	xform := cmp.Transformer("jsoncmp", func(s string) (m map[string]any) {
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			panic(fmt.Sprintf("json.Unmarshal(%s): %s", s, err))
		}
		return m
	})
	opt := cmp.FilterPath(func(p cmp.Path) bool {
		for _, ps := range p {
			if tr, ok := ps.(cmp.Transform); ok && tr.Option() == xform {
				return false
			}
		}
		return true
	}, xform)

	return cmp.Diff(x, y, opt)
}
