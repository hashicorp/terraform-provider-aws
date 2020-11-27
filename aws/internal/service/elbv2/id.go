package elbv2

import (
	"fmt"
	"strings"
)

const listnerCertificateIDSeparator = "_"

func ListnerCertificateParseID(id string) (string, string, error) {
	parts := strings.Split(id, listnerCertificateIDSeparator)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "",
		fmt.Errorf("unexpected format for ID (%q), expected listner-arn"+listnerCertificateIDSeparator+
			"certificate-arn", id)
}
