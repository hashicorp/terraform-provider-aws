// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const clusterParameterGroupMaxParamsBulkEdit = 20

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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
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
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  docdb.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								docdb.ApplyMethodImmediate,
								docdb.ApplyMethodPendingReboot,
							}, false),
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
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = id.PrefixedUniqueId(v.(string))
	} else {
		groupName = id.UniqueId()
	}

	input := docdb.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        getTagsIn(ctx),
	}

	resp, err := conn.CreateDBClusterParameterGroupWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(input.DBClusterParameterGroupName))

	d.Set("arn", resp.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return append(diags, resourceClusterParameterGroupUpdate(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	describeOpts := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, describeOpts)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			log.Printf("[WARN] DocumentDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	if len(describeResp.DBClusterParameterGroups) != 1 ||
		aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName) != d.Id() {
		return sdkdiag.AppendErrorf(diags, "Unable to find Cluster Parameter Group: %#v", describeResp.DBClusterParameterGroups)
	}

	arn := aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", describeResp.DBClusterParameterGroups[0].Description)
	d.Set("family", describeResp.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set("name", describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName)

	describeParametersOpts := &docdb.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeParametersResp, err := conn.DescribeDBClusterParametersWithContext(ctx, describeParametersOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(describeParametersResp.Parameters, d.Get("parameter").(*schema.Set).List())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting DocumentDB cluster parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

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
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []*docdb.Parameter
				if len(parameters) <= clusterParameterGroupMaxParamsBulkEdit {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:clusterParameterGroupMaxParamsBulkEdit], parameters[clusterParameterGroupMaxParamsBulkEdit:]
				}
				parameterGroupName := d.Id()
				modifyOpts := docdb.ModifyDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(parameterGroupName),
					Parameters:                  paramsToModify,
				}

				_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, &modifyOpts)
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying DocumentDB Cluster Parameter Group: %s", err)
				}
			}
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	deleteOpts := &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBClusterParameterGroupWithContext(ctx, deleteOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	if err := WaitForClusterParameterGroupDeletion(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}
	return diags
}

func WaitForClusterParameterGroupDeletion(ctx context.Context, conn *docdb.DocDB, name string) error {
	params := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}

	err := retry.RetryContext(ctx, 10*time.Minute, func() *retry.RetryError {
		_, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, params)

		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return retry.RetryableError(fmt.Errorf("DocumentDB Parameter Group (%s) still exists", name))
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DescribeDBClusterParameterGroupsWithContext(ctx, params)
		if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
			return nil
		}
	}
	if err != nil {
		return fmt.Errorf("deleting DocumentDB cluster parameter group: %s", err)
	}
	return nil
}
