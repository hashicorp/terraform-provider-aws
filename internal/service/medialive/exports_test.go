// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

// Exports for use in tests only.
var (
	ResourceChannel          = resourceChannel
	ResourceMultiplexProgram = newMultiplexProgramResource

	FindChannelByID          = findChannelByID
	FindMultiplexProgramByID = findMultiplexProgramByID
	ParseMultiplexProgramID  = parseMultiplexProgramID
)
