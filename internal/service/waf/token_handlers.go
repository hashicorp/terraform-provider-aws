// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type retryer struct {
	connection *waf.Client
}

type withTokenFunc func(token *string) (interface{}, error)

func (t *retryer) RetryWithToken(ctx context.Context, f withTokenFunc) (interface{}, error) {
	const (
		key = "WafRetryer"
	)
	conns.GlobalMutexKV.Lock(key)
	defer conns.GlobalMutexKV.Unlock(key)

	const (
		timeout = 15 * time.Minute
	)
	return tfresource.RetryWhenIsA[*awstypes.WAFStaleDataException](ctx, timeout, func() (interface{}, error) {
		input := &waf.GetChangeTokenInput{}
		output, err := t.connection.GetChangeToken(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("acquiring WAF change token: %w", err)
		}

		return f(output.ChangeToken)
	})
}

func newRetryer(conn *waf.Client) *retryer {
	return &retryer{connection: conn}
}
