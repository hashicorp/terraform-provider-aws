// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"time"
)

const (
	AliasNamePrefix = "alias/"
)

const (
	PolicyNameDefault = "default"
)

const (
	propagationTimeout = 2 * time.Minute
)
