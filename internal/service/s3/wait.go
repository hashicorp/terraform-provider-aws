package s3

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	bucketCreatedTimeout = 2 * time.Minute
	propagationTimeout   = 1 * time.Minute
)

func retryWhenBucketNotFound(f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAWSErrCodeEquals(propagationTimeout, f, s3.ErrCodeNoSuchBucket)
}
