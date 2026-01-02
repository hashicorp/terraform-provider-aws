// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"time"
)

const (
	budgetsPropagationTimeout = 30 * time.Second // nosemgrep:ci.budgets-in-const-name, ci.budgets-in-var-name
	iamPropagationTimeout     = 2 * time.Minute
)
