package keyspaces

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/keyspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	keyspaceExistsTimeout     = 1 * time.Minute
	keyspaceDisappearsTimeout = 1 * time.Minute
)

func waitKeyspaceExists(ctx context.Context, conn *keyspaces.Keyspaces, name string) error {
	err := resource.RetryContext(ctx, keyspaceExistsTimeout, func() *resource.RetryError {
		_, err := FindKeyspaceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = FindKeyspaceByName(ctx, conn, name)
	}

	return err
}

func waitKeyspaceDisappears(ctx context.Context, conn *keyspaces.Keyspaces, name string) error {
	stillExistsErr := fmt.Errorf("keyspace still exists")

	err := resource.RetryContext(ctx, keyspaceDisappearsTimeout, func() *resource.RetryError {
		_, err := FindKeyspaceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return resource.RetryableError(stillExistsErr)
	})

	if tfresource.TimedOut(err) {
		return stillExistsErr
	}

	return err
}
