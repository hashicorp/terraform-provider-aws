// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"encoding/json"
	"log"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func suppressEquivalentTaskSettings(k, state, proposed string, d *schema.ResourceData) bool {
	if !json.Valid([]byte(state)) || !json.Valid([]byte(proposed)) {
		return state == proposed
	}

	var stateMap, proposedMap map[string]any

	if err := json.Unmarshal([]byte(state), &stateMap); err != nil {
		log.Printf("[ERROR] failed to unmarshal task settings JSON: %v", err)
		return false
	}

	if s, ok := stateMap["Logging"]; ok {
		stateLogging := s.(map[string]any)
		if stateLogging != nil {
			delete(stateLogging, "CloudWatchLogGroup")
			delete(stateLogging, "CloudWatchLogStream")
		}
		normalizeLogComponents(stateLogging)
	}

	if err := json.Unmarshal([]byte(proposed), &proposedMap); err != nil {
		log.Printf("[ERROR] failed to unmarshal task settings JSON: %v", err)
		return false
	}

	if p, ok := proposedMap["Logging"]; ok {
		proposedLogging := p.(map[string]any)

		normalizeLogComponents(proposedLogging)
	}

	return taskSettingsEqual(stateMap, proposedMap)
}

func taskSettingsEqual(state, proposed any) bool {
	if proposed == nil {
		return true
	}

	switch x := state.(type) {
	case bool:
		p := proposed.(bool)
		return x == p

	case float64:
		p := proposed.(float64)
		return x == p

	case string:
		p := proposed.(string)
		return x == p

	case map[string]any:
		proposedMap := proposed.(map[string]any)
		for k, v := range x {
			if !taskSettingsEqual(v, proposedMap[k]) {
				return false
			}
			delete(proposedMap, k)
		}
		return len(proposedMap) == 0
	}
	return false
}

func normalizeLogComponents(m map[string]any) {
	if m == nil {
		return
	}

	components, ok := m["LogComponents"]
	if !ok {
		return
	}

	newComponents := make(map[string]any, len(components.([]any)))
	for _, c := range components.([]any) {
		component := c.(map[string]any)
		id := component["Id"].(string)
		newComponents[id] = component["Severity"].(string)
	}
	m["LogComponents"] = newComponents
}

func validateReplicationSettings(i any, path cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	v, ok := i.(string)
	if !ok {
		return append(diags, errs.NewIncorrectValueTypeAttributeError(path, "string"))
	}

	var m map[string]any

	if err := json.Unmarshal([]byte(v), &m); err != nil {
		return append(diags, errs.NewInvalidValueAttributeError(path, "Unable to parse as JSON"))
	}

	if l, ok := m["Logging"].(map[string]any); ok {
		if _, ok := l["EnableLogContext"]; ok {
			if enabled, ok := l["EnableLogging"]; !ok || !enabled.(bool) {
				diags = append(diags, errs.NewInvalidValueAttributeError(path, "The parameter Logging.EnableLogContext is not allowed when Logging.EnableLogging is not set to true."))
			}
		}

		if _, ok := l["CloudWatchLogGroup"]; ok {
			diags = append(diags, errs.NewInvalidValueAttributeError(path, "The parameter Logging.CloudWatchLogGroup is read-only and cannot be set."))
		}
		if _, ok := l["CloudWatchLogStream"]; ok {
			diags = append(diags, errs.NewInvalidValueAttributeError(path, "The parameter Logging.CloudWatchLogStream is read-only and cannot be set."))
		}
	}

	return diags
}

func normalizeReplicationSettings(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var m map[string]any

	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return s, err
	}

	// If EnableLogging is false, set EnableLogContext to false unless it is explicitly set.
	// Normally, if EnableLogContext is not set, it uses the existing value.
	if l, ok := m["Logging"].(map[string]any); ok {
		if enabled, ok := l["EnableLogging"]; ok && !enabled.(bool) {
			delete(l, "EnableLogContext")
			if _, ok := l["EnableLogContext"]; !ok {
				l["EnableLogContext"] = false
				b, err := json.Marshal(m)
				if err != nil {
					return s, err
				}
				s = string(b)
			}
		}
	}

	return s, nil
}
