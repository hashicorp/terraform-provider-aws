// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"github.com/hashicorp/terraform-plugin-testing/config"
)

func StringVariable[T ~string](value T) config.Variable {
	return config.StringVariable(string(value))
}
