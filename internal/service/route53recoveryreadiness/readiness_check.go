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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReadinessCheck() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReadinessCheckCreate,
		ReadWithoutTimeout:   resourceReadinessCheckRead,
		UpdateWithoutTimeout: resourceReadinessCheckUpdate,
		DeleteWithoutTimeout: resourceReadinessCheckDelete,
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
			"readiness_check_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_set_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReadinessCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53recoveryreadiness.CreateReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	resp, err := conn.CreateReadinessCheckWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	d.SetId(aws.StringValue(resp.ReadinessCheckName))

	if len(tags) > 0 {
		arn := aws.StringValue(resp.ReadinessCheckArn)
		if err := UpdateTags(ctx, conn, arn, nil, tags); err != nil {
			return sdkdiag.AppendErrorf(diags, "adding Route53 Recovery Readiness ReadinessCheck (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}

	resp, err := conn.GetReadinessCheckWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Route53RecoveryReadiness Readiness Check (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	d.Set("arn", resp.ReadinessCheckArn)
	d.Set("readiness_check_name", resp.ReadinessCheckName)
	d.Set("resource_set_name", resp.ResourceSet)

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Route53 Recovery Readiness ReadinessCheck (%s): %s", d.Id(), err)
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

func resourceReadinessCheckUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()

	input := &route53recoveryreadiness.UpdateReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Get("readiness_check_name").(string)),
		ResourceSetName:    aws.String(d.Get("resource_set_name").(string)),
	}

	_, err := conn.UpdateReadinessCheckWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		arn := d.Get("arn").(string)
		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Recovery Readiness ReadinessCheck (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceReadinessCheckRead(ctx, d, meta)...)
}

func resourceReadinessCheckDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53RecoveryReadinessConn()

	input := &route53recoveryreadiness.DeleteReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}
	_, err := conn.DeleteReadinessCheckWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Recovery Readiness ReadinessCheck: %s", err)
	}

	gcinput := &route53recoveryreadiness.GetReadinessCheckInput{
		ReadinessCheckName: aws.String(d.Id()),
	}
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.GetReadinessCheckWithContext(ctx, gcinput)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, route53recoveryreadiness.ErrCodeResourceNotFoundException) {
				return nil
			}
			return resource.NonRetryableError(err)
		}
		return resource.RetryableError(fmt.Errorf("Route 53 Recovery Readiness ReadinessCheck (%s) still exists", d.Id()))
	})

	if tfresource.TimedOut(err) {
		_, err = conn.GetReadinessCheckWithContext(ctx, gcinput)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route 53 Recovery Readiness ReadinessCheck (%s) deletion: %s", d.Id(), err)
	}

	return diags
}
