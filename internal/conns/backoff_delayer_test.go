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
	v2compat := &v1CompatibleBackoff{
		maxRetryDelay: maxBackoff,
	}

	err := awserr.New("ThrottlingException", "Rate exceeded", nil)
	req := request.Request{
		Error: err,
		HTTPResponse: &http.Response{
			StatusCode: 400,
		},
	}
	for i := 0; i < maxRetries; i++ {
		d1 := v1.RetryRules(&req)
		req.RetryCount++
		d2, _ := v2.BackoffDelay(i, err)
		d2compat, _ := v2compat.BackoffDelay(i, err)

		t.Logf("%d v1: %s, v2: %s v2compat: %s\n", i, d1.String(), d2.String(), d2compat.String())
	}
}
