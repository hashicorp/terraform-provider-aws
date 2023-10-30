// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_detective_member")
func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberCreate,
		ReadWithoutTimeout:   resourceMemberRead,
		DeleteWithoutTimeout: resourceMemberDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"administrator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"graph_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"invited_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_usage_in_bytes": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	accountID := d.Get("account_id").(string)
	graphARN := d.Get("graph_arn").(string)
	id := memberCreateResourceID(graphARN, accountID)
	input := &detective.CreateMembersInput{
		Accounts: []*detective.Account{{
			AccountId:    aws.String(accountID),
			EmailAddress: aws.String(d.Get("email_address").(string)),
		}},
		GraphArn: aws.String(graphARN),
	}

	if v := d.Get("disable_email_notification").(bool); v {
		input.DisableEmailNotification = aws.Bool(v)
	}

	if v, ok := d.GetOk("message"); ok {
		input.Message = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateMembersWithContext(ctx, input)
	}, detective.ErrCodeInternalServerException)

	if err != nil {
		return diag.Errorf("creating Detective Member (%s): %s", id, err)
	}

	if _, err = MemberStatusUpdated(ctx, conn, graphARN, accountID, detective.MemberStatusInvited); err != nil {
		return diag.Errorf("waiting for Detective Member (%s) to be invited: %s", d.Id(), err)
	}

	d.SetId(id)

	return resourceMemberRead(ctx, d, meta)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graphArn, accountId, err := MemberParseResourceID(d.Id())
	if err != nil {
		return diag.Errorf("decoding ID Detective Member (%s): %s", d.Id(), err)
	}

	resp, err := FindMemberByGraphARNAndAccountID(ctx, conn, graphArn, accountId)

	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) ||
			tfresource.NotFound(err) {
			log.Printf("[WARN] Detective Member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("reading Detective Member (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && resp == nil {
		log.Printf("[WARN] Detective Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", resp.AccountId)
	d.Set("administrator_id", resp.AdministratorId)
	d.Set("disabled_reason", resp.DisabledReason)
	d.Set("email_address", resp.EmailAddress)
	d.Set("graph_arn", resp.GraphArn)
	d.Set("invited_time", aws.TimeValue(resp.InvitedTime).Format(time.RFC3339))
	d.Set("status", resp.Status)
	d.Set("updated_time", aws.TimeValue(resp.UpdatedTime).Format(time.RFC3339))
	d.Set("volume_usage_in_bytes", resp.VolumeUsageInBytes)

	return nil
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graphARN, accountID, err := MemberParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Deleting Detective Member: %s", d.Id())
	_, err = conn.DeleteMembersWithContext(ctx, &detective.DeleteMembersInput{
		AccountIds: aws.StringSlice([]string{accountID}),
		GraphArn:   aws.String(graphARN),
	})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Detective Member (%s): %s", d.Id(), err)
	}

	return nil
}

const memberResourceIDSeparator = "/"

func memberCreateResourceID(graphARN, accountID string) string {
	parts := []string{graphARN, accountID}
	id := strings.Join(parts, memberResourceIDSeparator)

	return id
}

func MemberParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, memberResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected graph_arn%[2]saccount_id", id, memberResourceIDSeparator)
}
