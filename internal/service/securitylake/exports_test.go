// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securitylake

// Exports for use in tests only.
var (
	ResourceAWSLogSource           = newAWSLogSourceResource
	ResourceCustomLogSource        = newCustomLogSourceResource
	ResourceDataLake               = newDataLakeResource
	ResourceSubscriber             = newSubscriberResource
	ResourceSubscriberNotification = newSubscriberNotificationResource

	FindAWSLogSourceBySourceName             = findAWSLogSourceBySourceName
	FindCustomLogSourceBySourceName          = findCustomLogSourceBySourceName
	FindDataLakeByARN                        = findDataLakeByARN
	FindDataLakes                            = findDataLakes
	FindSubscriberByID                       = findSubscriberByID
	FindSubscriberNotificationBySubscriberID = findSubscriberNotificationBySubscriberID
)
