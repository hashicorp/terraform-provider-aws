// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	dbParameterGroupMaxParamsBulkEdit            = 20
	dbParameterGroupDeleteRetryTimeout           = 3 * time.Minute
	dbParameterGroupParametersDeleteRetryTimeout = 30 * time.Second
)

// @SDKResource("aws_neptune_parameter_group", name="Parameter Group")
// @Tags(identifierAttribute="arn")
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

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &neptune.CreateDBParameterGroupInput{
		DBParameterGroupFamily: aws.String(d.Get(names.AttrFamily).(string)),
		DBParameterGroupName:   aws.String(name),
		Description:            aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                   getTagsIn(ctx),
	}

	output, err := conn.CreateDBParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Parameter Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.DBParameterGroup.DBParameterGroupName))

	if v, ok := d.GetOk(names.AttrParameter); ok && v.(*schema.Set).Len() > 0 {
		if err := addDBParameterGroupParameters(ctx, conn, d.Id(), expandParameters(v.(*schema.Set).List())); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	dbParameterGroup, err := findDBParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbParameterGroup.DBParameterGroupArn)
	d.Set(names.AttrDescription, dbParameterGroup.Description)
	d.Set(names.AttrFamily, dbParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbParameterGroup.DBParameterGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(dbParameterGroup.DBParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones.
	input := &neptune.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
		Source:               aws.String("user"),
	}

	parameters, err := findDBParameters(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrParameter, flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	if d.HasChange(names.AttrParameter) {
		o, n := d.GetChange(names.AttrParameter)
		os, ns := o.(*schema.Set), n.(*schema.Set)
		add, del := ns.Difference(os).List(), os.Difference(ns).List()

		if len(del) > 0 {
			if err := delDBParameterGroupParameters(ctx, conn, d.Id(), expandParameters(del)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		if len(add) > 0 {
			if err := addDBParameterGroupParameters(ctx, conn, d.Id(), expandParameters(add)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneClient(ctx)

	log.Printf("[DEBUG] Deleting Neptune Parameter Group: %s", d.Id())
	_, err := tfresource.RetryWhenIsA[*awstypes.InvalidDBParameterGroupStateFault](ctx, dbParameterGroupDeleteRetryTimeout, func() (any, error) {
		return conn.DeleteDBParameterGroup(ctx, &neptune.DeleteDBParameterGroupInput{
			DBParameterGroupName: aws.String(d.Id()),
		})
	})

	if errs.IsA[*awstypes.DBParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func addDBParameterGroupParameters(ctx context.Context, conn *neptune.Client, name string, parameters []awstypes.Parameter) error { // We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for chunk := range slices.Chunk(parameters, dbParameterGroupMaxParamsBulkEdit) {
		input := &neptune.ModifyDBParameterGroupInput{
			DBParameterGroupName: aws.String(name),
			Parameters:           chunk,
		}

		_, err := conn.ModifyDBParameterGroup(ctx, input)

		if err != nil {
			return fmt.Errorf("modifying Neptune Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func delDBParameterGroupParameters(ctx context.Context, conn *neptune.Client, name string, parameters []awstypes.Parameter) error { // We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for chunk := range slices.Chunk(parameters, dbParameterGroupMaxParamsBulkEdit) {
		input := &neptune.ResetDBParameterGroupInput{
			DBParameterGroupName: aws.String(name),
			Parameters:           chunk,
		}

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidDBParameterGroupStateFault](ctx, dbParameterGroupParametersDeleteRetryTimeout, func() (any, error) {
			return conn.ResetDBParameterGroup(ctx, input)
		}, "has pending changes")

		if err != nil {
			return fmt.Errorf("resetting Neptune Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func findDBParameterGroupByName(ctx context.Context, conn *neptune.Client, name string) (*awstypes.DBParameterGroup, error) {
	input := &neptune.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(name),
	}
	output, err := findDBParameterGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.DBParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBParameterGroup(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBParameterGroupsInput) (*awstypes.DBParameterGroup, error) {
	output, err := findDBParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBParameterGroups(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBParameterGroupsInput) ([]awstypes.DBParameterGroup, error) {
	var output []awstypes.DBParameterGroup

	pages := neptune.NewDescribeDBParameterGroupsPaginator(conn, input)
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

		output = append(output, page.DBParameterGroups...)
	}

	return output, nil
}

func findDBParameters(ctx context.Context, conn *neptune.Client, input *neptune.DescribeDBParametersInput) ([]awstypes.Parameter, error) {
	var output []awstypes.Parameter

	pages := neptune.NewDescribeDBParametersPaginator(conn, input)

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

		output = append(output, page.Parameters...)
	}

	return output, nil
}
