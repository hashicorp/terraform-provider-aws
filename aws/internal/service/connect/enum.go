package connect

import "github.com/aws/aws-sdk-go/service/connect"

const InstanceStatusStatusNotFound = "ResourceNotFoundException"

func InstanceAttributeMapping() map[string]string {
	return map[string]string{
		connect.InstanceAttributeTypeAutoResolveBestVoices: "auto_resolve_best_voices_enabled",
		connect.InstanceAttributeTypeContactflowLogs:       "contact_flow_logs_enabled",
		connect.InstanceAttributeTypeContactLens:           "contact_lens_enabled",
		connect.InstanceAttributeTypeEarlyMedia:            "early_media_enabled",
		connect.InstanceAttributeTypeInboundCalls:          "inbound_calls_enabled",
		connect.InstanceAttributeTypeOutboundCalls:         "outbound_calls_enabled",
		connect.InstanceAttributeTypeUseCustomTtsVoices:    "use_custom_tts_voices_enabled",
	}
}
