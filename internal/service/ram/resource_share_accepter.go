// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ram"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ram/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	resourceShareInvitationPropagationTimeout = 2 * time.Minute
)

// @SDKResource("aws_ram_resource_share_accepter", name="Resource Share Accepter")
func resourceResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceShareAccepterCreate,
		ReadWithoutTimeout:   resourceResourceShareAccepterRead,
		DeleteWithoutTimeout: resourceResourceShareAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"invitation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"receiver_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResources: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sender_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"share_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceResourceShareAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	shareARN := d.Get("share_arn").(string)
	maybeInvitation, err := findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, shareARN, string(awstypes.ResourceShareInvitationStatusPending))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading pending RAM Resource Share (%s) invitation: %s", shareARN, err)
	}

	var invitationExists bool
	var invitationARN string
	if maybeInvitation.IsSome() {
		invitationExists = true
		invitationARN = aws.ToString(maybeInvitation.MustUnwrap().ResourceShareInvitationArn)
	}

	if !invitationExists || invitationARN == "" {
		return sdkdiag.AppendErrorf(diags, "No pending RAM Resource Share (%s) invitation found\n\n"+
			"NOTE: If both AWS accounts are in the same AWS Organization and RAM Sharing with AWS Organizations is enabled, this resource is not necessary",
			shareARN)
	}

	input := &ram.AcceptResourceShareInvitationInput{
		ClientToken:                aws.String(id.UniqueId()),
		ResourceShareInvitationArn: aws.String(invitationARN),
	}

	output, err := conn.AcceptResourceShareInvitation(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting RAM Resource Share (%s) invitation (%s): %s", shareARN, invitationARN, err)
	}

	d.SetId(shareARN)

	invitationARN = aws.ToString(output.ResourceShareInvitation.ResourceShareInvitationArn)
	if _, err := waitResourceShareInvitationAccepted(ctx, conn, invitationARN, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) invitation (%s) accept: %s", shareARN, invitationARN, err)
	}

	_, err = tfresource.RetryWhenNotFound(ctx, resourceSharePropagationTimeout, func() (interface{}, error) {
		return findResourceShareOwnerOtherAccountsByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) create: %s", shareARN, err)
	}

	return append(diags, resourceResourceShareAccepterRead(ctx, d, meta)...)
}

func resourceResourceShareAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID := meta.(*conns.AWSClient).AccountID
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	maybeInvitation, err := findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, d.Id(), string(awstypes.ResourceShareInvitationStatusAccepted))

	if err != nil && !errs.IsA[*awstypes.ResourceShareInvitationArnNotFoundException](err) {
		return sdkdiag.AppendErrorf(diags, "reading accepted RAM Resource Share (%s) invitation: %s", d.Id(), err)
	}

	if maybeInvitation.IsSome() {
		invitation := maybeInvitation.MustUnwrap()
		d.Set("invitation_arn", invitation.ResourceShareInvitationArn)
		d.Set("receiver_account_id", invitation.ReceiverAccountId)
	} else {
		d.Set("receiver_account_id", accountID)
	}

	resourceShare, err := findResourceShareOwnerOtherAccountsByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] No RAM Resource Share %s found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s): %s", d.Id(), err)
	}

	d.Set("sender_account_id", resourceShare.OwningAccountId)
	d.Set("share_arn", resourceShare.ResourceShareArn)
	d.Set("share_id", resourceResourceShareIDFromARN(d.Id()))
	d.Set("share_name", resourceShare.Name)
	d.Set(names.AttrStatus, resourceShare.Status)

	input := &ram.ListResourcesInput{
		MaxResults:        aws.Int32(500),
		ResourceOwner:     awstypes.ResourceOwnerOtherAccounts,
		ResourceShareArns: []string{d.Id()},
	}
	resources, err := findResources(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) resources: %s", d.Id(), err)
	}

	resourceARNs := tfslices.ApplyToAll(resources, func(r awstypes.Resource) string {
		return aws.ToString(r.Arn)
	})
	d.Set(names.AttrResources, resourceARNs)

	return diags
}

func resourceResourceShareAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMClient(ctx)

	receiverAccountID := d.Get("receiver_account_id").(string)
	if receiverAccountID == "" {
		return sdkdiag.AppendErrorf(diags, "The receiver account ID is required to leave a resource share")
	}

	input := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(id.UniqueId()),
		Principals:       []string{receiverAccountID},
		ResourceShareArn: aws.String(d.Id()),
	}

	_, err := conn.DisassociateResourceShare(ctx, input)

	switch {
	case errs.IsA[*awstypes.UnknownResourceException](err):
		return diags

	case errs.IsA[*awstypes.OperationNotPermittedException](err):
		log.Printf("[WARN] Resource share could not be disassociated, but continuing: %s", err)

	case err != nil:
		return sdkdiag.AppendErrorf(diags, "leaving RAM resource share: %s", err)
	}

	if _, err := waitResourceShareOwnedBySelfDisassociated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM Resource Share (%s) disassociate: %s", d.Id(), err)
	}

	return diags
}

func resourceResourceShareIDFromARN(arn string) string {
	return strings.Replace(arn[strings.LastIndex(arn, ":")+1:], "resource-share/", "rs-", -1)
}

func findResourceShareOwnerOtherAccountsByARN(ctx context.Context, conn *ram.Client, arn string) (*awstypes.ResourceShare, error) {
	input := &ram.GetResourceSharesInput{
		ResourceOwner:     awstypes.ResourceOwnerOtherAccounts,
		ResourceShareArns: []string{arn},
	}

	output, err := findResourceShare(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Deleted resource share OK.

	return output, nil
}

func findResources(ctx context.Context, conn *ram.Client, input *ram.ListResourcesInput) ([]awstypes.Resource, error) {
	var output []awstypes.Resource

	pages := ram.NewListResourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.Resources...)
	}

	return output, nil
}

func findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx context.Context, conn *ram.Client, resourceShareARN, status string) (option.Option[awstypes.ResourceShareInvitation], error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareArns: []string{resourceShareARN},
	}

	return findMaybeResourceShareInvitationRetry(ctx, conn, input, func(v *awstypes.ResourceShareInvitation) bool {
		return string(v.Status) == status
	})
}

func findMaybeResourceShareInvitationByARN(ctx context.Context, conn *ram.Client, arn string) (option.Option[awstypes.ResourceShareInvitation], error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareInvitationArns: []string{arn},
	}

	return findMaybeResourceShareInvitationRetry(ctx, conn, input, tfslices.PredicateTrue[*awstypes.ResourceShareInvitation]())
}

func findMaybeResourceShareInvitationRetry(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*awstypes.ResourceShareInvitation]) (option.Option[awstypes.ResourceShareInvitation], error) {
	// Retry for RAM resource share invitation eventual consistency.
	errNotFound := errors.New("not found")
	var output option.Option[awstypes.ResourceShareInvitation]
	err := tfresource.Retry(ctx, resourceShareInvitationPropagationTimeout, func() *retry.RetryError {
		var err error
		output, err = findMaybeResourceShareInvitation(ctx, conn, input, filter)

		if err != nil {
			return retry.NonRetryableError(err)
		}

		if output.IsNone() {
			return retry.RetryableError(errNotFound)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = findMaybeResourceShareInvitation(ctx, conn, input, filter)
	}

	if errors.Is(err, errNotFound) {
		output, err = option.None[awstypes.ResourceShareInvitation](), nil
	}

	return output, err
}

func findMaybeResourceShareInvitation(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*awstypes.ResourceShareInvitation]) (option.Option[awstypes.ResourceShareInvitation], error) {
	output, err := findResourceShareInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertMaybeSingleValueResult(output)
}

func findResourceShareInvitations(ctx context.Context, conn *ram.Client, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*awstypes.ResourceShareInvitation]) ([]awstypes.ResourceShareInvitation, error) {
	var output []awstypes.ResourceShareInvitation

	pages := ram.NewGetResourceShareInvitationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceShareInvitationArnNotFoundException](err) || errs.IsA[*awstypes.UnknownResourceException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ResourceShareInvitations {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusResourceShareInvitation(ctx context.Context, conn *ram.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		maybeInvitation, err := findMaybeResourceShareInvitationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) || maybeInvitation.IsNone() {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		output := maybeInvitation.MustUnwrap()

		return output, string(output.Status), nil
	}
}

func waitResourceShareInvitationAccepted(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceShareInvitation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareInvitationStatusPending),
		Target:  enum.Slice(awstypes.ResourceShareInvitationStatusAccepted),
		Refresh: statusResourceShareInvitation(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.ResourceShareInvitation); ok {
		return v, err
	}

	return nil, err
}

func waitResourceShareOwnedBySelfDisassociated(ctx context.Context, conn *ram.Client, arn string, timeout time.Duration) (*awstypes.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceShareAssociationStatusAssociated),
		Target:  []string{},
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
