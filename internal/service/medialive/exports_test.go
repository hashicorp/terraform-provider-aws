// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package medialive

// Exports for use in tests only.
var (
	ResourceChannel            = resourceChannel
	ResourceInput              = resourceInput
	ResourceInputSecurityGroup = resourceInputSecurityGroup
	ResourceMultiplex          = resourceMultiplex
	ResourceMultiplexProgram   = newMultiplexProgramResource

	FindChannelByID            = findChannelByID
	FindInputByID              = findInputByID
	FindInputSecurityGroupByID = findInputSecurityGroupByID
	FindMultiplexByID          = findMultiplexByID
	FindMultiplexProgramByID   = findMultiplexProgramByID
	ParseMultiplexProgramID    = parseMultiplexProgramID
)
