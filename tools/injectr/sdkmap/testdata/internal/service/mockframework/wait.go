// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframework

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/mockframework"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockframework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusFoo(conn *mockframework.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFooByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitFooCreated(ctx context.Context, conn *mockframework.Client, id string, timeout time.Duration) (*awstypes.Foo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FooStatusCreating),
		Target:  enum.Slice(awstypes.FooStatusAvailable),
		Refresh: statusFoo(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Foo); ok {
		return output, err
	}

	return nil, err
}

func waitFooDeleted(ctx context.Context, conn *mockframework.Client, id string, timeout time.Duration) (*awstypes.Foo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FooStatusAvailable, awstypes.FooStatusDeleting),
		Target:  []string{},
		Refresh: statusFoo(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Foo); ok {
		return output, err
	}

	return nil, err
}
