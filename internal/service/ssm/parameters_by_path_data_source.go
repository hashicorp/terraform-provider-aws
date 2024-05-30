// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ssm_parameters_by_path", name="Parameters By Path")
func dataSourceParametersByPath() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceParametersReadByPath,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrNames: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Required: true,
			},
			"recursive": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"types": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrValues: {
				Type:      schema.TypeList,
				Computed:  true,
				Sensitive: true,
				Elem:      &schema.Schema{Type: schema.TypeString},
			},
			"with_decryption": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func dataSourceParametersReadByPath(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SSMClient(ctx)

	path := d.Get(names.AttrPath).(string)
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		Recursive:      aws.Bool(d.Get("recursive").(bool)),
		WithDecryption: aws.Bool(d.Get("with_decryption").(bool)),
	}
	var output []awstypes.Parameter

	pages := ssm.NewGetParametersByPathPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading SSM Parameters by path (%s): %s", path, err)
		}

		output = append(output, page.Parameters...)
	}

	d.SetId(path)
	d.Set(names.AttrARNs, tfslices.ApplyToAll(output, func(v awstypes.Parameter) string {
		return aws.ToString(v.ARN)
	}))
	d.Set(names.AttrNames, tfslices.ApplyToAll(output, func(v awstypes.Parameter) string {
		return aws.ToString(v.Name)
	}))
	d.Set("types", tfslices.ApplyToAll(output, func(v awstypes.Parameter) awstypes.ParameterType {
		return v.Type
	}))
	d.Set(names.AttrValues, tfslices.ApplyToAll(output, func(v awstypes.Parameter) string {
		return aws.ToString(v.Value)
	}))

	return diags
}
