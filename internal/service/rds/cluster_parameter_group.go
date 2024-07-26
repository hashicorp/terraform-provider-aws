// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_rds_cluster_parameter_group", name="Cluster Parameter Group")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func resourceClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterParameterGroupCreate,
		ReadWithoutTimeout:   resourceClusterParameterGroupRead,
		UpdateWithoutTimeout: resourceClusterParameterGroupUpdate,
		DeleteWithoutTimeout: resourceClusterParameterGroupDelete,

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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validParamGroupName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validParamGroupNamePrefix,
			},
			names.AttrParameter: {
				Type:     schema.TypeSet,
				Optional: true,
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
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &rds.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(name),
		DBParameterGroupFamily:      aws.String(d.Get(names.AttrFamily).(string)),
		Description:                 aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                        getTagsInV2(ctx),
	}

	output, err := conn.CreateDBClusterParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS Cluster Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	// Set for update.
	d.Set(names.AttrARN, output.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return append(diags, resourceClusterParameterGroupUpdate(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	dbClusterParameterGroup, err := findDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set(names.AttrDescription, dbClusterParameterGroup.Description)
	d.Set(names.AttrFamily, dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(dbClusterParameterGroup.DBClusterParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones.
	input := &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String(parameterSourceUser),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input, tfslices.PredicateTrue[*types.Parameter]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s) user parameters: %s", d.Id(), err)
	}

	// Add only system parameters that are set in the config.
	p := d.Get(names.AttrParameter)
	if p == nil {
		p = new(schema.Set)
	}
	s := p.(*schema.Set)
	configParameters := expandParameters(s.List())

	input = &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String(parameterSourceSystem),
	}

	systemParameters, err := findDBClusterParameters(ctx, conn, input, func(v *types.Parameter) bool {
		return slices.ContainsFunc(configParameters, func(p types.Parameter) bool {
			return aws.ToString(p.ParameterName) == aws.ToString(v.ParameterName)
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s) system parameters: %s", d.Id(), err)
	}

	parameters = append(parameters, systemParameters...)

	if err := d.Set(names.AttrParameter, flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	const (
		maxParamModifyChunk = 20
	)
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		os, ns := o.(*schema.Set), n.(*schema.Set)

		for _, chunk := range tfslices.Chunks(expandParameters(ns.Difference(os).List()), maxParamModifyChunk) {
			input := &rds.ModifyDBClusterParameterGroupInput{
				DBClusterParameterGroupName: aws.String(d.Id()),
				Parameters:                  chunk,
			}

			_, err := conn.ModifyDBClusterParameterGroup(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying RDS Cluster Parameter Group (%s): %s", d.Id(), err)
			}
		}

		toRemove := map[string]types.Parameter{}

		for _, p := range expandParameters(os.List()) {
			if p.ParameterName != nil {
				toRemove[aws.ToString(p.ParameterName)] = p
			}
		}

		for _, p := range expandParameters(ns.List()) {
			if p.ParameterName != nil {
				delete(toRemove, aws.ToString(p.ParameterName))
			}
		}

		// Reset parameters that have been removed.
		for _, chunk := range tfslices.Chunks(maps.Values(toRemove), maxParamModifyChunk) {
			input := &rds.ResetDBClusterParameterGroupInput{
				DBClusterParameterGroupName: aws.String(d.Id()),
				Parameters:                  chunk,
				ResetAllParameters:          aws.Bool(false),
			}

			const (
				timeout = 3 * time.Minute
			)
			_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidDBParameterGroupStateFault](ctx, timeout, func() (interface{}, error) {
				return conn.ResetDBClusterParameterGroup(ctx, input)
			}, "has pending changes")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting RDS Cluster Parameter Group (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient(ctx)

	log.Printf("[DEBUG] Deleting RDS Cluster Parameter Group: %s", d.Id())
	const (
		timeout = 3 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.InvalidDBParameterGroupStateFault](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteDBClusterParameterGroup(ctx, &rds.DeleteDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func findDBClusterParameterGroupByName(ctx context.Context, conn *rds.Client, name string) (*types.DBClusterParameterGroup, error) {
	input := &rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}
	output, err := findDBClusterParameterGroup(ctx, conn, input, tfslices.PredicateTrue[*types.DBClusterParameterGroup]())

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBClusterParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBClusterParameterGroup(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterParameterGroupsInput, filter tfslices.Predicate[*types.DBClusterParameterGroup]) (*types.DBClusterParameterGroup, error) {
	output, err := findDBClusterParameterGroups(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterParameterGroups(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterParameterGroupsInput, filter tfslices.Predicate[*types.DBClusterParameterGroup]) ([]types.DBClusterParameterGroup, error) {
	var output []types.DBClusterParameterGroup

	pages := rds.NewDescribeDBClusterParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterParameterGroups {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findDBClusterParameters(ctx context.Context, conn *rds.Client, input *rds.DescribeDBClusterParametersInput, filter tfslices.Predicate[*types.Parameter]) ([]types.Parameter, error) {
	var output []types.Parameter

	pages := rds.NewDescribeDBClusterParametersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Parameters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
