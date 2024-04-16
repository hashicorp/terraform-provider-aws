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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      docdb.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice(docdb.ApplyMethod_Values(), false),
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

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &docdb.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(name),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        getTagsIn(ctx),
	}

	_, err := conn.CreateDBClusterParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DocumentDB Cluster Parameter Group (%s): %s", name, err)
	}

	d.SetId(name)

	if v, ok := d.GetOk("parameter"); ok && v.(*schema.Set).Len() > 0 {
		err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), expandParameters(v.(*schema.Set).List()))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	dbClusterParameterGroup, err := FindDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DocumentDB Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set("description", dbClusterParameterGroup.Description)
	d.Set("family", dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName)))

	input := &docdb.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DocumentDB Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(parameters, d.Get("parameter").(*schema.Set).List())); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
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
	conn := meta.(*conns.AWSClient).DocDBConn(ctx)

	log.Printf("[DEBUG] Deleting DocumentDB Cluster Parameter Group: %s", d.Id())
	_, err := conn.DeleteDBClusterParameterGroupWithContext(ctx, &docdb.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DocumentDB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, 10*time.Minute, func() (interface{}, error) {
		return FindDBClusterParameterGroupByName(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DocumentDB Cluster Parameter Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func modifyClusterParameterGroupParameters(ctx context.Context, conn *docdb.DocDB, name string, parameters []*docdb.Parameter) error {
	const (
		clusterParameterGroupMaxParamsBulkEdit = 20
	)
	// We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for _, chunk := range tfslices.Chunks(parameters, clusterParameterGroupMaxParamsBulkEdit) {
		input := &docdb.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  chunk,
		}

		_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("modifying DocumentDB Cluster Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func FindDBClusterParameterGroupByName(ctx context.Context, conn *docdb.DocDB, name string) (*docdb.DBClusterParameterGroup, error) {
	input := &docdb.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}
	output, err := findDBClusterParameterGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBClusterParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBClusterParameterGroup(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeDBClusterParameterGroupsInput) (*docdb.DBClusterParameterGroup, error) {
	output, err := findDBClusterParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBClusterParameterGroups(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeDBClusterParameterGroupsInput) ([]*docdb.DBClusterParameterGroup, error) {
	var output []*docdb.DBClusterParameterGroup

	err := conn.DescribeDBClusterParameterGroupsPagesWithContext(ctx, input, func(page *docdb.DescribeDBClusterParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusterParameterGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findDBClusterParameters(ctx context.Context, conn *docdb.DocDB, input *docdb.DescribeDBClusterParametersInput) ([]*docdb.Parameter, error) {
	var output []*docdb.Parameter

	err := conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *docdb.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeDBParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
