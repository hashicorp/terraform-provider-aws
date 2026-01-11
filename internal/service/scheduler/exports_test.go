// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package scheduler

// Exports for use in tests only.
var (
	ResourceSchedule      = resourceSchedule
	ResourceScheduleGroup = resourceScheduleGroup

	FindScheduleByTwoPartKey = findScheduleByTwoPartKey
	FindScheduleGroupByName  = findScheduleGroupByName

	ScheduleResourceIDFromARN = scheduleResourceIDFromARN
	ScheduleParseResourceID   = scheduleParseResourceID
)
