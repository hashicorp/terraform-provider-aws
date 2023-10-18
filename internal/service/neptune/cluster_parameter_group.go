// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const clusterParameterGroupMaxParamsBulkEdit = 20

// @SDKResource("aws_neptune_cluster_parameter_group", name="Cluster Parameter Group")
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
							Default:      neptune.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice(neptune.ApplyMethod_Values(), false),
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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = id.PrefixedUniqueId(v.(string))
	} else {
		groupName = id.UniqueId()
	}

	input := &neptune.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        getTagsIn(ctx),
	}

	_, err := conn.CreateDBClusterParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster Parameter Group (%s): %s", groupName, err)
	}

	d.SetId(groupName)

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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	dbClusterParameterGroup, err := FindDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", dbClusterParameterGroup.Description)
	d.Set("family", dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbClusterParameterGroup.DBClusterParameterGroupName)

	// Only include user customized parameters as there's hundreds of system/default ones
	input := &neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}
	var parameters []*neptune.Parameter

	err = conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *neptune.DescribeDBClusterParametersOutput, lastPage bool) bool {
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
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	// Add only system parameters that are set in the config
	p := d.Get("parameter")
	if p == nil {
		p = new(schema.Set)
	}
	s := p.(*schema.Set)
	configParameters := expandParameters(s.List())

	input = &neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("engine-default"),
	}

	err = conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *neptune.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			if v != nil {
				for _, p := range configParameters {
					if aws.StringValue(v.ParameterName) == aws.StringValue(p.ParameterName) {
						parameters = append(parameters, v)
					}
				}
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	log.Printf("[DEBUG] Deleting Neptune Cluster Parameter Group: %s", d.Id())
	_, err := conn.DeleteDBClusterParameterGroupWithContext(ctx, &neptune.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func modifyClusterParameterGroupParameters(ctx context.Context, conn *neptune.Neptune, name string, parameters []*neptune.Parameter) error {
	// We can only modify 20 parameters at a time, so chunk them until we've got them all.
	for _, chunk := range tfslices.Chunks(parameters, clusterParameterGroupMaxParamsBulkEdit) {
		input := &neptune.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  chunk,
		}

		_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("modifying Neptune Cluster Parameter Group (%s): %w", name, err)
		}
	}

	return nil
}

func FindDBClusterParameterGroupByName(ctx context.Context, conn *neptune.Neptune, name string) (*neptune.DBClusterParameterGroup, error) {
	input := &neptune.DescribeDBClusterParameterGroupsInput{
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

func findDBClusterParameterGroup(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBClusterParameterGroupsInput) (*neptune.DBClusterParameterGroup, error) {
	output, err := findDBClusterParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBClusterParameterGroups(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBClusterParameterGroupsInput) ([]*neptune.DBClusterParameterGroup, error) {
	var output []*neptune.DBClusterParameterGroup

	err := conn.DescribeDBClusterParameterGroupsPagesWithContext(ctx, input, func(page *neptune.DescribeDBClusterParameterGroupsOutput, lastPage bool) bool {
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

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
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
