package opensearchserverless

// Exports for use in tests only.
var (
	ResourceAccessPolicy   = newResourceAccessPolicy
	ResourceCollection     = newResourceCollection
	ResourceSecurityConfig = newResourceSecurityConfig
	ResourceSecurityPolicy = newResourceSecurityPolicy
	ResourceVPCEndpoint    = newResourceVPCEndpoint

	FindAccessPolicyByNameAndType   = findAccessPolicyByNameAndType
	FindCollectionByID              = findCollectionByID
	FindSecurityConfigByID          = findSecurityConfigByID
	FindSecurityPolicyByNameAndType = findSecurityPolicyByNameAndType
	FindVPCEndpointByID             = findVPCEndpointByID
)
