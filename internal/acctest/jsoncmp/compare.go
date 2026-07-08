// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package jsoncmp

// Inspired by https://github.com/hgsgtk/jsoncmp.

import (
	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func Diff(x, y string) string {
	xform := cmp.Transformer("jsoncmp", func(s string) map[string]any {
		var m map[string]any
		tfjson.DecodeFromString(s, &m)
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
