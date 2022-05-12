package appflow

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindFlowByArn(ctx context.Context, conn *appflow.Appflow, arn string) (*appflow.FlowDefinition, error) {
	in := &appflow.ListFlowsInput{}
	var result *appflow.FlowDefinition

	err := conn.ListFlowsPagesWithContext(ctx, in, func(page *appflow.ListFlowsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, flow := range page.Flows {
			if flow == nil {
				continue
			}

			if aws.StringValue(flow.FlowArn) == arn {
				result = flow
				return false
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No flow with arn %q", arn),
			LastRequest: in,
		}
	}

	return result, nil
}

func FindConnectorProfileByArn(ctx context.Context, conn *appflow.Appflow, arn string) (*appflow.ConnectorProfile, error) {
	params := &appflow.DescribeConnectorProfilesInput{}
	var result *appflow.ConnectorProfile

	err := conn.DescribeConnectorProfilesPagesWithContext(ctx, params, func(page *appflow.DescribeConnectorProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, connectorProfile := range page.ConnectorProfileDetails {
			if connectorProfile == nil {
				continue
			}

			if aws.StringValue(connectorProfile.ConnectorProfileArn) == arn {
				result = connectorProfile
				return false
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: params,
		}
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No connector profile with arn %q", arn),
			LastRequest: params,
		}
	}

	return result, nil
}
