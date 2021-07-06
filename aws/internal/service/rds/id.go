package rds

import (
	"fmt"
	"strings"
)

func ResourceAwsDbProxyEndpointParseID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "/", 2)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected db_proxy_name/db_proxy_endpoint_name", id)
	}
	return idParts[0], idParts[1], nil
}

const dbClusterRoleAssociationResourceIDSeparator = ","

func DBClusterRoleAssociationCreateResourceID(dbClusterID, roleARN string) string {
	parts := []string{dbClusterID, roleARN}
	id := strings.Join(parts, dbClusterRoleAssociationResourceIDSeparator)

	return id
}

func DBClusterRoleAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, dbClusterRoleAssociationResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DBCLUSTERID%[2]sROLEARN", id, dbClusterRoleAssociationResourceIDSeparator)
}
