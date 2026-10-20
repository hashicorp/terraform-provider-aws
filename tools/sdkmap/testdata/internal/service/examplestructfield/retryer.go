// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplestructfield

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/examplestructfield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplestructfield/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

type retryer struct {
	connection *examplestructfield.Client
}

type withTokenFunc func(token *string) (any, error)

func (t *retryer) RetryWithToken(ctx context.Context, f withTokenFunc) (any, error) {
	const (
		key = "ExampleStructFieldRetryer"
	)
	conns.GlobalMutexKV.Lock(key)
	defer conns.GlobalMutexKV.Unlock(key)

	const (
		timeout = 15 * time.Minute
	)
	return tfresource.RetryWhenIsA[any, *awstypes.StaleDataException](ctx, timeout, func(ctx context.Context) (any, error) {
		input := &examplestructfield.GetChangeTokenInput{}
		output, err := t.connection.GetChangeToken(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("acquiring change token: %w", err)
		}

		return f(output.ChangeToken)
	})
}

func newRetryer(conn *examplestructfield.Client) *retryer {
	return &retryer{connection: conn}
}
