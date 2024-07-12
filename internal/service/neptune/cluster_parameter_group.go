// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
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
							Type:         schema.TypeString,
							Optional:     true,
							Default:      neptune.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice(neptune.ApplyMethod_Values(), false),
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
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &neptune.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(name),
		DBParameterGroupFamily:      aws.String(d.Get(names.AttrFamily).(string)),
		Description:                 aws.String(d.Get(names.AttrDescription).(string)),
		Tags:                        getTagsIn(ctx),
	}

	_, err := conn.CreateDBClusterParameterGroupWithContext(ctx, input)

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
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, dbClusterParameterGroup.Description)
	d.Set(names.AttrFamily, dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set(names.AttrName, dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones.
	input := &neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	parameters, err := findDBClusterParameters(ctx, conn, input, tfslices.PredicateTrue[*neptune.Parameter]())

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

	systemParameters, err := findDBClusterParameters(ctx, conn, input, func(v *neptune.Parameter) bool {
		return slices.ContainsFunc(configParameters, func(p *neptune.Parameter) bool {
			return aws.StringValue(v.ParameterName) == aws.StringValue(p.ParameterName)
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

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

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
	const (
		clusterParameterGroupMaxParamsBulkEdit = 20
	)
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

func findDBClusterParameters(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBClusterParametersInput, filter tfslices.Predicate[*neptune.Parameter]) ([]*neptune.Parameter, error) {
	var output []*neptune.Parameter

	err := conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *neptune.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			if v != nil && filter(v) {
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
