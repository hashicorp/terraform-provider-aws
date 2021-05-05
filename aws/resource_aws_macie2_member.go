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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsMacie2Member() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2MemberCreate,
		ReadWithoutTimeout:   resourceMacie2MemberRead,
		UpdateWithoutTimeout: resourceMacie2MemberUpdate,
		DeleteWithoutTimeout: resourceMacie2MemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"relationship_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"administrator_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invited_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(macie2.MacieStatus_Values(), false),
			},
		},
	}
}

func resourceMacie2MemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	accountId := d.Get("account_id").(string)
	input := &macie2.CreateMemberInput{
		Account: &macie2.AccountDetail{
			AccountId: aws.String(accountId),
			Email:     aws.String(d.Get("email").(string)),
		},
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().Macie2Tags()
	}

	var err error
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateMemberWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateMemberWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie Member: %w", err))
	}

	d.SetId(accountId)

	return resourceMacie2MemberRead(ctx, d, meta)
}

func resourceMacie2MemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetMemberInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetMemberWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "member accounts are associated with your account") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "account is not associated with your account") {
			log.Printf("[WARN] Macie Member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Macie Member (%s): %w", d.Id(), err))
	}

	d.Set("account_id", resp.AccountId)
	d.Set("email", resp.Email)
	d.Set("relationship_status", resp.RelationshipStatus)
	d.Set("administrator_account_id", resp.AdministratorAccountId)
	d.Set("invited_at", aws.TimeValue(resp.InvitedAt).Format(time.RFC3339))
	d.Set("updated_at", aws.TimeValue(resp.UpdatedAt).Format(time.RFC3339))
	d.Set("arn", resp.Arn)
	tags := keyvaluetags.Macie2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie Member (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie Member (%s): %w", "tags_all", d.Id(), err))
	}

	return nil
}

func resourceMacie2MemberUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateMemberSessionInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChange("status") {
		input.Status = aws.String(d.Get("status").(string))
	}

	_, err := conn.UpdateMemberSessionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Macie Member (%s): %w", d.Id(), err))
	}

	return resourceMacie2MemberRead(ctx, d, meta)
}

func resourceMacie2MemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteMemberInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteMemberWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeConflictException, "member accounts are associated with your account") ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeValidationException, "account is not associated with your account") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie Member (%s): %w", d.Id(), err))
	}
	return nil
}
