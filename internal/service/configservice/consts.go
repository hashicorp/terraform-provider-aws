package configservice

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	ResNameAggregateAuthorization      = "Aggregate Authorization"
	ResNameConfigurationAggregator     = "Configuration Aggregator"
	ResNameConfigurationRecorderStatus = "Configuration Recorder Status"
	ResNameConfigurationRecorder       = "Configuration Recorder"
	ResNameDeliveryChannel             = "Delivery Channel"
	ResNameOrganizationManagedRule     = "Organization Managed Rule"
	ResNameOrganizationCustomRule      = "Organization Custom Rule"
	ResNameRemediationConfiguration    = "Remediation Configuration"
)
