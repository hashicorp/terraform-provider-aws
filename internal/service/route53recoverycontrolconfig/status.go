// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusCluster(conn *r53rcc.Client, clusterArn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findClusterByARN(ctx, conn, clusterArn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusRoutingControl(conn *r53rcc.Client, routingControlArn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findRoutingControlByARN(ctx, conn, routingControlArn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusControlPanel(conn *r53rcc.Client, controlPanelArn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findControlPanelByARN(ctx, conn, controlPanelArn)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return output, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSafetyRule(conn *r53rcc.Client, safetyRuleArn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSafetyRuleByARN(ctx, conn, safetyRuleArn)

		if retry.NotFound(err) {
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
