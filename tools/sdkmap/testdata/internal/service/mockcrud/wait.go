// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockcrud

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/mockcrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/mockcrud/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func waitFooCreated(ctx context.Context, conn *mockcrud.Client, id string, timeout time.Duration) (*awstypes.Foo, error) {
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

func waitFooDeleted(ctx context.Context, conn *mockcrud.Client, id string, timeout time.Duration) (*awstypes.Foo, error) {
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
