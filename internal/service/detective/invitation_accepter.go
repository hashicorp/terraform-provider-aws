// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/detective"
	awstypes "github.com/aws/aws-sdk-go-v2/service/detective/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_detective_invitation_accepter")
func ResourceInvitationAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInvitationAccepterCreate,
		ReadWithoutTimeout:   resourceInvitationAccepterRead,
		DeleteWithoutTimeout: resourceInvitationAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"graph_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceInvitationAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	graphARN := d.Get("graph_arn").(string)
	input := &detective.AcceptInvitationInput{
		GraphArn: aws.String(graphARN),
	}

	_, err := conn.AcceptInvitation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Detective Invitation (%s): %s", graphARN, err)
	}

	d.SetId(graphARN)

	return append(diags, resourceInvitationAccepterRead(ctx, d, meta)...)
}

func resourceInvitationAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	member, err := FindInvitationByGraphARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Detective Invitation Accepter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Detective Invitation Accepter (%s): %s", d.Id(), err)
	}

	d.Set("graph_arn", member.GraphArn)

	return diags
}

func resourceInvitationAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveClient(ctx)

	log.Printf("[DEBUG] Deleting Detective Invitation Accepter: %s", d.Id())
	_, err := conn.DisassociateMembership(ctx, &detective.DisassociateMembershipInput{
		GraphArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Detective InvitationAccepter (%s): %s", d.Id(), err)
	}

	return diags
}

func FindInvitationByGraphARN(ctx context.Context, conn *detective.Client, graphARN string) (*awstypes.MemberDetail, error) {
	input := &detective.ListInvitationsInput{}

	return findInvitation(ctx, conn, input, func(v awstypes.MemberDetail) bool {
		return aws.ToString(v.GraphArn) == graphARN
	})
}

func findInvitation(ctx context.Context, conn *detective.Client, input *detective.ListInvitationsInput, filter tfslices.Predicate[awstypes.MemberDetail]) (*awstypes.MemberDetail, error) {
	output, err := findInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInvitations(ctx context.Context, conn *detective.Client, input *detective.ListInvitationsInput, filter tfslices.Predicate[awstypes.MemberDetail]) ([]awstypes.MemberDetail, error) {
	var output []awstypes.MemberDetail

	pages := detective.NewListInvitationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Invitations {
			if filter(v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
