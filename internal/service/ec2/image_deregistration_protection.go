// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_ec2_image_deregistration_protection", name="Image Deregistration Protection")
func resourceImageDeregistrationProtection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImageDeregistrationProtectionRequestCreate,
		ReadWithoutTimeout:   resourceImageDeregistrationProtectionRequestRead,
		DeleteWithoutTimeout: resourceImageDeregistrationProtectionRequestDelete,
		UpdateWithoutTimeout: resourceImageDeregistrationProtectionRequestUpdate,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"ami_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"with_cooldown": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"deregistration_protection": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceImageDeregistrationProtectionRequestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.EnableImageDeregistrationProtectionInput{
		ImageId:      aws.String(d.Get("ami_id").(string)),
		WithCooldown: aws.Bool(d.Get("with_cooldown").(bool)),
	}
	_, err := conn.EnableImageDeregistrationProtection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Enabling Image Deregistration Protection for AMI ID (%s): %s", d.Get("ami_id").(string), err)
	}
	d.SetId(d.Get("ami_id").(string))
	return append(diags, resourceImageDeregistrationProtectionRequestRead(ctx, d, meta)...)
}

func resourceImageDeregistrationProtectionRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := checkDeregistrationProtection(ctx, conn, &ec2.DescribeImagesInput{
		ImageIds: []string{d.Id()},
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Reading the ami id (%s): %s", d.Id(), err)
	}

	d.Set("ami_id", d.Id())
	d.Set("with_cooldown", d.Get("with_cooldown").(bool))
	d.Set("deregistration_protection", output.Images[0].DeregistrationProtection)

	return diags
}

func resourceImageDeregistrationProtectionRequestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.EnableImageDeregistrationProtectionInput{
		ImageId:      aws.String(d.Id()),
		WithCooldown: aws.Bool(d.Get("with_cooldown").(bool)),
	}

	_, err := conn.EnableImageDeregistrationProtection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Updating Image Deregistration Protection for AMI ID (%s): %s", d.Get("ami_id").(string), err)
	}
	return append(diags, resourceImageDeregistrationProtectionRequestRead(ctx, d, meta)...)
}

func resourceImageDeregistrationProtectionRequestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DisableImageDeregistrationProtectionInput{
		ImageId: aws.String(d.Id()),
	}
	log.Printf("[INFO] Disabling image deregistration protection for ami id: %s", d.Id())

	_, err := conn.DisableImageDeregistrationProtection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Disabling Image Deregistration Protection for AMI ID (%s): %s", d.Id(), err)
	}
	return diags
}

func checkDeregistrationProtection(ctx context.Context, conn *ec2.Client, input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	output, err := conn.DescribeImages(ctx, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}
