package backup

const (
	frameworkStatusCompleted          = "COMPLETED"
	frameworkStatusCreationInProgress = "CREATE_IN_PROGRESS"
	frameworkStatusDeletionInProgress = "DELETE_IN_PROGRESS"
	frameworkStatusFailed             = "FAILED"
	frameworkStatusUpdateInProgress   = "UPDATE_IN_PROGRESS"
)

const (
	reportPlanDeploymentStatusCompleted        = "COMPLETED"
	reportPlanDeploymentStatusCreateInProgress = "CREATE_IN_PROGRESS"
	reportPlanDeploymentStatusDeleteInProgress = "DELETE_IN_PROGRESS"
	reportPlanDeploymentStatusUpdateInProgress = "UPDATE_IN_PROGRESS"
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
	reportSettingTemplateJobReport                = "BACKUP_JOB_REPORT"
	reportSettingTemplateControlComplianceReport  = "CONTROL_COMPLIANCE_REPORT"
	reportSettingTemplateCopyJobReport            = "COPY_JOB_REPORT"
	reportSettingTemplateResourceComplianceReport = "RESOURCE_COMPLIANCE_REPORT"
	reportSettingTemplateRestoreJobReport         = "RESTORE_JOB_REPORT"
)

func reportSettingTemplate_Values() []string {
	return []string{
		reportSettingTemplateJobReport,
		reportSettingTemplateControlComplianceReport,
		reportSettingTemplateCopyJobReport,
		reportSettingTemplateResourceComplianceReport,
		reportSettingTemplateRestoreJobReport,
	}
}
