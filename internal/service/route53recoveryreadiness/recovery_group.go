package route53recoveryreadiness

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53recoveryreadiness"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRecoveryGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRecoveryGroupCreate,
		ReadWithoutTimeout:   resourceRecoveryGroupRead,
		UpdateWithoutTimeout: resourceRecoveryGroupUpdate,
		DeleteWithoutTimeout: resourceRecoveryGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cells": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"recovery_group_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRecoveryGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53recoveryreadiness.CreateRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Get("recovery_group_name").(string)),
		Cells:             flex.ExpandStringList(d.Get("cells").([]interface{})),
	}

	resp, err := conn.CreateRecoveryGroupWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	d.SetId(aws.StringValue(resp.RecoveryGroupName))

	if len(tags) > 0 {
		arn := aws.StringValue(resp.RecoveryGroupArn)
		if err := UpdateTags(ctx, conn, arn, nil, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Route53 Recovery Readiness RecoveryGroup (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRecoveryGroupRead(ctx, d, meta)...)
}

func resourceRecoveryGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &route53recoveryreadiness.GetRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}
	resp, err := conn.GetRecoveryGroupWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53RecoveryReadiness Recovery Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	d.Set("arn", resp.RecoveryGroupArn)
	d.Set("recovery_group_name", resp.RecoveryGroupName)
	d.Set("cells", resp.Cells)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Route53 Recovery Readiness RecoveryGroup (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceRecoveryGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()

	input := &route53recoveryreadiness.UpdateRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
		Cells:             flex.ExpandStringList(d.Get("cells").([]interface{})),
	}

	_, err := conn.UpdateRecoveryGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)
		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness RecoveryGroup (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceRecoveryGroupRead(ctx, d, meta)...)
}

func resourceRecoveryGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()

	input := &route53recoveryreadiness.DeleteRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteRecoveryGroupWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness RecoveryGroup: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetRecoveryGroupInput{
		RecoveryGroupName: aws.String(d.Id()),
	}

	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetRecoveryGroupWithContext(ctx, gcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness RecoveryGroup (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetRecoveryGroupWithContext(ctx, gcinput)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness RecoveryGroup (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
