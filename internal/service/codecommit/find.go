// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codecommit

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// FindApprovalRuleTemplateAssociation validates that an approval rule template has the named associated repository
func FindApprovalRuleTemplateAssociation(ctx context.Context, conn *codecommit.CodeCommit, approvalRuleTemplateName, repositoryName string) error {
	input := &codecommit.ListRepositoriesForApprovalRuleTemplateInput{
		ApprovalRuleTemplateName: aws.String(approvalRuleTemplateName),
	}

	found := false

	err := conn.ListRepositoriesForApprovalRuleTemplatePagesWithContext(ctx, input, func(page *codecommit.ListRepositoriesForApprovalRuleTemplateOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, repoName := range page.RepositoryNames {
			if aws.StringValue(repoName) == repositoryName {
				found = true
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, codecommit.ErrCodeApprovalRuleTemplateDoesNotExistException) {
		return &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return err
	}

	if !found {
		return &retry.NotFoundError{
			Message:     fmt.Sprintf("No approval rule template (%q) associated with repository (%q)", approvalRuleTemplateName, repositoryName),
			LastRequest: input,
		}
	}

	return nil
}
