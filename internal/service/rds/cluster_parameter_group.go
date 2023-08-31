// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
	"golang.org/x/exp/maps"
)

// @SDKResource("aws_rds_cluster_parameter_group", name="Cluster Parameter Group")
// @Tags(identifierAttribute="arn")
func ResourceClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterParameterGroupCreate,
		ReadWithoutTimeout:   resourceClusterParameterGroupRead,
		UpdateWithoutTimeout: resourceClusterParameterGroupUpdate,
		DeleteWithoutTimeout: resourceClusterParameterGroupDelete,

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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validParamGroupNamePrefix,
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "immediate",
						},
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

func resourceClusterParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	groupName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        getTagsIn(ctx),
	}

	output, err := conn.CreateDBClusterParameterGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DB Cluster Parameter Group (%s): %s", groupName, err)
	}

	d.SetId(groupName)

	// Set for update
	d.Set("arn", output.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return append(diags, resourceClusterParameterGroupUpdate(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	dbClusterParameterGroup, err := FindDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", dbClusterParameterGroup.Description)
	d.Set("family", dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones
	input := &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}
	var parameters []*rds.Parameter

	err = conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *rds.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			if v != nil {
				parameters = append(parameters, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	// add only system parameters that are set in the config
	p := d.Get("parameter")
	if p == nil {
		p = new(schema.Set)
	}
	s := p.(*schema.Set)
	configParameters := expandParameters(s.List())

	input = &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("system"),
	}

	err = conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *rds.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			for _, p := range configParameters {
				if aws.StringValue(v.ParameterName) == aws.StringValue(p.ParameterName) {
					parameters = append(parameters, v)
				}
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		maxParamModifyChunk = 20
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

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

		// Expand the "parameter" set to aws-sdk-go compat []rds.Parameter.
		for _, chunk := range slices.Chunks(expandParameters(ns.Difference(os).List()), maxParamModifyChunk) {
			input := &rds.ModifyDBClusterParameterGroupInput{
				DBClusterParameterGroupName: aws.String(d.Id()),
				Parameters:                  chunk,
			}

			_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying DB Cluster Parameter Group (%s): %s", d.Id(), err)
			}
		}

		toRemove := map[string]*rds.Parameter{}

		for _, p := range expandParameters(os.List()) {
			if p.ParameterName != nil {
				toRemove[*p.ParameterName] = p
			}
		}

		for _, p := range expandParameters(ns.List()) {
			if p.ParameterName != nil {
				delete(toRemove, *p.ParameterName)
			}
		}

		// Reset parameters that have been removed.
		for _, chunk := range slices.Chunks(maps.Values(toRemove), maxParamModifyChunk) {
			input := &rds.ResetDBClusterParameterGroupInput{
				DBClusterParameterGroupName: aws.String(d.Id()),
				Parameters:                  chunk,
				ResetAllParameters:          aws.Bool(false),
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 3*time.Minute, func() (interface{}, error) {
				return conn.ResetDBClusterParameterGroupWithContext(ctx, input)
			}, rds.ErrCodeInvalidDBParameterGroupStateFault, "has pending changes")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting DB Cluster Parameter Group (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	input := &rds_sdkv2.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting RDS DB Cluster Parameter Group: %s", d.Id())
	err := retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteDBClusterParameterGroup(ctx, input)
		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil
		} else if errs.IsA[*types.InvalidDBParameterGroupStateFault](err) {
			return retry.RetryableError(err)
		}
		if err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDBClusterParameterGroup(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDBClusterParameterGroupByName(ctx context.Context, conn *rds.RDS, name string) (*rds.DBClusterParameterGroup, error) {
	input := &rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBClusterParameterGroups) == 0 || output.DBClusterParameterGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.DBClusterParameterGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	dbClusterParameterGroup := output.DBClusterParameterGroups[0]

	// Eventual consistency check.
	if aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return dbClusterParameterGroup, nil
}
