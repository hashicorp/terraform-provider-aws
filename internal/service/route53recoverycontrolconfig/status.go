package route53recoverycontrolconfig

import (
	"github.com/aws/aws-sdk-go/aws"
	r53rcc "github.com/aws/aws-sdk-go/service/route53recoverycontrolconfig"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusCluster(conn *r53rcc.Route53RecoveryControlConfig, clusterArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeClusterInput{
			ClusterArn: aws.String(clusterArn),
		}

		output, err := conn.DescribeCluster(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.Cluster.Status), nil
	}
}

func statusRoutingControl(conn *r53rcc.Route53RecoveryControlConfig, routingControlArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeRoutingControlInput{
			RoutingControlArn: aws.String(routingControlArn),
		}

		output, err := conn.DescribeRoutingControl(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.RoutingControl.Status), nil
	}
}

func statusControlPanel(conn *r53rcc.Route53RecoveryControlConfig, controlPanelArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeControlPanelInput{
			ControlPanelArn: aws.String(controlPanelArn),
		}

		output, err := conn.DescribeControlPanel(input)

		if err != nil {
			return output, "", err
		}

		return output, aws.StringValue(output.ControlPanel.Status), nil
	}
}

func statusSafetyRule(conn *r53rcc.Route53RecoveryControlConfig, safetyRuleArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &r53rcc.DescribeSafetyRuleInput{
			SafetyRuleArn: aws.String(safetyRuleArn),
		}

		output, err := conn.DescribeSafetyRule(input)

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
