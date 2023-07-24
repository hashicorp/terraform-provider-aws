// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_parameter_group", name="Parameter Group")
// @Tags(identifierAttribute="arn")
func ResourceParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceParameterGroupCreate,
		ReadWithoutTimeout:   resourceParameterGroupRead,
		UpdateWithoutTimeout: resourceParameterGroupUpdate,
		DeleteWithoutTimeout: resourceParameterGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[0-9a-z-]+$`), "must contain only lowercase alphanumeric characters and hyphens"),
					validation.StringMatch(regexp.MustCompile(`(?i)^[a-z]`), "first character must be a letter"),
					validation.StringDoesNotMatch(regexp.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexp.MustCompile(`-$`), "cannot end with a hyphen"),
				),
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	name := d.Get("name").(string)
	input := &redshift.CreateClusterParameterGroupInput{
		Description:          aws.String(d.Get("description").(string)),
		ParameterGroupFamily: aws.String(d.Get("family").(string)),
		ParameterGroupName:   aws.String(name),
		Tags:                 getTagsIn(ctx),
	}

	_, err := conn.CreateClusterParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v := d.Get("parameter").(*schema.Set); v.Len() > 0 {
		input := &redshift.ModifyClusterParameterGroupInput{
			ParameterGroupName: aws.String(d.Id()),
			Parameters:         expandParameters(v.List()),
		}

		_, err := conn.ModifyClusterParameterGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	parameterGroup, err := FindParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Redshift Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "redshift",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("parametergroup:%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("description", parameterGroup.Description)
	d.Set("family", parameterGroup.ParameterGroupFamily)
	d.Set("name", parameterGroup.ParameterGroupName)

	setTagsOut(ctx, parameterGroup.Tags)

	input := &redshift.DescribeClusterParametersInput{
		ParameterGroupName: aws.String(d.Id()),
		Source:             aws.String("user"),
	}

	output, err := conn.DescribeClusterParametersWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	d.Set("parameter", flattenParameters(output.Parameters))

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		parameters := expandParameters(ns.Difference(os).List())
		if len(parameters) > 0 {
			input := &redshift.ModifyClusterParameterGroupInput{
				ParameterGroupName: aws.String(d.Id()),
				Parameters:         parameters,
			}

			_, err := conn.ModifyClusterParameterGroupWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Redshift Parameter Group (%s) parameters: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	log.Printf("[DEBUG] Deleting Redshift Parameter Group: %s", d.Id())
	_, err := conn.DeleteClusterParameterGroupWithContext(ctx, &redshift.DeleteClusterParameterGroupInput{
		ParameterGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindParameterGroupByName(ctx context.Context, conn *redshift.Redshift, name string) (*redshift.ClusterParameterGroup, error) {
	input := &redshift.DescribeClusterParameterGroupsInput{
		ParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeClusterParameterGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshift.ErrCodeClusterParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ParameterGroups) == 0 || output.ParameterGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	parameterGroup := output.ParameterGroups[0]

	// Eventual consistency check.
	if aws.StringValue(parameterGroup.ParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return parameterGroup, nil
}

func resourceParameterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["name"].(string)))
	// Store the value as a lower case string, to match how we store them in FlattenParameters
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["value"].(string))))

	return create.StringHashcode(buf.String())
}

func expandParameters(configured []interface{}) []*redshift.Parameter {
	var parameters []*redshift.Parameter

	// Loop over our configured parameters and create
	// an array of aws-sdk-go compatible objects
	for _, pRaw := range configured {
		data := pRaw.(map[string]interface{})

		if data["name"].(string) == "" {
			continue
		}

		p := &redshift.Parameter{
			ParameterName:  aws.String(data["name"].(string)),
			ParameterValue: aws.String(data["value"].(string)),
		}

		parameters = append(parameters, p)
	}

	return parameters
}

func flattenParameters(list []*redshift.Parameter) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		result = append(result, map[string]interface{}{
			"name":  aws.StringValue(i.ParameterName),
			"value": aws.StringValue(i.ParameterValue),
		})
	}
	return result
}
