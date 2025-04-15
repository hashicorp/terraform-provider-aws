// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	tfmaps "github.com/hashicorp/terraform-provider-aws/internal/maps"
)

var (
	performanceTargetLevels = map[int32]string{
		1:   "LOW_COST",
		25:  "ECONOMICAL",
		50:  "BALANCED",
		75:  "RESOURCEFUL",
		100: "HIGH_PERFORMANCE",
	}
)

func performanceTargetLevel_Values() []string {
	return tfmaps.Values(performanceTargetLevels)
}
