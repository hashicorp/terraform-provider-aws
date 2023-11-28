// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"
	"time"

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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// We can only modify 20 parameters at a time, so walk them until
// we've got them all.
const maxParams = 20

// @SDKResource("aws_neptune_parameter_group", name="Parameter Group")
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

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &neptune.CreateDBParameterGroupInput{
		DBParameterGroupFamily: aws.String(d.Get("family").(string)),
		DBParameterGroupName:   aws.String(name),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   getTagsIn(ctx),
	}

	output, err := conn.CreateDBParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Parameter Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.DBParameterGroup.DBParameterGroupName))
	d.Set("arn", output.DBParameterGroup.DBParameterGroupArn)

	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	dbParameterGroup, err := FindDBParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Neptune Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", dbParameterGroup.DBParameterGroupArn)
	d.Set("description", dbParameterGroup.Description)
	d.Set("family", dbParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbParameterGroup.DBParameterGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(dbParameterGroup.DBParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones,
	input := &neptune.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
		Source:               aws.String("user"),
	}

	parameters, err := findDBParameters(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	return diags
}

func resourceParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

		toRemove := expandParameters(os.Difference(ns).List())

		log.Printf("[DEBUG] Parameters to remove: %#v", toRemove)

		toAdd := expandParameters(ns.Difference(os).List())

		log.Printf("[DEBUG] Parameters to add: %#v", toAdd)

		for len(toRemove) > 0 {
			var paramsToModify []*neptune.Parameter
			if len(toRemove) <= maxParams {
				paramsToModify, toRemove = toRemove[:], nil
			} else {
				paramsToModify, toRemove = toRemove[:maxParams], toRemove[maxParams:]
			}
			resetOpts := neptune.ResetDBParameterGroupInput{
				DBParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:           paramsToModify,
			}

			log.Printf("[DEBUG] Reset Neptune Parameter Group: %s", resetOpts)
			err := retry.RetryContext(ctx, 30*time.Second, func() *retry.RetryError {
				_, err := conn.ResetDBParameterGroupWithContext(ctx, &resetOpts)
				if err != nil {
					if tfawserr.ErrMessageContains(err, "InvalidDBParameterGroupState", " has pending changes") {
						return retry.RetryableError(err)
					}
					return retry.NonRetryableError(err)
				}
				return nil
			})
			if tfresource.TimedOut(err) {
				_, err = conn.ResetDBParameterGroupWithContext(ctx, &resetOpts)
			}
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resetting Neptune Parameter Group: %s", err)
			}
		}

		for len(toAdd) > 0 {
			var paramsToModify []*neptune.Parameter
			if len(toAdd) <= maxParams {
				paramsToModify, toAdd = toAdd[:], nil
			} else {
				paramsToModify, toAdd = toAdd[:maxParams], toAdd[maxParams:]
			}
			modifyOpts := neptune.ModifyDBParameterGroupInput{
				DBParameterGroupName: aws.String(d.Get("name").(string)),
				Parameters:           paramsToModify,
			}

			log.Printf("[DEBUG] Modify Neptune Parameter Group: %s", modifyOpts)
			_, err := conn.ModifyDBParameterGroupWithContext(ctx, &modifyOpts)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying Neptune Parameter Group: %s", err)
			}
		}
	}

	return append(diags, resourceParameterGroupRead(ctx, d, meta)...)
}

func resourceParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	log.Printf("[DEBUG] Deleting Neptune Parameter Group: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 3*time.Minute, func() (interface{}, error) {
		return conn.DeleteDBParameterGroupWithContext(ctx, &neptune.DeleteDBParameterGroupInput{
			DBParameterGroupName: aws.String(d.Id()),
		})
	}, neptune.ErrCodeInvalidDBParameterGroupStateFault)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDBParameterGroupByName(ctx context.Context, conn *neptune.Neptune, name string) (*neptune.DBParameterGroup, error) {
	input := &neptune.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(name),
	}
	output, err := findDBParameterGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.DBParameterGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDBParameterGroup(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBParameterGroupsInput) (*neptune.DBParameterGroup, error) {
	output, err := findDBParameterGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findDBParameterGroups(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBParameterGroupsInput) ([]*neptune.DBParameterGroup, error) {
	var output []*neptune.DBParameterGroup

	err := conn.DescribeDBParameterGroupsPagesWithContext(ctx, input, func(page *neptune.DescribeDBParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBParameterGroups {
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

func findDBParameters(ctx context.Context, conn *neptune.Neptune, input *neptune.DescribeDBParametersInput) ([]*neptune.Parameter, error) {
	var output []*neptune.Parameter

	err := conn.DescribeDBParametersPagesWithContext(ctx, input, func(page *neptune.DescribeDBParametersOutput, lastPage bool) bool {
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
