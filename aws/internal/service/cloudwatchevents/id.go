package cloudwatchevents

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	eventBusARNPattern     = regexp.MustCompile(`^arn:aws[\w-]*:events:[a-z]{2}-[a-z]+-[\w-]+:[0-9]{12}:event-bus\/[\.\-_A-Za-z0-9]+$`)
	partnerEventBusPattern = regexp.MustCompile(`^aws\.partner(/[\.\-_A-Za-z0-9]+){2,}$`)
)

const permissionResourceIDSeparator = "/"

func PermissionCreateResourceID(eventBusName, statementID string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return statementID
	}

	parts := []string{eventBusName, statementID}
	id := strings.Join(parts, permissionResourceIDSeparator)

	return id
}

func PermissionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, permissionResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sSTATEMENTID or STATEMENTID", id, permissionResourceIDSeparator)
}

const ruleResourceIDSeparator = "/"

func RuleCreateResourceID(eventBusName, ruleName string) string {
	if eventBusName == "" || eventBusName == DefaultEventBusName {
		return ruleName
	}

	parts := []string{eventBusName, ruleName}
	id := strings.Join(parts, ruleResourceIDSeparator)

	return id
}

func RuleParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ruleResourceIDSeparator)

	if len(parts) == 1 && parts[0] != "" {
		return DefaultEventBusName, parts[0], nil
	}
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}
	if len(parts) > 2 {
		i := strings.LastIndex(id, ruleResourceIDSeparator)
		eventBusName := id[:i]
		ruleName := id[i+1:]
		if eventBusARNPattern.MatchString(eventBusName) && ruleName != "" {
			return eventBusName, ruleName, nil
		}
		if partnerEventBusPattern.MatchString(eventBusName) && ruleName != "" {
			return eventBusName, ruleName, nil
		}
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected EVENTBUSNAME%[2]sRULENAME or RULENAME", id, ruleResourceIDSeparator)
}

// Terraform resource IDs for Targets are not parseable as the separator used ("-") is also a valid character in both the rule name and the target ID.

const targetResourceIDSeparator = "-"
const targetImportIDSeparator = "/"

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
