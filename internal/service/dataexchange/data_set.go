// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dataexchange_data_set", name="Data Set")
// @Tags(identifierAttribute="arn")
func ResourceDataSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSetCreate,
		ReadWithoutTimeout:   resourceDataSetRead,
		UpdateWithoutTimeout: resourceDataSetUpdate,
		DeleteWithoutTimeout: resourceDataSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(dataexchange.AssetType_Values(), false),
			},
			"description": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 16348),
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDataSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeConn(ctx)

	input := &dataexchange.CreateDataSetInput{
		Name:        aws.String(d.Get("name").(string)),
		AssetType:   aws.String(d.Get("asset_type").(string)),
		Description: aws.String(d.Get("description").(string)),
		Tags:        getTagsIn(ctx),
	}

	out, err := conn.CreateDataSetWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataExchange DataSet: %s", err)
	}

	d.SetId(aws.StringValue(out.Id))

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeConn(ctx)

	dataSet, err := FindDataSetById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataExchange DataSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataExchange DataSet (%s): %s", d.Id(), err)
	}

	d.Set("asset_type", dataSet.AssetType)
	d.Set("name", dataSet.Name)
	d.Set("description", dataSet.Description)
	d.Set("arn", dataSet.Arn)

	setTagsOut(ctx, dataSet.Tags)

	return diags
}

func resourceDataSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &dataexchange.UpdateDataSetInput{
			DataSetId: aws.String(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		log.Printf("[DEBUG] Updating DataExchange DataSet: %s", d.Id())
		_, err := conn.UpdateDataSetWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataExchange DataSet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataSetRead(ctx, d, meta)...)
}

func resourceDataSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeConn(ctx)

	input := &dataexchange.DeleteDataSetInput{
		DataSetId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataExchange DataSet: %s", d.Id())
	_, err := conn.DeleteDataSetWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DataExchange DataSet: %s", err)
	}

	return diags
}
