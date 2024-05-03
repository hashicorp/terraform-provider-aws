// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"time"
)

const (
	aliasNamePrefix = "alias/"
	cmkAliasPrefix  = aliasNamePrefix + "aws/"
)

const (
	policyNameDefault = "default"
)

const (
	propagationTimeout = 2 * time.Minute
)
