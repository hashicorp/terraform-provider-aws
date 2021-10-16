package connect

import "github.com/aws/aws-sdk-go/service/connect"

const InstanceStatusStatusNotFound = "ResourceNotFoundException"

const (
	ListInstancesMaxResults = 10
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 1000
	ListContactFlowsMaxResults = 60
)

func InstanceAttributeMapping() map[string]string {
	return map[string]string{
		connect.InstanceAttributeTypeAutoResolveBestVoices: "auto_resolve_best_voices_enabled",
		connect.InstanceAttributeTypeContactflowLogs:       "contact_flow_logs_enabled",
		connect.InstanceAttributeTypeContactLens:           "contact_lens_enabled",
		connect.InstanceAttributeTypeEarlyMedia:            "early_media_enabled",
		connect.InstanceAttributeTypeInboundCalls:          "inbound_calls_enabled",
		connect.InstanceAttributeTypeOutboundCalls:         "outbound_calls_enabled",
		// Pre-release feature requiring allow-list from AWS. Removing all functionality until feature is GA
		//connect.InstanceAttributeTypeUseCustomTtsVoices:    "use_custom_tts_voices_enabled",
	}
}
