// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"reflect"

	"github.com/hashicorp/go-hclog"
)

const (
	logModule = "provider." + subsystemName
)

func expandingLogLine(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Expanding", sourceType, targetType)
}

func flatteningLogLine(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Flattening", sourceType, targetType)
}

func ignoredFieldLogLine(sourceType reflect.Type, fieldName string) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping ignored field",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: fieldName,
	}
}

func mapBlockKeyFieldLogLine(sourceType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping map block key",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: MapBlockKey,
	}
}

func infoLogLine(message string, sourceType, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           message,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}
