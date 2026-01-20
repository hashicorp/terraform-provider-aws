// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_db_parameter_group", name="DB Parameter Group")
func dataSourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceParameterGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFamily: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_method": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ApplyMethodImmediate,
							ValidateDiagFunc: enum.ValidateIgnoreCase[types.ApplyMethod](),
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: parameterHash,
			},
		},
	}
}

func dataSourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	output, err := findDBParameterGroupByName(ctx, conn, d.Get(names.AttrName).(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("RDS DB Parameter Group", err))
	}

	d.SetId(aws.ToString(output.DBParameterGroupName))
	d.Set(names.AttrARN, output.DBParameterGroupArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set(names.AttrFamily, output.DBParameterGroupFamily)
	d.Set(names.AttrName, output.DBParameterGroupName)

	input := rds.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	configParams := d.Get(names.AttrParameter).(*schema.Set)
	if configParams.Len() < 1 {
		input.Source = aws.String(parameterSourceUser)
	}

	parameters, err := findDBParameters(ctx, conn, &input, tfslices.PredicateTrue[*types.Parameter]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	var userParams []types.Parameter
	if configParams.Len() < 1 {
		userParams = parameters
	} else {
		for _, parameter := range parameters {
			if parameter.Source == nil || parameter.ParameterName == nil {
				continue
			}

			if aws.ToString(parameter.Source) == parameterSourceUser {
				userParams = append(userParams, parameter)
				continue
			}

			var paramFound bool
			for _, cp := range expandParameters(configParams.List()) {
				if cp.ParameterName == nil {
					continue
				}

				if aws.ToString(cp.ParameterName) == aws.ToString(parameter.ParameterName) {
					userParams = append(userParams, parameter)
					paramFound = true
					break
				}
			}
			if !paramFound {
				log.Printf("[DEBUG] Not getting %s, as its source is %q and it isn't in the config", aws.ToString(parameter.ParameterName), aws.ToString(parameter.Source))
			}
		}
	}

	if err := d.Set(names.AttrParameter, flattenParameters(userParams)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}
