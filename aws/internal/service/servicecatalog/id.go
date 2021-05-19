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

func ProductPortfolioAssociationParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected acceptLanguage:portfolioID:productID", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func ProductPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID string) string {
	return strings.Join([]string{acceptLanguage, portfolioID, productID}, ":")
}

func BudgetResourceAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), budgetName:resourceID", id)
	}

	return parts[0], parts[1], nil
}

func BudgetResourceAssociationID(budgetName, resourceID string) string {
	return strings.Join([]string{budgetName, resourceID}, ":")
}

func TagOptionResourceAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), tagOptionID:resourceID", id)
	}

	return parts[0], parts[1], nil
}

func TagOptionResourceAssociationID(tagOptionID, resourceID string) string {
	return strings.Join([]string{tagOptionID, resourceID}, ":")
}

func ProvisioningArtifactID(artifactID, productID string) string {
	return strings.Join([]string{artifactID, productID}, ":")
}

func ProvisioningArtifactParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Please make sure the ID is in the form artifact_id:product_id (i.e. pa-r2d2slrtcob:prod-c3pohcrhmisy")
	}
	return parts[0], parts[1], nil
}
