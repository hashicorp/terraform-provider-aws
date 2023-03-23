package connect

import "github.com/aws/aws-sdk-go/service/connect"

const InstanceStatusStatusNotFound = "ResourceNotFoundException"
const BotAssociationStatusNotFound = "ResourceNotFoundException"

const (
	ListInstancesMaxResults = 10
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 1000
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListContactFlows.html
	ListContactFlowsMaxResults = 60
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 1000
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListContactFlowModules.html
	ListContactFlowModulesMaxResults = 60
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 25
	ListBotsMaxResults = 25
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 1000
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListHoursOfOperations.html
	ListHoursOfOperationsMaxResults = 60
	// ListLambdaFunctionsMaxResults Valid Range: Minimum value of 1. Maximum value of 25.
	//https://docs.aws.amazon.com/connect/latest/APIReference/API_ListLambdaFunctions.html
	ListLambdaFunctionsMaxResults = 25
	// MaxResults Valid Range: Minimum value of 1. Maximum value of 1000
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListPrompts.html
	ListPromptsMaxResults = 60
	// ListQueueQuickConnectsMaxResults Valid Range: Minimum value of 1. Maximum value of 100.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListQueueQuickConnects.html
	ListQueueQuickConnectsMaxResults = 60
	// ListQueuesMaxResults Valid Range: Minimum value of 1. Maximum value of 1000.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListQueues.html
	ListQueuesMaxResults = 60
	// ListQuickConnectsMaxResults Valid Range: Minimum value of 1. Maximum value of 1000.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListQuickConnects.html
	ListQuickConnectsMaxResults = 60
	// ListRoutingProfileQueuesMaxResults Valid Range: Minimum value of 1. Maximum value of 100.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListRoutingProfileQueues.html
	ListRoutingProfileQueuesMaxResults = 60
	// ListRoutingProfilesMaxResults Valid Range: Minimum value of 1. Maximum value of 1000.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListRoutingProfiles.html
	ListRoutingProfilesMaxResults = 60
	// ListSecurityProfilePermissionsMaxResults Valid Range: Minimum value of 1. Maximum value of 1000.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListSecurityProfilePermissions.html
	ListSecurityProfilePermissionsMaxResults = 60
	// ListSecurityProfilesMaxResults Valid Range: Minimum value of 1. Maximum value of 1000.
	// https://docs.aws.amazon.com/connect/latest/APIReference/API_ListSecurityProfiles.html
	ListSecurityProfilesMaxResults = 60
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
