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

const (
	reportSettingTemplateBackupJobReport          = "BACKUP_JOB_REPORT"
	reportSettingTemplateControlComplianceReport  = "CONTROL_COMPLIANCE_REPORT"
	reportSettingTemplateCopyJobReport            = "COPY_JOB_REPORT"
	reportSettingTemplateResourceComplianceReport = "RESOURCE_COMPLIANCE_REPORT"
	reportSettingTemplateRestoreJobReport         = "RESTORE_JOB_REPORT"
)

func reportSettingTemplate_Values() []string {
	return []string{
		reportSettingTemplateBackupJobReport,
		reportSettingTemplateControlComplianceReport,
		reportSettingTemplateCopyJobReport,
		reportSettingTemplateResourceComplianceReport,
		reportSettingTemplateRestoreJobReport,
	}
}
