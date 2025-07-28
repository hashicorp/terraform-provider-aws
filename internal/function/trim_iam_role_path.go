// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

const (
	// IAM role ARN reference:
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/list_awsidentityandaccessmanagementiam.html#awsidentityandaccessmanagementiam-resources-for-iam-policies

	// resourceSectionPrefix is the expected prefix in the resource section of
	// an IAM role ARN
	resourceSectionPrefix = "role/"

	// serviceSection is the expected service section of an IAM role ARN
	serviceSection = "iam"
)

var _ function.Function = trimIAMRolePathFunction{}

func NewTrimIAMRolePathFunction() function.Function {
	return &trimIAMRolePathFunction{}
}

type trimIAMRolePathFunction struct{}

func (f trimIAMRolePathFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "trim_iam_role_path"
}

func (f trimIAMRolePathFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "trim_iam_role_path Function",
		MarkdownDescription: "Trims the path prefix from an IAM role Amazon Resource Name (ARN). This " +
			"function can be used when services require role ARNs to be passed without a path.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "arn",
				MarkdownDescription: "IAM role Amazon Resource Name (ARN)",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f trimIAMRolePathFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var arg string

	resp.Error = function.ConcatFuncErrors(req.Arguments.Get(ctx, &arg))
	if resp.Error != nil {
		return
	}

	result, err := trimPath(arg)
	if err != nil {
		resp.Error = function.ConcatFuncErrors(resp.Error, function.NewFuncError(err.Error()))
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Result.Set(ctx, result))
}

// trimPath removes all path prefixes from the resource section of a role ARN
func trimPath(s string) (string, error) {
	rarn, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	if rarn.Service != serviceSection {
		return "", fmt.Errorf(`service must be "%s"`, serviceSection)
	}
	if rarn.Region != "" {
		return "", fmt.Errorf("region must be empty")
	}
	if !strings.HasPrefix(rarn.Resource, resourceSectionPrefix) {
		return "", fmt.Errorf(`resource must begin with "%s"`, resourceSectionPrefix)
	}

	sec := strings.Split(rarn.Resource, "/")
	rarn.Resource = fmt.Sprintf("%s%s", resourceSectionPrefix, sec[len(sec)-1])

	return rarn.String(), nil
}
