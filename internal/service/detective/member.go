// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			names.AttrAccountID: {
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
			names.AttrMessage: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrStatus: {
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

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	graphARN := d.Get("graph_arn").(string)
	id := memberCreateResourceID(graphARN, accountID)
	input := &detective.CreateMembersInput{
		Accounts: []awstypes.Account{{
			AccountId:    aws.String(accountID),
			EmailAddress: aws.String(d.Get("email_address").(string)),
		}},
		GraphArn: aws.String(graphARN),
	}

	if v := d.Get("disable_email_notification").(bool); v {
		input.DisableEmailNotification = v
	}

	if v, ok := d.GetOk(names.AttrMessage); ok {
		input.Message = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenIsA[*awstypes.InternalServerException](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateMembers(ctx, input)
	})

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

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

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

	d.Set(names.AttrAccountID, member.AccountId)
	d.Set("administrator_id", member.AdministratorId)
	d.Set("disabled_reason", member.DisabledReason)
	d.Set("email_address", member.EmailAddress)
	d.Set("graph_arn", member.GraphArn)
	d.Set("invited_time", aws.ToTime(member.InvitedTime).Format(time.RFC3339))
	d.Set(names.AttrStatus, member.Status)
	d.Set("updated_time", aws.ToTime(member.UpdatedTime).Format(time.RFC3339))
	d.Set("volume_usage_in_bytes", member.VolumeUsageInBytes)

	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	graphARN, accountID, err := MemberParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Detective Member: %s", d.Id())
	_, err = conn.DeleteMembers(ctx, &detective.DeleteMembersInput{
		AccountIds: []string{accountID},
		GraphArn:   aws.String(graphARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindMemberByGraphByTwoPartKey(ctx context.Context, conn *detective.Client, graphARN, accountID string) (*awstypes.MemberDetail, error) {
	input := &detective.ListMembersInput{
		GraphArn: aws.String(graphARN),
	}

	return findMember(ctx, conn, input, func(v awstypes.MemberDetail) bool {
		return aws.ToString(v.AccountId) == accountID
	})
}

func findMember(ctx context.Context, conn *detective.Client, input *detective.ListMembersInput, filter tfslices.Predicate[awstypes.MemberDetail]) (*awstypes.MemberDetail, error) {
	output, err := findMembers(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMembers(ctx context.Context, conn *detective.Client, input *detective.ListMembersInput, filter tfslices.Predicate[awstypes.MemberDetail]) ([]awstypes.MemberDetail, error) {
	var output []awstypes.MemberDetail

	pages := detective.NewListMembersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.MemberDetails {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusMember(ctx context.Context, conn *detective.Client, graphARN, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMemberByGraphByTwoPartKey(ctx, conn, graphARN, adminAccountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitMemberInvited(ctx context.Context, conn *detective.Client, graphARN, adminAccountID string) (*awstypes.MemberDetail, error) {
	const (
		timeout = 4 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MemberStatusVerificationInProgress),
		Target:  enum.Slice(awstypes.MemberStatusInvited),
		Refresh: statusMember(ctx, conn, graphARN, adminAccountID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.MemberDetail); ok {
		return output, err
	}

	return nil, err
}
