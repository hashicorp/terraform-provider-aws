// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package actions

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TimeoutOr(timeout types.Int64, defaultValue time.Duration) time.Duration {
	v := defaultValue
	if !timeout.IsNull() {
		v = time.Duration(timeout.ValueInt64()) * time.Second
	}
	return v
}
