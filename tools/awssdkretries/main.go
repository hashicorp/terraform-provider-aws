// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/request"
)

func main() {
	maxBackoff := 300 * time.Second
	maxRetries := 25
	v2 := retry.NewExponentialJitterBackoff(maxBackoff)
	v1 := client.DefaultRetryer{
		NumMaxRetries:    maxRetries,
		MinRetryDelay:    client.DefaultRetryerMinRetryDelay,
		MinThrottleDelay: client.DefaultRetryerMinThrottleDelay,
		MaxRetryDelay:    maxBackoff,
		MaxThrottleDelay: maxBackoff,
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

		fmt.Printf("%d v1: %s, v2: %s\n", i, d1.String(), d2.String())
	}
}
