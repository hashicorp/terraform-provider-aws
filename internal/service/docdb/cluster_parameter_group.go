// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/docdb/types"
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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_docdb_cluster_parameter_group", name="Cluster Parameter Group")
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &docdb.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(name),
		DBParameterGroupFamily:      aws.String(d.Get(names.AttrFamily).(string)),
		Description:                 aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                        getTagsIn(ctx),
	}

	_, err := conn.CreateDBClusterParameterGroup(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk(names.AttrParameter); ok && v.(*schema.Set).Len() > 0 {
		err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), expandParameters(v.(*schema.Set).List()))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	dbClusterParameterGroup, err := findDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set(names.AttrDescription, dbClusterParameterGroup.Description)
	d.Set(names.AttrFamily, dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(dbClusterParameterGroup.DBClusterParameterGroupName)))

	input := &docdb.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrParameter, flattenParameters(parameters, d.Get(names.AttrParameter).(*schema.Set).List())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

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

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBClient(ctx)

	log.Printf("[DEBUG] Deleting DocumentDB Cluster Parameter Group: %s", d.Id())
	_, err := conn.DeleteDBClusterParameterGroup(ctx, &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.DBParameterGroupNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 10*time.Minute, func() (interface{}, error) {
		return findDBClusterParameterGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Parameter Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func modifyClusterParameterGroupParameters(ctx context.Context, conn *docdb.Client, name string, parameters []awstypes.Parameter) error {
	const (
		clusterParameterGroupMaxParamsBulkEdit = 20
	)
	// We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for _, chunk := range tfslices.Chunks(parameters, clusterParameterGroupMaxParamsBulkEdit) {
		input := &docdb.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  chunk,
		}

		_, err := conn.ModifyDBClusterParameterGroup(ctx, input)

		if err != nil {
			return fmt.Errorf("modifying DocumentDB Cluster Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func findDBClusterParameterGroupByName(ctx context.Context, conn *docdb.Client, name string) (*awstypes.DBClusterParameterGroup, error) {
	input := &docdb.DescribeDBClusterParameterGroupsInput{
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

func findDBClusterParameterGroup(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClusterParameterGroupsInput) (*awstypes.DBClusterParameterGroup, error) {
	output, err := findDBClusterParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findDBClusterParameterGroups(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClusterParameterGroupsInput) ([]awstypes.DBClusterParameterGroup, error) {
	var output []awstypes.DBClusterParameterGroup

	pages := docdb.NewDescribeDBClusterParameterGroupsPaginator(conn, input)
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

		for _, v := range page.DBClusterParameterGroups {
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findDBClusterParameters(ctx context.Context, conn *docdb.Client, input *docdb.DescribeDBClusterParametersInput) ([]awstypes.Parameter, error) {
	var output []awstypes.Parameter

	pages := docdb.NewDescribeDBClusterParametersPaginator(conn, input)
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
			if !reflect.ValueOf(v).IsZero() {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
