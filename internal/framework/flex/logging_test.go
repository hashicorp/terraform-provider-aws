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

func infoLogLine(message string, sourceType, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           message,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}
