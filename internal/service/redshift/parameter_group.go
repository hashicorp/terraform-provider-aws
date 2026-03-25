// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package redshift

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_parameter_group", name="Parameter Group")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/redshift/types;awstypes;awstypes.ClusterParameterGroup")
func resourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			names.AttrFamily: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexache.MustCompile(`(?i)^[a-z]`), "first character must be a letter"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				Set: resourceParameterHash,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := redshift.CreateClusterParameterGroupInput{
		Description:          aws.String(d.Get(names.AttrDescription).(string)),
		ParameterGroupFamily: aws.String(d.Get(names.AttrFamily).(string)),
		ParameterGroupName:   aws.String(name),
		Tags:                 getTagsIn(ctx),
	}

	_, err := conn.CreateClusterParameterGroup(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v := d.Get(names.AttrParameter).(*schema.Set); v.Len() > 0 {
		input := redshift.ModifyClusterParameterGroupInput{
			ParameterGroupName: aws.String(d.Id()),
			Parameters:         expandParameters(v.List()),
		}

		_, err := conn.ModifyClusterParameterGroup(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.RedshiftClient(ctx)

	parameterGroup, err := findParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, parameterGroupARN(ctx, c, d.Id()))
	d.Set(names.AttrDescription, parameterGroup.Description)
	d.Set(names.AttrFamily, parameterGroup.ParameterGroupFamily)
	d.Set(names.AttrName, parameterGroup.ParameterGroupName)

	setTagsOut(ctx, parameterGroup.Tags)

	input := redshift.DescribeClusterParametersInput{
		ParameterGroupName: aws.String(d.Id()),
		Source:             aws.String("user"),
	}
	parameters, err := findClusterParameters(ctx, conn, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	d.Set(names.AttrParameter, flattenParameters(parameters))

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if parameters := expandParameters(ns.Difference(os).List()); len(parameters) > 0 {
			input := redshift.ModifyClusterParameterGroupInput{
				ParameterGroupName: aws.String(d.Id()),
				Parameters:         parameters,
			}

			_, err := conn.ModifyClusterParameterGroup(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	log.Printf("[DEBUG] Deleting Redshift Parameter Group: %s", d.Id())
	input := redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	}
	_, err := conn.DeleteClusterParameterGroup(ctx, &input)

	if errs.IsA[*awstypes.ClusterParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceParameterHash(v any) int {
	var buf bytes.Buffer
	m := v.(map[string]any)
	fmt.Fprintf(&buf, "%s-", m[names.AttrName].(string))
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	fmt.Fprintf(&buf, "%s-", strings.ToLower(m[names.AttrValue].(string)))

	return create.StringHashcode(buf.String())
}

func expandParameters(tfList []any) []awstypes.Parameter {
	var apiObjects []awstypes.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]any)

		if tfMap[names.AttrName].(string) == "" {
			continue
		}

		apiObject := awstypes.Parameter{
			ParameterName:  aws.String(tfMap[names.AttrName].(string)),
			ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenParameters(apiObjects []awstypes.Parameter) []any {
	tfList := make([]any, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			names.AttrName:  aws.ToString(apiObject.ParameterName),
			names.AttrValue: aws.ToString(apiObject.ParameterValue),
		})
	}

	return tfList
}

func parameterGroupARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.RegionalARN(ctx, names.Redshift, "parametergroup:"+id)
}
