// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package devopsagent

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/devopsagent/document"
)

// jsonToDocument converts a JSON string to a document.Interface suitable for SDK calls.
func jsonToDocument(jsonStr string) (document.Interface, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("parsing metadata JSON: %w", err)
	}
	return document.NewLazyDocument(raw), nil
}

// documentToJSON converts a document.Interface from SDK responses to a JSON string.
func documentToJSON(doc document.Interface) (string, error) {
	var raw map[string]interface{}
	if err := doc.UnmarshalSmithyDocument(&raw); err != nil {
		return "", fmt.Errorf("unmarshaling metadata document: %w", err)
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("marshaling metadata to JSON: %w", err)
	}
	return string(b), nil
}
