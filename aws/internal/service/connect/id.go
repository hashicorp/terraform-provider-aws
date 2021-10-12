package connect

import (
	"fmt"
	"strings"
)

const lambdaFunctionAssociationIDSeparator = ":"

func LambdaFunctionAssociationParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, lambdaFunctionAssociationIDSeparator, 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "",
		fmt.Errorf("unexpected format for ID (%q), expected instanceID"+lambdaFunctionAssociationIDSeparator+
			"function-arn", id)
}

func LambdaFunctionAssociationID(instanceID string, functionArn string) string {
	return strings.Join([]string{instanceID, functionArn}, lambdaFunctionAssociationIDSeparator)
}
