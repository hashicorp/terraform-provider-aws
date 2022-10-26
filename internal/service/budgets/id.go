package budgets

import (
	"fmt"
	"strings"
)

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
