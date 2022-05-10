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

func FindConnectorProfileByName(conn *appflow.Appflow, name string) (*appflow.ConnectorProfile, error) {
	params := &appflow.DescribeConnectorProfilesInput{
		ConnectorProfileNames: []*string{aws.String(name)},
	}

	for {
		output, err := conn.DescribeConnectorProfiles(params)

		if err != nil {
			return nil, err
		}

		for _, connectorProfile := range output.ConnectorProfileDetails {
			if aws.StringValue(connectorProfile.ConnectorProfileName) == name {
				return connectorProfile, nil
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		params.NextToken = output.NextToken
	}

	return nil, fmt.Errorf("No connector profile found with name: %s", name)
}
