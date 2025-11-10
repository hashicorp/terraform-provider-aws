// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCluster(ctx context.Context, conn *r53rcc.Client, clusterArn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findClusterByARN(ctx, conn, clusterArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusRoutingControl(ctx context.Context, conn *r53rcc.Client, routingControlArn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findRoutingControlByARN(ctx, conn, routingControlArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusControlPanel(ctx context.Context, conn *r53rcc.Client, controlPanelArn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findControlPanelByARN(ctx, conn, controlPanelArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSafetyRule(ctx context.Context, conn *r53rcc.Client, safetyRuleArn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSafetyRuleByARN(ctx, conn, safetyRuleArn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		if output.AssertionRule != nil {
			return output, string(output.AssertionRule.Status), nil
		}

		if output.GatingRule != nil {
			return output, string(output.GatingRule.Status), nil
		}

		return output, "", nil
	}
}
