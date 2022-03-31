package backup

const (
	frameworkStatusCompleted          = "COMPLETED"
	frameworkStatusCreationInProgress = "CREATE_IN_PROGRESS"
	frameworkStatusDeletionInProgress = "DELETE_IN_PROGRESS"
	frameworkStatusFailed             = "FAILED"
	frameworkStatusUpdateInProgress   = "UPDATE_IN_PROGRESS"
)

const (
	reportDeliveryChannelFormatCSV  = "CSV"
	reportDeliveryChannelFormatJSON = "JSON"
)

func reportDeliveryChannelFormat_Values() []string {
	return []string{
		reportDeliveryChannelFormatCSV,
		reportDeliveryChannelFormatJSON,
	}
}
