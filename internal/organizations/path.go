package organizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

// organizationsClient defines the subset of AWS Organizations API used by BuildPrincipalOrgPath
type organizationsClient interface {
	ListParents(ctx context.Context, params *organizations.ListParentsInput, optFns ...func(*organizations.Options)) (*organizations.ListParentsOutput, error)
	DescribeOrganization(ctx context.Context, params *organizations.DescribeOrganizationInput, optFns ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error)
}

// BuildPrincipalOrgPath constructs the principal path for an Organization entity.
// Uses the ID prefixes returned by AWS API: Org (o-...), Root (r-...), OU (ou-...), Account (12-digit ID)
// See official AWS documentation for path format:
// https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_last-accessed-view-data-orgs.html#access_policies_last-accessed-viewing-orgs-entity-path
func BuildPrincipalOrgPath(ctx context.Context, client organizationsClient, childID string) (string, error) {
	var pathSegments []string
	currentID := childID

	for {
		pathSegments = append([]string{currentID}, pathSegments...)

		output, err := client.ListParents(ctx, &organizations.ListParentsInput{
			ChildId: aws.String(currentID),
		})
		if err != nil {
			return "", fmt.Errorf("listing parents for %s: %w", currentID, err)
		}

		if len(output.Parents) == 0 {
			return "", fmt.Errorf("no parent found for %s", currentID)
		}

		parent := output.Parents[0]
		parentID := aws.ToString(parent.Id)

		if parent.Type == types.ParentTypeRoot {
			orgOut, err := client.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
			if err != nil {
				return "", fmt.Errorf("describing organization: %w", err)
			}
			orgID := aws.ToString(orgOut.Organization.Id)
			rootID := parentID

			fullPath := fmt.Sprintf("%s/%s/%s/", orgID, rootID, strings.Join(pathSegments, "/"))
			return fullPath, nil
		}

		currentID = parentID
	}
}
