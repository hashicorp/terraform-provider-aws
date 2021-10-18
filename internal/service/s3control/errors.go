package s3control

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3control/#pkg-constants
//nolint:deadcode,varcheck // These constants are missing from the AWS SDK
const (
	errCodeNoSuchAccessPoint       = "NoSuchAccessPoint"
	errCodeNoSuchAccessPointPolicy = "NoSuchAccessPointPolicy"
)
