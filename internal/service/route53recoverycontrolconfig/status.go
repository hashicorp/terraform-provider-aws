// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func statusCluster(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeClusterInput{
			ClusterArn: aws.String(clusterArn),
		}

		output, err := conn.DescribeClusterWithContext(ctx, input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.Cluster.Status), nil
	}
}

func statusRoutingControl(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(routingControlArn),
		}

		output, err := conn.DescribeRoutingControlWithContext(ctx, input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.RoutingControl.Status), nil
	}
}

func statusControlPanel(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeControlPanelInput{
			ControlPanelArn: aws.String(controlPanelArn),
		}

		output, err := conn.DescribeControlPanelWithContext(ctx, input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.ControlPanel.Status), nil
	}
}

func statusSafetyRule(ctx context.Context, conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(safetyRuleArn),
		}

		output, err := conn.DescribeSafetyRuleWithContext(ctx, input)

		if err != nil {
			return output, "", err
		}

		if output.AssertionRule != nil {
			return output, aws.StringValue(output.AssertionRule.Status), nil
		}

		if output.GatingRule != nil {
			return output, aws.StringValue(output.GatingRule.Status), nil
		}

		return output, "", nil
	}
}
