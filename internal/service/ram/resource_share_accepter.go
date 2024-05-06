// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
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
			"resources": {
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
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	shareARN := d.Get("share_arn").(string)
	maybeInvitation, err := findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, shareARN, ram.ResourceShareInvitationStatusPending)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading pending RAM Resource Share (%s) invitation: %s", shareARN, err)
	}

	var invitationExists bool
	var invitationARN string
	if maybeInvitation.IsSome() {
		invitationExists = true
		invitationARN = aws.StringValue(maybeInvitation.MustUnwrap().ResourceShareInvitationArn)
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

	output, err := conn.AcceptResourceShareInvitationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting RAM Resource Share (%s) invitation (%s): %s", shareARN, invitationARN, err)
	}

	d.SetId(shareARN)

	invitationARN = aws.StringValue(output.ResourceShareInvitation.ResourceShareInvitationArn)
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
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	maybeInvitation, err := findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, d.Id(), ram.ResourceShareInvitationStatusAccepted)

	if err != nil && !tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceShareInvitationArnNotFoundException) {
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
		MaxResults:        aws.Int64(500),
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{d.Id()}),
	}
	resources, err := findResources(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM Resource Share (%s) resources: %s", d.Id(), err)
	}

	resourceARNs := tfslices.ApplyToAll(resources, func(r *ram.Resource) string {
		return aws.StringValue(r.Arn)
	})
	d.Set("resources", resourceARNs)

	return diags
}

func resourceResourceShareAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn(ctx)

	receiverAccountID := d.Get("receiver_account_id").(string)
	if receiverAccountID == "" {
		return sdkdiag.AppendErrorf(diags, "The receiver account ID is required to leave a resource share")
	}

	input := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(id.UniqueId()),
		Principals:       []*string{aws.String(receiverAccountID)},
		ResourceShareArn: aws.String(d.Id()),
	}

	_, err := conn.DisassociateResourceShareWithContext(ctx, input)

	switch {
	case tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException):
		return diags

	case tfawserr.ErrCodeEquals(err, ram.ErrCodeOperationNotPermittedException):
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

func findResourceShareOwnerOtherAccountsByARN(ctx context.Context, conn *ram.RAM, arn string) (*ram.ResourceShare, error) {
	input := &ram.GetResourceSharesInput{
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{arn}),
	}

	output, err := findResourceShare(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Deleted resource share OK.

	return output, nil
}

func findResources(ctx context.Context, conn *ram.RAM, input *ram.ListResourcesInput) ([]*ram.Resource, error) {
	var output []*ram.Resource

	err := conn.ListResourcesPagesWithContext(ctx, input, func(page *ram.ListResourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Resources {
			if v != nil {
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

func findMaybeResourceShareInvitationByResourceShareARNAndStatus(ctx context.Context, conn *ram.RAM, resourceShareARN, status string) (option.Option[*ram.ResourceShareInvitation], error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	return findMaybeResourceShareInvitationRetry(ctx, conn, input, func(v *ram.ResourceShareInvitation) bool {
		return aws.StringValue(v.Status) == status
	})
}

func findMaybeResourceShareInvitationByARN(ctx context.Context, conn *ram.RAM, arn string) (option.Option[*ram.ResourceShareInvitation], error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareInvitationArns: aws.StringSlice([]string{arn}),
	}

	return findMaybeResourceShareInvitationRetry(ctx, conn, input, tfslices.PredicateTrue[*ram.ResourceShareInvitation]())
}

func findMaybeResourceShareInvitationRetry(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*ram.ResourceShareInvitation]) (option.Option[*ram.ResourceShareInvitation], error) {
	// Retry for RAM resource share invitation eventual consistency.
	errNotFound := errors.New("not found")
	var output option.Option[*ram.ResourceShareInvitation]
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
		output, err = option.None[*ram.ResourceShareInvitation](), nil
	}

	return output, err
}

func findMaybeResourceShareInvitation(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*ram.ResourceShareInvitation]) (option.Option[*ram.ResourceShareInvitation], error) {
	output, err := findResourceShareInvitations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertMaybeSinglePtrResult(output)
}

func findResourceShareInvitations(ctx context.Context, conn *ram.RAM, input *ram.GetResourceShareInvitationsInput, filter tfslices.Predicate[*ram.ResourceShareInvitation]) ([]*ram.ResourceShareInvitation, error) {
	var output []*ram.ResourceShareInvitation

	err := conn.GetResourceShareInvitationsPagesWithContext(ctx, input, func(page *ram.GetResourceShareInvitationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourceShareInvitations {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceShareInvitationArnNotFoundException, ram.ErrCodeUnknownResourceException) {
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

func statusResourceShareInvitation(ctx context.Context, conn *ram.RAM, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		maybeInvitation, err := findMaybeResourceShareInvitationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) || maybeInvitation.IsNone() {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		output := maybeInvitation.MustUnwrap()

		return output, aws.StringValue(output.Status), nil
	}
}

func waitResourceShareInvitationAccepted(ctx context.Context, conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShareInvitation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareInvitationStatusPending},
		Target:  []string{ram.ResourceShareInvitationStatusAccepted},
		Refresh: statusResourceShareInvitation(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ram.ResourceShareInvitation); ok {
		return v, err
	}

	return nil, err
}

func waitResourceShareOwnedBySelfDisassociated(ctx context.Context, conn *ram.RAM, arn string, timeout time.Duration) (*ram.ResourceShare, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated},
		Target:  []string{},
		Refresh: statusResourceShareOwnerSelf(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ram.ResourceShare); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
