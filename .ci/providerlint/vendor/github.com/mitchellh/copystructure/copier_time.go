// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package copystructure

import (
	"reflect"
	"time"
)

func init() {
	Copiers[reflect.TypeOf(time.Time{})] = timeCopier
}

func timeCopier(v interface{}) (interface{}, error) {
	// Just... copy it.
	return v.(time.Time), nil
}
