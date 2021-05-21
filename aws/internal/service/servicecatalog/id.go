package servicecatalog

import (
	"fmt"
	"strings"
)

func PortfolioShareParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected portfolioID:type:principalID", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func PortfolioShareCreateResourceID(portfolioID, shareType, principalID string) string {
	return strings.Join([]string{portfolioID, shareType, principalID}, ":")
}
