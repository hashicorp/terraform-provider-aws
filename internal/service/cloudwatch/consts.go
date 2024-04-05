// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

const (
	lowSampleCountPercentilesEvaluate          = "evaluate"
	lowSampleCountPercentilesmissingDataIgnore = "ignore"
)

func lowSampleCountPercentiles_Values() []string {
	return []string{
		lowSampleCountPercentilesEvaluate,
		lowSampleCountPercentilesmissingDataIgnore,
	}
}

const (
	missingDataBreaching    = "breaching"
	missingDataIgnore       = "ignore"
	missingDataMissing      = "missing"
	missingDataNotBreaching = "notBreaching"
)

func missingData_Values() []string {
	return []string{
		missingDataBreaching,
		missingDataIgnore,
		missingDataMissing,
		missingDataNotBreaching,
	}
}
