package backup

const (
	frameworkStatusCompleted          = "COMPLETED"
	frameworkStatusCreationInProgress = "CREATE_IN_PROGRESS"
	frameworkStatusDeletionInProgress = "DELETE_IN_PROGRESS"
	frameworkStatusFailed             = "FAILED"
	frameworkStatusUpdateInProgress   = "UPDATE_IN_PROGRESS"
)

const (
	reportDeliveryFormatCSV  = "CSV"
	reportDeliveryFormatJSON = "JSON"
)

func reportDeliveryFormat_Values() []string {
	return []string{
		reportDeliveryFormatCSV,
		reportDeliveryFormatJSON,
	}
}
