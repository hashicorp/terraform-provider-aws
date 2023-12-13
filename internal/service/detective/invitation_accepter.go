// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	graphARN := d.Get("graph_arn").(string)
	input := &detective.AcceptInvitationInput{
		GraphArn: aws.String(graphARN),
	}

	_, err := conn.AcceptInvitationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Detective Invitation (%s): %s", graphARN, err)
	}

	d.SetId(graphARN)

	return append(diags, resourceInvitationAccepterRead(ctx, d, meta)...)
}

func resourceInvitationAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

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

	conn := meta.(*conns.AWSClient).DetectiveConn(ctx)

	log.Printf("[DEBUG] Deleting Detective Invitation Accepter: %s", d.Id())
	_, err := conn.DisassociateMembershipWithContext(ctx, &detective.DisassociateMembershipInput{
		GraphArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Detective InvitationAccepter (%s): %s", d.Id(), err)
	}

	return diags
}

func FindInvitationByGraphARN(ctx context.Context, conn *detective.Detective, graphARN string) (*detective.MemberDetail, error) {
	input := &detective.ListInvitationsInput{}

	return findInvitation(ctx, conn, input, func(v *detective.MemberDetail) bool {
		return aws.StringValue(v.GraphArn) == graphARN
	})
}

func findInvitation(ctx context.Context, conn *detective.Detective, input *detective.ListInvitationsInput, filter tfslices.Predicate[*detective.MemberDetail]) (*detective.MemberDetail, error) {
	output, err := findInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findInvitations(ctx context.Context, conn *detective.Detective, input *detective.ListInvitationsInput, filter tfslices.Predicate[*detective.MemberDetail]) ([]*detective.MemberDetail, error) {
	var output []*detective.MemberDetail

	err := conn.ListInvitationsPagesWithContext(ctx, input, func(page *detective.ListInvitationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Invitations {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
