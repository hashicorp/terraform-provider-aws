// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_qbusiness_index", name="Index")
// @Tags(identifierAttribute="arn")
func ResourceIndex() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceIndexCreate,
		ReadWithoutTimeout:   resourceIndexRead,
		UpdateWithoutTimeout: resourceIndexUpdate,
		DeleteWithoutTimeout: resourceIndexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the Amazon Q application associated with the index.",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{35}$`), "must be a valid application ID"),
				),
			},
			"arn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Amazon Resource Name (ARN) of the Amazon Q index.",
			},
			"capacity_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"units": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The number of additional storage units for the Amazon Q index.",
						},
					},
				},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description for the Amazon Q index.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 1000),
					validation.StringMatch(regexache.MustCompile(`^\P{C}*$`), "must not contain control characters"),
				),
			},
			"display_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Amazon Q application.",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`), "must begin with a letter or number and contain only alphanumeric, underscore, or hyphen characters"),
				),
			},
			"index_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier of the Amazon Q index.",
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	application_id := d.Get("application_id").(string)
	display_name := d.Get("display_name").(string)

	input := &qbusiness.CreateIndexInput{
		ApplicationId: aws.String(application_id),
		DisplayName:   aws.String(display_name),
	}

	input.Tags = getTagsIn(ctx)

	if v, ok := d.GetOk("capacity_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CapacityConfiguration = &qbusiness.IndexCapacityConfiguration{
			Units: aws.Int64(int64(v.([]interface{})[0].(map[string]interface{})["units"].(int))),
		}
	}

	output, err := conn.CreateIndexWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating qbusiness index: %s", err)
	}

	d.SetId(aws.StringValue(output.IndexId))

	if _, err := waitIndexCreated(ctx, conn, application_id, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for qbusiness index (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceAppRead(ctx, d, meta)...)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func FindIndexById(ctx context.Context, application_id, index_id string, meta interface{}) (*qbusiness.GetIndexOutput, error) {
	input := &qbusiness.GetIndexInput{
		ApplicationId: aws.String(application_id),
		IndexId:       aws.String(index_id),
	}

	conn := meta.(*conns.AWSClient).QBusinessConn(ctx)

	output, err := conn.GetIndexWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, qbusiness.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
