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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
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
	var diags diag.Diagnostics

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
		return sdkdiag.AppendErrorf(diags, "creating Detective Member (%s): %s", id, err)
	}

	if _, err := waitMemberInvited(ctx, conn, graphARN, accountID); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Detective Member (%s) invited: %s", d.Id(), err)
	}

	d.SetId(id)

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graphARN, accountID, err := MemberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	member, err := FindMemberByGraphByTwoPartKey(ctx, conn, graphARN, accountID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Member (%s): %s", d.Id(), err)
	}

	d.Set("account_id", member.AccountId)
	d.Set("administrator_id", member.AdministratorId)
	d.Set("disabled_reason", member.DisabledReason)
	d.Set("email_address", member.EmailAddress)
	d.Set("graph_arn", member.GraphArn)
	d.Set("invited_time", aws.TimeValue(member.InvitedTime).Format(time.RFC3339))
	d.Set("status", member.Status)
	d.Set("updated_time", aws.TimeValue(member.UpdatedTime).Format(time.RFC3339))
	d.Set("volume_usage_in_bytes", member.VolumeUsageInBytes)

	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graphARN, accountID, err := MemberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Detective Member: %s", d.Id())
	_, err = conn.DeleteMembersWithContext(ctx, &detective.DeleteMembersInput{
		AccountIds: aws.StringSlice([]string{accountID}),
		GraphArn:   aws.String(graphARN),
	})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Detective Member (%s): %s", d.Id(), err)
	}

	return diags
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

func FindMemberByGraphByTwoPartKey(ctx context.Context, conn *detective.Detective, graphARN, accountID string) (*detective.MemberDetail, error) {
	input := &detective.ListMembersInput{
		GraphArn: aws.String(graphARN),
	}

	return findMember(ctx, conn, input, func(v *detective.MemberDetail) bool {
		return aws.StringValue(v.AccountId) == accountID
	})
}

func findMember(ctx context.Context, conn *detective.Detective, input *detective.ListMembersInput, filter tfslices.Predicate[*detective.MemberDetail]) (*detective.MemberDetail, error) {
	output, err := findMembers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findMembers(ctx context.Context, conn *detective.Detective, input *detective.ListMembersInput, filter tfslices.Predicate[*detective.MemberDetail]) ([]*detective.MemberDetail, error) {
	var output []*detective.MemberDetail

	err := conn.ListMembersPagesWithContext(ctx, input, func(page *detective.ListMembersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MemberDetails {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusMember(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMemberByGraphByTwoPartKey(ctx, conn, graphARN, adminAccountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitMemberInvited(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID string) (*detective.MemberDetail, error) {
	const (
		timeout = 4 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{detective.MemberStatusVerificationInProgress},
		Target:  []string{detective.MemberStatusInvited},
		Refresh: statusMember(ctx, conn, graphARN, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*detective.MemberDetail); ok {
		return output, err
	}

	return nil, err
}
