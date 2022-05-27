package ram

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	FindInvitationTimeout    = 2 * time.Minute
	FindResourceShareTimeout = 1 * time.Minute
)

// FindResourceShareOwnerOtherAccountsByARN returns the resource share owned by other accounts corresponding to the specified ARN.
// Returns nil if no configuration is found.
func FindResourceShareOwnerOtherAccountsByARN(conn *ram.RAM, arn string) (*ram.ResourceShare, error) {
	listResourceSharesInput := &ram.GetResourceSharesInput{
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{arn}),
	}

	return resourceShare(conn, listResourceSharesInput)
}

// FindResourceShareOwnerSelfByARN returns the resource share owned by own account corresponding to the specified ARN.
// Returns nil if no configuration is found.
func FindResourceShareOwnerSelfByARN(conn *ram.RAM, arn string) (*ram.ResourceShare, error) {
	listResourceSharesInput := &ram.GetResourceSharesInput{
		ResourceOwner:     aws.String(ram.ResourceOwnerSelf),
		ResourceShareArns: aws.StringSlice([]string{arn}),
	}

	return resourceShare(conn, listResourceSharesInput)
}

// FindResourceShareInvitationByResourceShareARNAndStatus returns the resource share invitation corresponding to the specified resource share ARN.
// Returns nil if no configuration is found.
func FindResourceShareInvitationByResourceShareARNAndStatus(conn *ram.RAM, resourceShareArn, status string) (*ram.ResourceShareInvitation, error) {
	var invitation *ram.ResourceShareInvitation

	// Retry for Ram resource share invitation eventual consistency
	err := resource.Retry(FindInvitationTimeout, func() *resource.RetryError {
		i, err := resourceShareInvitationByResourceShareARNAndStatus(conn, resourceShareArn, status)
		invitation = i

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if invitation == nil {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		invitation, err = resourceShareInvitationByResourceShareARNAndStatus(conn, resourceShareArn, status)
	}

	if invitation == nil {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

// FindResourceShareInvitationByARN returns the resource share invitation corresponding to the specified ARN.
// Returns nil if no configuration is found.
func FindResourceShareInvitationByARN(conn *ram.RAM, arn string) (*ram.ResourceShareInvitation, error) {
	var invitation *ram.ResourceShareInvitation

	// Retry for Ram resource share invitation eventual consistency
	err := resource.Retry(FindInvitationTimeout, func() *resource.RetryError {
		i, err := resourceShareInvitationByARN(conn, arn)
		invitation = i

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if invitation == nil {
			resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		invitation, err = resourceShareInvitationByARN(conn, arn)
	}

	if invitation == nil {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func resourceShare(conn *ram.RAM, input *ram.GetResourceSharesInput) (*ram.ResourceShare, error) {
	var shares *ram.GetResourceSharesOutput

	// Retry for Ram resource share eventual consistency
	err := resource.Retry(FindResourceShareTimeout, func() *resource.RetryError {
		ss, err := conn.GetResourceShares(input)
		shares = ss

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if len(shares.ResourceShares) == 0 {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		shares, err = conn.GetResourceShares(input)
	}

	if err != nil {
		return nil, err
	}

	if shares == nil || len(shares.ResourceShares) == 0 {
		return nil, nil
	}

	return shares.ResourceShares[0], nil
}

func resourceShareInvitationByResourceShareARNAndStatus(conn *ram.RAM, resourceShareArn, status string) (*ram.ResourceShareInvitation, error) {
	var invitation *ram.ResourceShareInvitation

	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareArns: []*string{aws.String(resourceShareArn)},
	}

	err := conn.GetResourceShareInvitationsPages(input, func(page *ram.GetResourceShareInvitationsOutput, lastPage bool) bool {
		for _, rsi := range page.ResourceShareInvitations {
			if aws.StringValue(rsi.Status) == status {
				invitation = rsi
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func resourceShareInvitationByARN(conn *ram.RAM, arn string) (*ram.ResourceShareInvitation, error) {
	input := &ram.GetResourceShareInvitationsInput{
		ResourceShareInvitationArns: []*string{aws.String(arn)},
	}

	output, err := conn.GetResourceShareInvitations(input)

	if err != nil {
		return nil, err
	}

	if len(output.ResourceShareInvitations) == 0 {
		return nil, nil
	}

	return output.ResourceShareInvitations[0], nil
}

func FindResourceSharePrincipalAssociationByShareARNPrincipal(conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
		Principal:         aws.String(principal),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	output, err := conn.GetResourceShareAssociations(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ResourceShareAssociations) == 0 || output.ResourceShareAssociations[0] == nil {
		return nil, nil
	}

	return output.ResourceShareAssociations[0], nil
}
