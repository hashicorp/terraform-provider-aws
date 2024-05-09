// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
	smithy "github.com/aws/smithy-go"
)

func TestBackoffDelayer(t *testing.T) {
	t.Parallel()

	maxBackoff := 300 * time.Second
	maxRetries := 25
	v2 := retry.NewExponentialJitterBackoff(maxBackoff)
	v1 := awsclient.DefaultRetryer{
		NumMaxRetries:    maxRetries,
		MaxRetryDelay:    maxBackoff,
		MaxThrottleDelay: maxBackoff,
	}
	v1compat := &v1CompatibleBackoff{
		maxRetryDelay: maxBackoff,
	}

	err1 := awserr.New("ThrottlingException", "Rate exceeded", nil)
	err2 := &smithy.GenericAPIError{
		Code:    "ThrottlingException",
		Message: "Rate exceeded",
	}
	req := request.Request{
		Error: err1,
		HTTPResponse: &http.Response{
			StatusCode: http.StatusBadRequest,
		},
	}
	for i := 0; i < maxRetries; i++ {
		d1 := v1.RetryRules(&req)
		req.RetryCount++
		d2, _ := v2.BackoffDelay(i, err2)
		d1compat, _ := v1compat.BackoffDelay(i, err2)

		t.Logf("%d v1: %s, v2: %s v1compat: %s\n", i, d1.String(), d2.String(), d1compat.String())
	}
}
