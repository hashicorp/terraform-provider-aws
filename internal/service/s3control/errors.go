package s3control

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3control/#pkg-constants
const (
	errCodeNoSuchAccessPoint            = "NoSuchAccessPoint"
	errCodeNoSuchAccessPointPolicy      = "NoSuchAccessPointPolicy"
	errCodeNoSuchAsyncRequest           = "NoSuchAsyncRequest"
	errCodeNoSuchBucket                 = "NoSuchBucket"
	errCodeNoSuchLifecycleConfiguration = "NoSuchLifecycleConfiguration"
	errCodeNoSuchMultiRegionAccessPoint = "NoSuchMultiRegionAccessPoint"
	errCodeNoSuchOutpost                = "NoSuchOutpost"
)
