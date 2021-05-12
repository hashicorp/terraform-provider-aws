package aws

import (
	"context"
	"fmt"
	"log"
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"disable_email_notification": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"invited_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2InvitationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	accountID := d.Get("account_id").(string)
	input := &macie2.CreateInvitationsInput{
		AccountIds: []*string{aws.String(accountID)},
	}

	if v, ok := d.GetOk("disable_email_notification"); ok {
		input.DisableEmailNotification = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("message"); ok {
		input.Message = aws.String(v.(string))
	}

	var err error
	var output *macie2.CreateInvitationsOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		output, err = conn.CreateInvitationsWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if len(output.UnprocessedAccounts) > 0 {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateInvitationsWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie Invitation: %w", err))
	}

	if len(output.UnprocessedAccounts) != 0 {
		return diag.FromErr(fmt.Errorf("error creating Macie Invitation: %w", fmt.Errorf("%s: %s", aws.StringValue(output.UnprocessedAccounts[0].ErrorCode), aws.StringValue(output.UnprocessedAccounts[0].ErrorMessage))))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return resourceMacie2InvitationRead(ctx, d, meta)
}

func resourceMacie2InvitationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	var err error

	input := &macie2.ListMembersInput{
		OnlyAssociated: aws.String("false"),
	}
	var result *macie2.Member
	err = conn.ListMembersPages(input, func(page *macie2.ListMembersOutput, lastPage bool) bool {
		for _, member := range page.Members {
			if aws.StringValue(member.AdministratorAccountId) == d.Id() {
				result = member
				return false
			}
		}
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") ||
		result == nil {
		log.Printf("[WARN] Macie Invitation (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Macie Invitation (%s): %w", d.Id(), err))
	}

	if aws.StringValue(result.RelationshipStatus) == macie2.RelationshipStatusRemoved ||
		aws.StringValue(result.RelationshipStatus) == macie2.RelationshipStatusResigned {
		log.Printf("[WARN] Macie InvitationAccepter (%s) %s, removing from state", d.Id(), aws.StringValue(result.RelationshipStatus))
		d.SetId("")
		return nil
	}

	if aws.StringValue(result.RelationshipStatus) == macie2.RelationshipStatusEmailVerificationFailed {
		log.Printf("[WARN] Macie InvitationAccepter (%s) %s", d.Id(), aws.StringValue(result.RelationshipStatus))
	}

	d.Set("invited_at", aws.TimeValue(result.InvitedAt).Format(time.RFC3339))
	d.Set("account_id", result.AccountId)

	return nil
}

func resourceMacie2InvitationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteInvitationsInput{
		AccountIds: []*string{aws.String(d.Id())},
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

	if len(output.UnprocessedAccounts) != 0 {
		return diag.FromErr(fmt.Errorf("error deleting Macie Invitation: %w", fmt.Errorf("%s: %s", aws.StringValue(output.UnprocessedAccounts[0].ErrorCode), aws.StringValue(output.UnprocessedAccounts[0].ErrorMessage))))
	}

	return nil
}
