// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_tag_option")
func ResourceTagOption() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTagOptionCreate,
		ReadWithoutTimeout:   resourceTagOptionRead,
		UpdateWithoutTimeout: resourceTagOptionUpdate,
		DeleteWithoutTimeout: resourceTagOptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(TagOptionReadyTimeout),
			Read:   schema.DefaultTimeout(TagOptionReadTimeout),
			Update: schema.DefaultTimeout(TagOptionUpdateTimeout),
			Delete: schema.DefaultTimeout(TagOptionDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrKey: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrValue: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTagOptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.CreateTagOptionInput{
		Key:   aws.String(d.Get(names.AttrKey).(string)),
		Value: aws.String(d.Get(names.AttrValue).(string)),
	}

	var output *servicecatalog.CreateTagOptionOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.CreateTagOptionWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateTagOptionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Tag Option: %s", err)
	}

	if output == nil || output.TagOptionDetail == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Tag Option: empty response")
	}

	d.SetId(aws.StringValue(output.TagOptionDetail.Id))

	// Active is not a field of CreateTagOption but is a field of UpdateTagOption. In order to create an
	// inactive Tag Option, you must create an active one and then update it (but calling this resource's
	// Update will error with ErrCodeDuplicateResourceException because Value is unchanged).
	if v, ok := d.GetOk("active"); !ok {
		_, err = conn.UpdateTagOptionWithContext(ctx, &servicecatalog.UpdateTagOptionInput{
			Id:     aws.String(d.Id()),
			Active: aws.Bool(v.(bool)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Service Catalog Tag Option, updating active (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTagOptionRead(ctx, d, meta)...)
}

func resourceTagOptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := WaitTagOptionReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Service Catalog Tag Option (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Tag Option (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Tag Option (%s): empty response", d.Id())
	}

	d.Set("active", output.Active)
	d.Set(names.AttrKey, output.Key)
	d.Set(names.AttrOwner, output.Owner)
	d.Set(names.AttrValue, output.Value)

	return diags
}

func resourceTagOptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdateTagOptionInput{
		Id: aws.String(d.Id()),
	}

	// UpdateTagOption() is very particular about what it receives. Only fields that change should
	// be included or it will throw servicecatalog.ErrCodeDuplicateResourceException, "already exists"

	if d.HasChange("active") {
		input.Active = aws.Bool(d.Get("active").(bool))
	}

	if d.HasChange(names.AttrValue) {
		input.Value = aws.String(d.Get(names.AttrValue).(string))
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.UpdateTagOptionWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateTagOptionWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Tag Option (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTagOptionRead(ctx, d, meta)...)
}

func resourceTagOptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.DeleteTagOptionInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteTagOptionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Tag Option (%s): %s", d.Id(), err)
	}

	if err := WaitTagOptionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Tag Option (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
