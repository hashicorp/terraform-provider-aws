package s3

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	ErrCodeInvalidBucketState                   = "InvalidBucketState"
	ErrCodeNoSuchConfiguration                  = "NoSuchConfiguration"
	ErrCodeNoSuchCORSConfiguration              = "NoSuchCORSConfiguration"
	ErrCodeNoSuchLifecycleConfiguration         = "NoSuchLifecycleConfiguration"
	ErrCodeNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	ErrCodeNoSuchWebsiteConfiguration           = "NoSuchWebsiteConfiguration"
	ErrCodeOperationAborted                     = "OperationAborted"
	ErrCodeReplicationConfigurationNotFound     = "ReplicationConfigurationNotFoundError"
)
