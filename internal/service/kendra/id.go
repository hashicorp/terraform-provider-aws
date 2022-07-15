package kendra

import (
	"fmt"
	"strings"
)

func ExperienceParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format EXPERIENCE_ID/INDEX_ID")
	}

	return parts[0], parts[1], nil
}

func FaqParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format FAQ_ID/INDEX_ID")
	}

	return parts[0], parts[1], nil
}

func QuerySuggestionsBlockListParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format QUERY_SUGGESTIONS_BLOCK_LIST_ID/INDEX_ID")
	}

	return parts[0], parts[1], nil
}

func ThesaurusParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, "/")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("please make sure ID is in format THESAURUS_ID/INDEX_ID")
	}

	return parts[0], parts[1], nil
}
