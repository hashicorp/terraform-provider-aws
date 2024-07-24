// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/go-hclog"
)

const (
	subsystemName = "autoflex"
)

const (
	logAttrKeySourceType = "autoflex.source.type"

	logAttrKeyTargetKind = "autoflex.target.kind"
	logAttrKeyTargetType = "autoflex.target.type"
)

const (
	defaultLogLevel = hclog.Error
	envvar          = "TF_LOG_AWS_AUTOFLEX"
)

func fullTypeName(t reflect.Type) string {
	if t == nil {
		return "<nil>"
	}
	if t.Kind() == reflect.Pointer {
		return "*" + fullTypeName(t.Elem())
	}
	if path := t.PkgPath(); path != "" {
		return fmt.Sprintf("%s.%s", path, t.Name())
	}
	return t.Name()
}
