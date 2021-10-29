package connect

import (
	"fmt"
	"strings"
)

const botV1AssociationIDSeparator = ":"

func BotV1AssociationParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, botV1AssociationIDSeparator, 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of Connect Bot V1 Association ID (%s), expected instanceID:name:region", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func BotV1AssociationCreateResourceID(instanceID string, botName string, region string) string {
	parts := []string{instanceID, botName, region}
	id := strings.Join(parts, botV1AssociationIDSeparator)

	return id
}
