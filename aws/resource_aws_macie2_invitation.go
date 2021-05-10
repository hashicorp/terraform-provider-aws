package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsMacie2Invitation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2InvitationCreate,
		ReadWithoutTimeout:   resourceMacie2InvitationRead,
		DeleteWithoutTimeout: resourceMacie2InvitationDelete,
		Schema: map[string]*schema.Schema{
			"disable_email_notification": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"account_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
		},
	}
}

func resourceMacie2InvitationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.CreateInvitationsInput{
		AccountIds: expandStringList(d.Get("account_ids").([]interface{})),
	}

	if v, ok := d.GetOk("disable_email_notification"); ok {
		input.DisableEmailNotification = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("message"); ok {
		input.Message = aws.String(v.(string))
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateInvitationsWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateInvitationsWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie Invitation: %w", err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceMacie2InvitationRead(ctx, d, meta)
}

func resourceMacie2InvitationRead(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceMacie2InvitationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteInvitationsInput{
		AccountIds: expandStringList(d.Get("account_ids").([]interface{})),
	}

	output, err := conn.DeleteInvitationsWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie Invitation (%s): %w", d.Id(), err))
	}

	for _, invitation := range output.UnprocessedAccounts {
		if strings.Contains(aws.StringValue(invitation.ErrorMessage), "either because no such invitation exists") {
			return nil
		}
	}
	return nil
}
