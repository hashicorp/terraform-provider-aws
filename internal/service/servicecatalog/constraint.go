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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_constraint")
func ResourceConstraint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceConstraintCreate,
		ReadWithoutTimeout:   resourceConstraintRead,
		UpdateWithoutTimeout: resourceConstraintUpdate,
		DeleteWithoutTimeout: resourceConstraintDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ConstraintReadyTimeout),
			Read:   schema.DefaultTimeout(ConstraintReadTimeout),
			Update: schema.DefaultTimeout(ConstraintUpdateTimeout),
			Delete: schema.DefaultTimeout(ConstraintDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      AcceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(AcceptLanguage_Values(), false),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrParameters: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"portfolio_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"product_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ConstraintType_Values(), false),
			},
		},
	}
}

func resourceConstraintCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.CreateConstraintInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		Parameters:       aws.String(d.Get(names.AttrParameters).(string)),
		PortfolioId:      aws.String(d.Get("portfolio_id").(string)),
		ProductId:        aws.String(d.Get("product_id").(string)),
		Type:             aws.String(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	var output *servicecatalog.CreateConstraintOutput
	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		var err error

		output, err = conn.CreateConstraintWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateConstraintWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Constraint: %s", err)
	}

	if output == nil || output.ConstraintDetail == nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Constraint: empty response")
	}

	d.SetId(aws.StringValue(output.ConstraintDetail.ConstraintId))

	return append(diags, resourceConstraintRead(ctx, d, meta)...)
}

func resourceConstraintRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	output, err := WaitConstraintReady(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Catalog Constraint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Constraint (%s): %s", d.Id(), err)
	}

	if output == nil || output.ConstraintDetail == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Constraint (%s): empty response", d.Id())
	}

	acceptLanguage := d.Get("accept_language").(string)

	if acceptLanguage == "" {
		acceptLanguage = AcceptLanguageEnglish
	}

	d.Set("accept_language", acceptLanguage)

	d.Set(names.AttrParameters, output.ConstraintParameters)
	d.Set(names.AttrStatus, output.Status)

	detail := output.ConstraintDetail

	d.Set(names.AttrDescription, detail.Description)
	d.Set(names.AttrOwner, detail.Owner)
	d.Set("portfolio_id", detail.PortfolioId)
	d.Set("product_id", detail.ProductId)
	d.Set(names.AttrType, detail.Type)

	return diags
}

func resourceConstraintUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.UpdateConstraintInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("accept_language") {
		input.AcceptLanguage = aws.String(d.Get("accept_language").(string))
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	if d.HasChange(names.AttrParameters) {
		input.Parameters = aws.String(d.Get(names.AttrParameters).(string))
	}

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutUpdate), func() *retry.RetryError {
		_, err := conn.UpdateConstraintWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, servicecatalog.ErrCodeInvalidParametersException, "profile does not exist") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.UpdateConstraintWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Constraint (%s): %s", d.Id(), err)
	}

	return append(diags, resourceConstraintRead(ctx, d, meta)...)
}

func resourceConstraintDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogConn(ctx)

	input := &servicecatalog.DeleteConstraintInput{
		Id: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	_, err := conn.DeleteConstraintWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Constraint (%s): %s", d.Id(), err)
	}

	err = WaitConstraintDeleted(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Constraint (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}
