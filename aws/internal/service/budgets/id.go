package budgets

import (
	"fmt"
	"strings"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const budgetActionResourceIDSeparator = ":"

func BudgetActionCreateResourceID(accountID, actionID, budgetName string) string {
	parts := []string{accountID, actionID, budgetName}
	id := strings.Join(parts, budgetActionResourceIDSeparator)

	return id
}

func BudgetActionParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, budgetActionResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sActionID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}

const budgetResourceIDSeparator = ":"

func BudgetCreateResourceID(accountID, budgetName string) string {
	parts := []string{accountID, budgetName}
	id := strings.Join(parts, budgetResourceIDSeparator)

	return id
}

func BudgetParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, budgetResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AccountID%[2]sBudgetName", id, budgetActionResourceIDSeparator)
}
