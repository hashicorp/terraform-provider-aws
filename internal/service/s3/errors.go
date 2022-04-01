package s3

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	ErrCodeInvalidBucketState                        = "InvalidBucketState"
	ErrCodeInvalidRequest                            = "InvalidRequest"
	ErrCodeMethodNotAllowed                          = "MethodNotAllowed"
	ErrCodeNoSuchBucketPolicy                        = "NoSuchBucketPolicy"
	ErrCodeNoSuchConfiguration                       = "NoSuchConfiguration"
	ErrCodeNoSuchCORSConfiguration                   = "NoSuchCORSConfiguration"
	ErrCodeNoSuchLifecycleConfiguration              = "NoSuchLifecycleConfiguration"
	ErrCodeNoSuchPublicAccessBlockConfiguration      = "NoSuchPublicAccessBlockConfiguration"
	ErrCodeNoSuchWebsiteConfiguration                = "NoSuchWebsiteConfiguration"
	ErrCodeNotImplemented                            = "NotImplemented"
	ErrCodeObjectLockConfigurationNotFound           = "ObjectLockConfigurationNotFoundError"
	ErrCodeOperationAborted                          = "OperationAborted"
	ErrCodeReplicationConfigurationNotFound          = "ReplicationConfigurationNotFoundError"
	ErrCodeServerSideEncryptionConfigurationNotFound = "ServerSideEncryptionConfigurationNotFoundError"
	ErrCodeUnsupportedArgument                       = "UnsupportedArgument"

	// ErrCodeXNotImplemented is returned from Third Party S3 implementations
	// and so far has been noticed with calls to GetBucketWebsite.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14645
	ErrCodeXNotImplemented = "XNotImplemented"
)
