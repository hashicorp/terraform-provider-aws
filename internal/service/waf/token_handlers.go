// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type WafRetryer struct {
	Connection *waf.Client
}

type withTokenFunc func(token *string) (interface{}, error)

func (t *WafRetryer) RetryWithToken(ctx context.Context, f withTokenFunc) (interface{}, error) {
	conns.GlobalMutexKV.Lock("WafRetryer")
	defer conns.GlobalMutexKV.Unlock("WafRetryer")

	var out interface{}
	var tokenOut *waf.GetChangeTokenOutput
	err := retry.RetryContext(ctx, 15*time.Minute, func() *retry.RetryError {
		var err error
		tokenOut, err = t.Connection.GetChangeToken(ctx, &waf.GetChangeTokenInput{})

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
		tokenOut, err = t.Connection.GetChangeToken(ctx, &waf.GetChangeTokenInput{})

		if err != nil {
			return nil, fmt.Errorf("getting WAF change token: %w", err)
		}

		out, err = f(tokenOut.ChangeToken)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

func NewRetryer(conn *waf.Client) *WafRetryer {
	return &WafRetryer{Connection: conn}
}
