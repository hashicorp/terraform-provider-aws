// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"fmt"
	"strings"
)

func portfolioShareParseResourceID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected portfolioID:type:principalID", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func portfolioShareCreateResourceID(portfolioID, shareType, principalID string) string {
	return strings.Join([]string{portfolioID, shareType, principalID}, ":")
}

func productPortfolioAssociationParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected acceptLanguage:portfolioID:productID", id)
	}

	return parts[0], parts[1], parts[2], nil
}

func productPortfolioAssociationCreateID(acceptLanguage, portfolioID, productID string) string {
	return strings.Join([]string{acceptLanguage, portfolioID, productID}, ":")
}

func budgetResourceAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), budgetName:resourceID", id)
	}

	return parts[0], parts[1], nil
}

func budgetResourceAssociationID(budgetName, resourceID string) string {
	return strings.Join([]string{budgetName, resourceID}, ":")
}

func tagOptionResourceAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), tagOptionID:resourceID", id)
	}

	return parts[0], parts[1], nil
}

func tagOptionResourceAssociationID(tagOptionID, resourceID string) string {
	return strings.Join([]string{tagOptionID, resourceID}, ":")
}

func provisioningArtifactID(artifactID, productID string) string {
	return strings.Join([]string{artifactID, productID}, ":")
}

func provisioningArtifactParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected artifactID:productID", id)
	}
	return parts[0], parts[1], nil
}

func portfolioConstraintsID(acceptLanguage, portfolioID, productID string) string {
	return strings.Join([]string{acceptLanguage, portfolioID, productID}, ":")
}
