package kendra

import (
	"fmt"
	"strings"
)

func ThesaurusParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format THESAURUS_ID/INDEX_ID")
	}

	return parts[0], parts[1], nil
}
