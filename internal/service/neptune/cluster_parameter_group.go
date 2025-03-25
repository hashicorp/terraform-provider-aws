// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptune/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_neptune_cluster_parameter_group", name="Cluster Parameter Group")
// @Tags(identifierAttribute="arn")
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
							Default:          awstypes.ApplyMethodPendingReboot,
							ValidateDiagFunc: enum.Validate[awstypes.ApplyMethod](),
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
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceClusterParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &neptune.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(name),
		DBParameterGroupFamily:      aws.String(d.Get(names.AttrFamily).(string)),
		Description:                 aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                        getTagsIn(ctx),
	}

	_, err := conn.CreateDBClusterParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk(names.AttrParameter); ok && v.(*schema.Set).Len() > 0 {
		if err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), expandParameters(v.(*schema.Set).List())); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	dbClusterParameterGroup, err := findDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	arn := aws.ToString(dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, dbClusterParameterGroup.Description)
	d.Set(names.AttrFamily, dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(dbClusterParameterGroup.DBClusterParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones.
	input := &neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input, tfslices.PredicateTrue[awstypes.Parameter]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s) user parameters: %s", d.Id(), err)
	}

	// Add only system parameters that are set in the config.
	p := d.Get(names.AttrParameter)
	if p == nil {
		p = new(schema.Set)
	}
	configParameters := expandParameters(p.(*schema.Set).List())

	input = &neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("engine-default"),
	}

	systemParameters, err := findDBClusterParameters(ctx, conn, input, func(v awstypes.Parameter) bool {
		return slices.ContainsFunc(configParameters, func(p awstypes.Parameter) bool {
			return aws.ToString(v.ParameterName) == aws.ToString(p.ParameterName)
		})
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s) system parameters: %s", d.Id(), err)
	}

	parameters = append(parameters, systemParameters...)

	if err := d.Set(names.AttrParameter, flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		os, ns := o.(*schema.Set), n.(*schema.Set)

		if parameters := expandParameters(ns.Difference(os).List()); len(parameters) > 0 {
			err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), parameters)

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	log.Printf("[DEBUG] Deleting Neptune Cluster Parameter Group: %s", d.Id())
	_, err := conn.DeleteDBClusterParameterGroup(ctx, &neptune.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DBParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func modifyClusterParameterGroupParameters(ctx context.Context, conn *neptune.Client, name string, parameters []awstypes.Parameter) error {
	const (
		clusterParameterGroupMaxParamsBulkEdit = 20
	)
	// We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for chunk := range slices.Chunk(parameters, clusterParameterGroupMaxParamsBulkEdit) {
		input := &neptune.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  chunk,
		}

		_, err := conn.ModifyDBClusterParameterGroup(ctx, input)

		if err != nil {
			return fmt.Errorf("modifying Neptune Cluster Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func findDBClusterParameterGroupByName(ctx context.Context, conn *neptune.Client, name string) (*awstypes.DBClusterParameterGroup, error) {
	input := &neptune.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}
	output, err := findDBClusterParameterGroup(ctx, conn, input)

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

func findDBClusterParameterGroup(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBClusterParameterGroupsInput) (*awstypes.DBClusterParameterGroup, error) {
	output, err := findDBClusterParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterParameterGroups(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBClusterParameterGroupsInput) ([]awstypes.DBClusterParameterGroup, error) {
	var output []awstypes.DBClusterParameterGroup

	pages := neptune.NewDescribeDBClusterParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DBClusterParameterGroups...)
	}

	return output, nil
}

func findDBClusterParameters(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBClusterParametersInput, filter tfslices.Predicate[awstypes.Parameter]) ([]awstypes.Parameter, error) {
	var output []awstypes.Parameter

	pages := neptune.NewDescribeDBClusterParametersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.DBParameterGroupNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.Parameters {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
