// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudformation_stack", name="Stack")
// @Tags
func dataSourceStack() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceStackRead,

		Schema: map[string]*schema.Schema{
			"capabilities": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_rollback": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrIAMRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"notification_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"outputs": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"template_body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"timeout_in_minutes": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceStackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFormationClient(ctx)

	name := d.Get(names.AttrName).(string)

	stack, err := findStackByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("CloudFormation Stack", err))
	}

	d.SetId(aws.ToString(stack.StackId))
	if len(stack.Capabilities) > 0 {
		d.Set("capabilities", stack.Capabilities)
	}
	d.Set(names.AttrDescription, stack.Description)
	d.Set("disable_rollback", stack.DisableRollback)
	d.Set(names.AttrIAMRoleARN, stack.RoleARN)
	if len(stack.NotificationARNs) > 0 {
		d.Set("notification_arns", stack.NotificationARNs)
	}
	d.Set("outputs", flattenOutputs(stack.Outputs))
	d.Set(names.AttrParameters, flattenAllParameters(stack.Parameters))
	d.Set("timeout_in_minutes", stack.TimeoutInMinutes)

	setTagsOut(ctx, stack.Tags)

	input := &cloudformation.GetTemplateInput{
		StackName: aws.String(name),
	}

	output, err := conn.GetTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFormation Stack (%s) template: %s", name, err)
	}

	template, err := verify.NormalizeJSONOrYAMLString(*output.TemplateBody)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("template_body", template)

	return diags
}
