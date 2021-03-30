package cloudwatchlogs

import (
	"fmt"
	"strings"
)

const queryDefinitionIDSeparator = "_"

func QueryDefinitionParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, queryDefinitionIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%s), expected <query-name>"+queryDefinitionIDSeparator+"<query-id>", id)
}
