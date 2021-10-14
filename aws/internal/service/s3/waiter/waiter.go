package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	BucketCreatedTimeout = 2 * time.Minute
	PropagationTimeout   = 1 * time.Minute
)

// RetryWhenBucketNotFound retries the specified function if the returned error indicates that a bucket is not found.
// If the retries time out the specified function is called one last time.
func RetryWhenBucketNotFound(f func() (interface{}, error)) (interface{}, error) {
	return tfresource.RetryWhenAwsErrCodeEquals(PropagationTimeout, f, s3.ErrCodeNoSuchBucket)
}
