package rds

import (
	"fmt"
	"strings"
)

func ResourceAwsDbProxyEndpointParseID(id string) (string, string, string, error) {
	idParts := strings.SplitN(id, "/", 3)
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/db_proxy_endpoint_name/arn", id)
	}
	return idParts[0], idParts[1], idParts[2], nil
}
