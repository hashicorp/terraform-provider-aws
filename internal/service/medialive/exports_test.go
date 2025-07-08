// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

// Exports for use in tests only.
var (
	ResourceChannel            = resourceChannel
	ResourceInput              = resourceInput
	ResourceInputSecurityGroup = resourceInputSecurityGroup
	ResourceMultiplexProgram   = newMultiplexProgramResource

	FindChannelByID            = findChannelByID
	FindInputByID              = findInputByID
	FindInputSecurityGroupByID = findInputSecurityGroupByID
	FindMultiplexProgramByID   = findMultiplexProgramByID
	ParseMultiplexProgramID    = parseMultiplexProgramID
)
