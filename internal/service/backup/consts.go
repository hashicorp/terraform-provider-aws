// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import "time"

const (
	propagationTimeout = 2 * time.Minute
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
