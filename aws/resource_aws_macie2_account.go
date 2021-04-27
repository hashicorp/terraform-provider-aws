package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	errorMacie2AccountCreate   = "error enabling Macie2 Account: %w"
	errorMacie2AccountRead     = "error reading Macie2 Account (%s): %w"
	errorMacie2AccountUpdating = "error updating Macie2 Account (%s): %w"
	errorMacie2AccountDelete   = "error disabling Macie2 Account (%s): %w"
	errorMacie2AccountSetting  = "error setting `%s` for Macie2 Account (%s): %w"
)

func resourceAwsMacie2Account() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2AccountCreate,
		ReadWithoutTimeout:   resourceMacie2AccountRead,
		UpdateWithoutTimeout: resourceMacie2AccountUpdate,
		DeleteWithoutTimeout: resourceMacie2AccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"finding_publishing_frequency": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"FIFTEEN_MINUTES", "ONE_HOUR", "SIX_HOURS"}, false),
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"PAUSED", "ENABLED"}, false),
			},
			"service_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2AccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.EnableMacieInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("finding_publishing_frequency"); ok {
		input.FindingPublishingFrequency = aws.String(v.(string))
	}
	if v, ok := d.GetOk("status"); ok {
		input.Status = aws.String(v.(string))
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err = conn.EnableMacieWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, _ = conn.EnableMacieWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountCreate, err))
	}

	//lintignore:R015 // Allow legacy unstable ID usage in managed resource
	d.SetId(resource.UniqueId())

	return resourceMacie2AccountRead(ctx, d, meta)
}

func resourceMacie2AccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.GetMacieSessionInput{}

	resp, err := conn.GetMacieSessionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
		log.Printf("[WARN] Macie2 Account is not enabled, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountRead, d.Id(), err))
	}

	if err = d.Set("status", resp.Status); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountSetting, "status", d.Id(), err))
	}
	if err = d.Set("finding_publishing_frequency", resp.FindingPublishingFrequency); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountSetting, "finding_publishing_frequency", d.Id(), err))
	}
	if err = d.Set("service_role", resp.ServiceRole); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountSetting, "service_role", d.Id(), err))
	}
	if err = d.Set("created_at", resp.CreatedAt.String()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountSetting, "created_at", d.Id(), err))
	}
	if err = d.Set("updated_at", resp.UpdatedAt.String()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountSetting, "updated_at", d.Id(), err))
	}

	return nil
}

func resourceMacie2AccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateMacieSessionInput{}

	if d.HasChange("finding_publishing_frequency") {
		input.FindingPublishingFrequency = aws.String(d.Get("finding_publishing_frequency").(string))
	}

	if d.HasChange("status") {
		input.Status = aws.String(d.Get("status").(string))
	}

	_, err := conn.UpdateMacieSessionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2AccountUpdating, d.Id(), err))
	}

	return resourceMacie2AccountRead(ctx, d, meta)
}

func resourceMacie2AccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DisableMacieInput{}

	_, err := conn.DisableMacieWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeInternalError) {
			return nil
		}
		return diag.FromErr(fmt.Errorf(errorMacie2AccountDelete, d.Id(), err))
	}
	return nil
}
