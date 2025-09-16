// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceMemory         = newResourceMemory
	ResourceMemoryStrategy = newResourceMemoryStrategy

	FindMemoryByID         = findMemoryByID
	FindMemoryStrategyByID = findMemoryStrategyByID
)
