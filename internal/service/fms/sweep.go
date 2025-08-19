// Copyright (c) HashiCorp, Inc.fms/sweep
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_fms_admin_account", &resource.Sweeper{
		Name: "aws_fms_admin_account",
		F:    sweepAdminAccount,
	})
}

// TODO: This sweeper has custom skip logic, so can't use a `sweep.SweeperFn`
func sweepAdminAccount(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.FMSClient(ctx)

	var sweepResources []sweep.Sweepable

	output, err := conn.GetAdminAccount(ctx, &fms.GetAdminAccountInput{})
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		tflog.Info(ctx, "No resources to sweep")
		return nil
	}
	if awsv2.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "AccessDeniedException") {
		tflog.Warn(ctx, "Skipping sweeper", map[string]any{
			"error": err.Error(),
		})
		return nil
	}
	if err != nil {
		return err
	}

	if output == nil {
		tflog.Warn(ctx, "Skipping sweeper", map[string]any{
			"skip_reason": "Empty result",
		})
		return nil
	}

	r := resourceAdminAccount()
	d := r.Data(nil)
	d.SetId(aws.ToString(output.AdminAccount))
	d.Set(names.AttrAccountID, output.AdminAccount)

	sweepResources = append(sweepResources, newAdminAccountSweeper(r, d, client))

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping FMS Admin Account (%s): %w", region, err)
	}

	return nil
}

type adminAccountSweeper struct {
	d         *schema.ResourceData
	sweepable sweep.Sweepable
}

func newAdminAccountSweeper(resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) *adminAccountSweeper {
	return &adminAccountSweeper{
		d:         d,
		sweepable: sdk.NewSweepResource(resource, d, client),
	}
}

func (aas adminAccountSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	err := aas.sweepable.Delete(ctx, optFns...)
	if err != nil && strings.Contains(err.Error(), "AccessDeniedException") {
		tflog.Warn(ctx, "Skipping resource", map[string]any{
			"attr.account_id": aas.d.Get(names.AttrAccountID),
			"error":           err.Error(),
		})
		return nil
	}
	if err != nil && errs.Must(regexp.MatchString(`InvalidOperationException: This operation is not supported in the '[-a-z0-9]+' region`, err.Error())) { // nosemgrep: ci.avoid-errs-Must
		tflog.Warn(ctx, "Skipping resource", map[string]any{
			"attr.account_id": aas.d.Get(names.AttrAccountID),
			"error":           err.Error(),
		})
		return nil
	}
	return err
}
