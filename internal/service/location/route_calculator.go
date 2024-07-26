// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
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

// @SDKResource("aws_location_route_calculator", name="Route Calculator")
// @Tags(identifierAttribute="calculator_arn")
func ResourceRouteCalculator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCalculatorCreate,
		ReadWithoutTimeout:   resourceRouteCalculatorRead,
		UpdateWithoutTimeout: resourceRouteCalculatorUpdate,
		DeleteWithoutTimeout: resourceRouteCalculatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"calculator_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"calculator_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			names.AttrCreateTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1000),
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRouteCalculatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	in := &locationservice.CreateRouteCalculatorInput{
		CalculatorName: aws.String(d.Get("calculator_name").(string)),
		DataSource:     aws.String(d.Get("data_source").(string)),
		Tags:           getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateRouteCalculatorWithContext(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Route Calculator (%s): %s", d.Get("calculator_name").(string), err)
	}

	if out == nil {
		return sdkdiag.AppendErrorf(diags, "creating Location Service Route Calculator (%s): empty output", d.Get("calculator_name").(string))
	}

	d.SetId(aws.StringValue(out.CalculatorName))

	return append(diags, resourceRouteCalculatorRead(ctx, d, meta)...)
}

func resourceRouteCalculatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	out, err := findRouteCalculatorByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Location Service Route Calculator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	d.Set("calculator_arn", out.CalculatorArn)
	d.Set("calculator_name", out.CalculatorName)
	d.Set(names.AttrCreateTime, aws.TimeValue(out.CreateTime).Format(time.RFC3339))
	d.Set("data_source", out.DataSource)
	d.Set(names.AttrDescription, out.Description)
	d.Set("update_time", aws.TimeValue(out.UpdateTime).Format(time.RFC3339))

	return diags
}

func resourceRouteCalculatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	update := false

	in := &locationservice.UpdateRouteCalculatorInput{
		CalculatorName: aws.String(d.Get("calculator_name").(string)),
	}

	if d.HasChange(names.AttrDescription) {
		in.Description = aws.String(d.Get(names.AttrDescription).(string))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating Location Service Route Calculator (%s): %#v", d.Id(), in)
	_, err := conn.UpdateRouteCalculatorWithContext(ctx, in)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRouteCalculatorRead(ctx, d, meta)...)
}

func resourceRouteCalculatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LocationConn(ctx)

	log.Printf("[INFO] Deleting Location Service Route Calculator %s", d.Id())

	_, err := conn.DeleteRouteCalculatorWithContext(ctx, &locationservice.DeleteRouteCalculatorInput{
		CalculatorName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	return diags
}

func findRouteCalculatorByName(ctx context.Context, conn *locationservice.LocationService, name string) (*locationservice.DescribeRouteCalculatorOutput, error) {
	in := &locationservice.DescribeRouteCalculatorInput{
		CalculatorName: aws.String(name),
	}

	out, err := conn.DescribeRouteCalculatorWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, locationservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
