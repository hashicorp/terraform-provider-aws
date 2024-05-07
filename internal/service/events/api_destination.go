// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_event_api_destination", name="API Destination")
func resourceAPIDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIDestinationCreate,
		ReadWithoutTimeout:   resourceAPIDestinationRead,
		UpdateWithoutTimeout: resourceAPIDestinationUpdate,
		DeleteWithoutTimeout: resourceAPIDestinationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"http_method": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ApiDestinationHttpMethod](),
			},
			"invocation_endpoint": {
				Type:     schema.TypeString,
				Required: true,
			},
			"invocation_rate_limit_per_second": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Default:      300,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+`), ""),
				),
			},
		},
	}
}

func resourceAPIDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &eventbridge.CreateApiDestinationInput{
		ConnectionArn: aws.String(d.Get("connection_arn").(string)),
		HttpMethod:    types.ApiDestinationHttpMethod(d.Get("http_method").(string)),
		Name:          aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int32(int32(v.(int)))
	}

	_, err := conn.CreateApiDestination(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EventBridge API Destination (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourceAPIDestinationRead(ctx, d, meta)...)
}

func resourceAPIDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	output, err := findAPIDestinationByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EventBridge API Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge API Destination (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.ApiDestinationArn)
	d.Set("connection_arn", output.ConnectionArn)
	d.Set(names.AttrDescription, output.Description)
	d.Set("http_method", output.HttpMethod)
	d.Set("invocation_endpoint", output.InvocationEndpoint)
	d.Set("invocation_rate_limit_per_second", output.InvocationRateLimitPerSecond)
	d.Set(names.AttrName, output.Name)

	return diags
}

func resourceAPIDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	input := &eventbridge.UpdateApiDestinationInput{
		ConnectionArn: aws.String(d.Get("connection_arn").(string)),
		HttpMethod:    types.ApiDestinationHttpMethod(d.Get("http_method").(string)),
		Name:          aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int32(int32(v.(int)))
	}

	_, err := conn.UpdateApiDestination(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge API Destination (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAPIDestinationRead(ctx, d, meta)...)
}

func resourceAPIDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	log.Printf("[INFO] Deleting EventBridge API Destination: %s", d.Id())
	_, err := conn.DeleteApiDestination(ctx, &eventbridge.DeleteApiDestinationInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge API Destination (%s): %s", d.Id(), err)
	}

	return diags
}

func findAPIDestinationByName(ctx context.Context, conn *eventbridge.Client, name string) (*eventbridge.DescribeApiDestinationOutput, error) {
	input := &eventbridge.DescribeApiDestinationInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeApiDestination(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
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
