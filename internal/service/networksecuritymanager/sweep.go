// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_networksecuritymanager_admin_account", sweepAdminAccounts,
		"aws_networksecuritymanager_policy",
		"aws_networksecuritymanager_rule",
		"aws_networksecuritymanager_scope",
		"aws_networksecuritymanager_template",
	)
	awsv2.Register("aws_networksecuritymanager_deployment", sweepDeployments)
	// A policy or scope cannot be deleted while a deployment uses it.
	awsv2.Register("aws_networksecuritymanager_policy", sweepPolicies, "aws_networksecuritymanager_deployment")
	awsv2.Register("aws_networksecuritymanager_rule", sweepRules, "aws_networksecuritymanager_policy", "aws_networksecuritymanager_template")
	awsv2.Register("aws_networksecuritymanager_scope", sweepScopes, "aws_networksecuritymanager_deployment")
	awsv2.Register("aws_networksecuritymanager_template", sweepTemplates)
}

func sweepDeployments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListDeploymentsInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListDeploymentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Deployments {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDeploymentResource, client,
				framework.NewAttribute(names.AttrARN, deploymentBaseARN(aws.ToString(v.DeploymentArn)))),
			)
		}
	}

	return sweepResources, nil
}

func sweepTemplates(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListTemplatesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListTemplatesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Templates {
			sweepResources = append(sweepResources, framework.NewSweepResource(newTemplateResource, client,
				framework.NewAttribute(names.AttrARN, templateBaseARN(aws.ToString(v.TemplateArn)))),
			)
		}
	}

	return sweepResources, nil
}

func sweepPolicies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListPoliciesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListPoliciesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Policies {
			sweepResources = append(sweepResources, framework.NewSweepResource(newPolicyResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.PolicyArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListRulesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListRulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Rules {
			sweepResources = append(sweepResources, framework.NewSweepResource(newRuleResource, client,
				framework.NewAttribute(names.AttrARN, ruleBaseARN(aws.ToString(v.RuleArn)))),
			)
		}
	}

	return sweepResources, nil
}

func sweepScopes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListScopesInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := networksecuritymanager.NewListScopesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Scopes {
			sweepResources = append(sweepResources, framework.NewSweepResource(newScopeResource, client,
				framework.NewAttribute(names.AttrARN, scopeBaseARN(aws.ToString(v.ScopeArn)))),
			)
		}
	}

	return sweepResources, nil
}

func sweepAdminAccounts(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var input networksecuritymanager.ListAdminAccountsInput
	conn := client.NetworkSecurityManagerClient(ctx)
	var sweepResources []sweep.Sweepable

	for v, err := range listAdminAccounts(ctx, conn, &input) {
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		sweepResources = append(sweepResources, framework.NewSweepResource(newAdminAccountResource, client,
			framework.NewAttribute(names.AttrAccountID, aws.ToString(v.AccountId))),
		)
	}

	return sweepResources, nil
}
