// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type retryer struct {
	connection *wafregional.Client
	region     string
}

type withTokenFunc func(token *string) (interface{}, error)

func (t *retryer) RetryWithToken(ctx context.Context, f withTokenFunc) (interface{}, error) {
	key := "WafRetryer-" + t.region
	conns.GlobalMutexKV.Lock(key)
	defer conns.GlobalMutexKV.Unlock(key)

	const (
		timeout = 15 * time.Minute
	)
	return tfresource.RetryWhenIsA[*awstypes.WAFStaleDataException](ctx, timeout, func() (interface{}, error) {
		input := &wafregional.GetChangeTokenInput{}
		output, err := t.connection.GetChangeToken(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("acquiring WAF Regional change token: %w", err)
		}

		return f(output.ChangeToken)
	})
}

func newRetryer(conn *wafregional.Client, region string) *retryer {
	return &retryer{connection: conn, region: region}
}
