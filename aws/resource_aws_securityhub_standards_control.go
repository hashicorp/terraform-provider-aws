package aws

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfsecurityhub "github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/securityhub/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func ResourceStandardsControl() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsSecurityHubStandardsControlPut,
		ReadContext:   resourceStandardsControlRead,
		UpdateContext: resourceAwsSecurityHubStandardsControlPut,
		DeleteContext: resourceStandardsControlDelete,

		Schema: map[string]*schema.Schema{
			"control_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"control_status": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(securityhub.ControlStatus_Values(), false),
			},

			"control_status_updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"disabled_reason": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"related_requirements": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"remediation_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"severity_rating": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"standards_control_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"title": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStandardsControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	standardsSubscriptionARN, err := tfsecurityhub.StandardsControlARNToStandardsSubscriptionARN(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	control, err := finder.StandardsControlByStandardsSubscriptionARNAndStandardsControlARN(ctx, conn, standardsSubscriptionARN, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Standards Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading Security Hub Standards Control (%s): %s", d.Id(), err)
	}

	d.Set("control_id", control.ControlId)
	d.Set("control_status", control.ControlStatus)
	d.Set("control_status_updated_at", control.ControlStatusUpdatedAt.Format(time.RFC3339))
	d.Set("description", control.Description)
	d.Set("disabled_reason", control.DisabledReason)
	d.Set("related_requirements", aws.StringValueSlice(control.RelatedRequirements))
	d.Set("remediation_url", control.RemediationUrl)
	d.Set("severity_rating", control.SeverityRating)
	d.Set("standards_control_arn", control.StandardsControlArn)
	d.Set("title", control.Title)

	return nil
}

func resourceAwsSecurityHubStandardsControlPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	d.SetId(d.Get("standards_control_arn").(string))

	input := &securityhub.UpdateStandardsControlInput{
		ControlStatus:       aws.String(d.Get("control_status").(string)),
		DisabledReason:      aws.String(d.Get("disabled_reason").(string)),
		StandardsControlArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Updating Security Hub Standards Control: %s", input)
	_, err := conn.UpdateStandardsControlWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("error updating Security Hub Standards Control (%s): %s", d.Id(), err)
	}

	return resourceStandardsControlRead(ctx, d, meta)
}

func resourceStandardsControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[WARN] Cannot delete Security Hub Standards Control. Terraform will remove this resource from the state.")
	return nil
}
