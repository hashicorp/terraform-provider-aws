// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type WafRegionalRetryer struct {
	Connection *wafregional.Client
	Region     string
}

type withRegionalTokenFunc func(token *string) (interface{}, error)

func (t *WafRegionalRetryer) RetryWithToken(ctx context.Context, f withRegionalTokenFunc) (interface{}, error) {
	conns.GlobalMutexKV.Lock(t.Region)
	defer conns.GlobalMutexKV.Unlock(t.Region)

	var out interface{}
	var tokenOut *wafregional.GetChangeTokenOutput
	err := retry.RetryContext(ctx, 15*time.Minute, func() *retry.RetryError {
		var err error

		tokenOut, err = t.Connection.GetChangeToken(ctx, &wafregional.GetChangeTokenInput{})
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("Failed to acquire change token: %w", err))
		}

		out, err = f(tokenOut.ChangeToken)
		if err != nil {
			if errs.IsA[*awstypes.WAFStaleDataException](err) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		tokenOut, err = t.Connection.GetChangeToken(ctx, &wafregional.GetChangeTokenInput{})

		if err != nil {
			return nil, fmt.Errorf("getting WAF Regional change token: %w", err)
		}

		out, err = f(tokenOut.ChangeToken)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NewRetryer(conn *wafregional.Client, region string) *WafRegionalRetryer {
	return &WafRegionalRetryer{Connection: conn, Region: region}
}
