// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAlertManagerDefinitionByID(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	input := &prometheusservice.DescribeAlertManagerDefinitionInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeAlertManagerDefinitionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AlertManagerDefinition == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AlertManagerDefinition, nil
}

func nameAndWorkspaceIDFromRuleGroupNamespaceARN(arn string) (string, string, error) {
	parts := strings.Split(arn, "/")
	if len(parts) != 3 {
		return "", "", fmt.Errorf("reading Prometheus Rule Group Namespace expected the arn to be like: arn:PARTITION:aps:REGION:ACCOUNT:rulegroupsnamespace/IDstring/namespace_name but got: %s", arn)
	}
	return parts[2], parts[1], nil
}

func FindRuleGroupNamespaceByARN(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) (*prometheusservice.RuleGroupsNamespaceDescription, error) {
	name, workspaceId, err := nameAndWorkspaceIDFromRuleGroupNamespaceARN(arn)
	if err != nil {
		return nil, err
	}

	input := &prometheusservice.DescribeRuleGroupsNamespaceInput{
		Name:        aws.String(name),
		WorkspaceId: aws.String(workspaceId),
	}

	output, err := conn.DescribeRuleGroupsNamespaceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RuleGroupsNamespace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RuleGroupsNamespace, nil
}

func FindWorkspaceByID(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.WorkspaceDescription, error) {
	input := &prometheusservice.DescribeWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspaceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workspace == nil || output.Workspace.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workspace, nil
}

func FindLoggingConfigurationByWorkspaceID(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.LoggingConfigurationMetadata, error) {
	input := &prometheusservice.DescribeLoggingConfigurationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeLoggingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfiguration == nil || output.LoggingConfiguration.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoggingConfiguration, nil
}

func FindWorkspaces(ctx context.Context, conn *prometheusservice.PrometheusService, alias string) ([]*prometheusservice.WorkspaceSummary, error) { // nosemgrep:ci.caps0-in-func-name
	input := &prometheusservice.ListWorkspacesInput{}
	if alias != "" {
		input.Alias = aws.String(alias)
	}
	var output []*prometheusservice.WorkspaceSummary

	err := conn.ListWorkspacesPagesWithContext(ctx, input, func(page *prometheusservice.ListWorkspacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Workspaces {
			if v == nil {
				continue
			}
			output = append(output, v)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
