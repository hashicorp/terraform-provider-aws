// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"fmt"
	"strings"
)

// Terraform resource IDs for Targets are not parseable as the separator used ("-") is also a valid character in both the rule name and the target ID.

const (
	targetResourceIDSeparator = "-"
	targetImportIDSeparator   = "/"
)

func TargetCreateResourceID(eventBusName, ruleName, targetID string) string {
	var parts []string

	if eventBusName == "" || eventBusName == DefaultEventBusName {
		parts = []string{ruleName, targetID}
	} else {
		parts = []string{eventBusName, ruleName, targetID}
	}

	id := strings.Join(parts, targetResourceIDSeparator)

	return id
}

func TargetParseImportID(id string) (string, string, string, error) {
	parts := strings.Split(id, targetImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return DefaultEventBusName, parts[0], parts[1], nil
	}
	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}
	if len(parts) > 3 {
		iTarget := strings.LastIndex(id, targetImportIDSeparator)
		targetID := id[iTarget+1:]
		iRule := strings.LastIndex(id[:iTarget], targetImportIDSeparator)
		eventBusName := id[:iRule]
		ruleName := id[iRule+1 : iTarget]
		if eventBusARNPattern.MatchString(eventBusName) && ruleName != "" && targetID != "" {
			return eventBusName, ruleName, targetID, nil
		}
		if partnerEventBusPattern.MatchString(eventBusName) && ruleName != "" && targetID != "" {
			return eventBusName, ruleName, targetID, nil
		}
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sRULENAME%[2]sTARGETID or RULENAME%[2]sTARGETID", id, targetImportIDSeparator)
}
