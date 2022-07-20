package detective

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindGraphByARN(conn *detective.Detective, ctx context.Context, arn string) (*detective.Graph, error) {
	input := &detective.ListGraphsInput{}
	var result *detective.Graph

	err := conn.ListGraphsPagesWithContext(ctx, input, func(page *detective.ListGraphsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, graph := range page.GraphList {
			if graph == nil {
				continue
			}

			if aws.StringValue(graph.Arn) == arn {
				result = graph
				return false
			}
		}
		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No detective graph with arn %q", arn),
			LastRequest: input,
		}
	}

	return result, nil
}

func FindInvitationByGraphARN(ctx context.Context, conn *detective.Detective, graphARN string) (*string, error) {
	input := &detective.ListInvitationsInput{}

	var result *string

	err := conn.ListInvitationsPagesWithContext(ctx, input, func(page *detective.ListInvitationsOutput, lastPage bool) bool {
		for _, invitation := range page.Invitations {
			if aws.StringValue(invitation.GraphArn) == graphARN {
				result = invitation.GraphArn
				return false
			}
		}
		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No member found with arn %q ", graphARN),
			LastRequest: input,
		}
	}

	return result, nil
}

func FindMemberByGraphARNAndAccountID(ctx context.Context, conn *detective.Detective, graphARN string, accountID string) (*detective.MemberDetail, error) {
	input := &detective.ListMembersInput{
		GraphArn: aws.String(graphARN),
	}

	var result *detective.MemberDetail

	err := conn.ListMembersPagesWithContext(ctx, input, func(page *detective.ListMembersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, member := range page.MemberDetails {
			if member == nil {
				continue
			}

			if aws.StringValue(member.AccountId) == accountID {
				result = member
				return false
			}
		}

		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &resource.NotFoundError{
			Message:     fmt.Sprintf("No member found with arn %q and accountID %q", graphARN, accountID),
			LastRequest: input,
		}
	}

	return result, nil
}
