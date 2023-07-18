// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_cloudwatch_event_api_destination")
func ResourceAPIDestination() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIDestinationCreate,
		ReadWithoutTimeout:   resourceAPIDestinationRead,
		UpdateWithoutTimeout: resourceAPIDestinationUpdate,
		DeleteWithoutTimeout: resourceAPIDestinationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[\.\-_A-Za-z0-9]+`), ""),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
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
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(eventbridge.ApiDestinationHttpMethod_Values(), true),
			},
			"connection_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAPIDestinationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := &eventbridge.CreateApiDestinationInput{}

	if name, ok := d.GetOk("name"); ok {
		input.Name = aws.String(name.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if invocationEndpoint, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(invocationEndpoint.(string))
	}
	if invocationRateLimitPerSecond, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int64(int64(invocationRateLimitPerSecond.(int)))
	}
	if httpMethod, ok := d.GetOk("http_method"); ok {
		input.HttpMethod = aws.String(httpMethod.(string))
	}
	if connectionArn, ok := d.GetOk("connection_arn"); ok {
		input.ConnectionArn = aws.String(connectionArn.(string))
	}

	_, err := conn.CreateApiDestinationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Creating EventBridge API Destination (%s) failed: %s", aws.StringValue(input.Name), err)
	}

	d.SetId(aws.StringValue(input.Name))

	log.Printf("[INFO] EventBridge API Destination (%s) created", d.Id())

	return append(diags, resourceAPIDestinationRead(ctx, d, meta)...)
}

func resourceAPIDestinationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := &eventbridge.DescribeApiDestinationInput{
		Name: aws.String(d.Id()),
	}

	output, err := conn.DescribeApiDestinationWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge API Destination (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge API Destination (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.ApiDestinationArn)
	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("invocation_endpoint", output.InvocationEndpoint)
	d.Set("invocation_rate_limit_per_second", output.InvocationRateLimitPerSecond)
	d.Set("http_method", output.HttpMethod)
	d.Set("connection_arn", output.ConnectionArn)

	return diags
}

func resourceAPIDestinationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	input := &eventbridge.UpdateApiDestinationInput{}

	if name, ok := d.GetOk("name"); ok {
		input.Name = aws.String(name.(string))
	}
	if description, ok := d.GetOk("description"); ok {
		input.Description = aws.String(description.(string))
	}
	if invocationEndpoint, ok := d.GetOk("invocation_endpoint"); ok {
		input.InvocationEndpoint = aws.String(invocationEndpoint.(string))
	}
	if invocationRateLimitPerSecond, ok := d.GetOk("invocation_rate_limit_per_second"); ok {
		input.InvocationRateLimitPerSecond = aws.Int64(int64(invocationRateLimitPerSecond.(int)))
	}
	if httpMethod, ok := d.GetOk("http_method"); ok {
		input.HttpMethod = aws.String(httpMethod.(string))
	}
	if connectionArn, ok := d.GetOk("connection_arn"); ok {
		input.ConnectionArn = aws.String(connectionArn.(string))
	}

	log.Printf("[DEBUG] Updating EventBridge API Destination: %s", input)
	_, err := conn.UpdateApiDestinationWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EventBridge API Destination (%s): %s", d.Id(), err)
	}
	return append(diags, resourceAPIDestinationRead(ctx, d, meta)...)
}

func resourceAPIDestinationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsConn(ctx)

	log.Printf("[INFO] Deleting EventBridge API Destination (%s)", d.Id())
	input := &eventbridge.DeleteApiDestinationInput{
		Name: aws.String(d.Id()),
	}

	_, err := conn.DeleteApiDestinationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, eventbridge.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] EventBridge API Destination (%s) not found", d.Id())
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EventBridge API Destination (%s): %s", d.Id(), err)
	}
	log.Printf("[INFO] EventBridge API Destination (%s) deleted", d.Id())

	return diags
}
