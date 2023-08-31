// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
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
							Default:  neptune.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								neptune.ApplyMethodImmediate,
								neptune.ApplyMethodPendingReboot,
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

func resourceParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	createOpts := neptune.CreateDBParameterGroupInput{
		DBParameterGroupName:   aws.String(d.Get("name").(string)),
		DBParameterGroupFamily: aws.String(d.Get("family").(string)),
		Description:            aws.String(d.Get("description").(string)),
		Tags:                   getTagsIn(ctx),
	}

	log.Printf("[DEBUG] Create Neptune Parameter Group: %#v", createOpts)
	resp, err := conn.CreateDBParameterGroupWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Parameter Group: %s", err)
	}

	d.SetId(aws.StringValue(resp.DBParameterGroup.DBParameterGroupName))
	d.Set("arn", resp.DBParameterGroup.DBParameterGroupArn)
	log.Printf("[INFO] Neptune Parameter Group ID: %s", d.Id())

	return append(diags, resourceParameterGroupUpdate(ctx, d, meta)...)
}

func resourceParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn(ctx)

	describeOpts := neptune.DescribeDBParameterGroupsInput{
		DBParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBParameterGroupsWithContext(ctx, &describeOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
			log.Printf("[WARN] Neptune Parameter Group (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s): %s", d.Id(), err)
	}

	if describeResp == nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s): empty result", d.Id())
	}

	if len(describeResp.DBParameterGroups) != 1 ||
		aws.StringValue(describeResp.DBParameterGroups[0].DBParameterGroupName) != d.Id() {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Parameter Group (%s): no match", d.Id())
	}

	arn := aws.StringValue(describeResp.DBParameterGroups[0].DBParameterGroupArn)
	d.Set("arn", arn)
	d.Set("name", describeResp.DBParameterGroups[0].DBParameterGroupName)
	d.Set("family", describeResp.DBParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", describeResp.DBParameterGroups[0].Description)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := neptune.DescribeDBParametersInput{
		DBParameterGroupName: aws.String(d.Id()),
		Source:               aws.String("user"),
	}

	var parameters []*neptune.Parameter
	err = conn.DescribeDBParametersPagesWithContext(ctx, &describeParametersOpts,
		func(describeParametersResp *neptune.DescribeDBParametersOutput, lastPage bool) bool {
			parameters = append(parameters, describeParametersResp.Parameters...)
			return !lastPage
		})
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

	deleteOpts := neptune.DeleteDBParameterGroupInput{
		DBParameterGroupName: aws.String(d.Id()),
	}
	err := retry.RetryContext(ctx, 3*time.Minute, func() *retry.RetryError {
		_, err := conn.DeleteDBParameterGroupWithContext(ctx, &deleteOpts)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
				return nil
			}
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeInvalidDBParameterGroupStateFault) {
				return retry.RetryableError(err)
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDBParameterGroupWithContext(ctx, &deleteOpts)
	}

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Parameter Group: %s", err)
	}

	return diags
}
