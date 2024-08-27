// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscalingplans

import (
	"fmt"
	"strconv"
	"strings"
)

const scalingPlanResourceIDSeparator = "/"

func scalingPlanCreateResourceID(scalingPlanName string, scalingPlanVersion int) string {
	return fmt.Sprintf("%[1]s%[2]s%[3]d", scalingPlanName, scalingPlanResourceIDSeparator, scalingPlanVersion)
}

func scalingPlanParseResourceID(id string) (string, int, error) {
	parts := strings.Split(id, scalingPlanResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		v, err := strconv.Atoi(parts[1])

		if err != nil {
			return "", 0, err
		}

		return parts[0], v, nil
	}

	return "", 0, fmt.Errorf("unexpected format for ID (%[1]s), expected SCALINGPLANNAME%[2]sSCALINGPLANVERSION", id, scalingPlanResourceIDSeparator)
}
