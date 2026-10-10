// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func validEndpointID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}

func validReplicationInstanceID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 63 {
		es = append(es, fmt.Errorf("%q must not be longer than 63 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}

func validReplicationSubnetGroupID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if val == "default" {
		es = append(es, fmt.Errorf("%q must not be default", k))
	}
	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z_ .-]+$`).MatchString(val) {
		es = append(es, fmt.Errorf("%q must only contain alphanumeric characters, periods, spaces, underscores and hyphens", k))
	}

	return
}

func validReplicationTaskID(v any, k string) (ws []string, es []error) {
	val := v.(string)

	if len(val) > 255 {
		es = append(es, fmt.Errorf("%q must not be longer than 255 characters", k))
	}
	if !regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$").MatchString(val) {
		es = append(es, fmt.Errorf("%q must start with a letter, only contain alphanumeric characters and hyphens", k))
	}
	if strings.Contains(val, "--") {
		es = append(es, fmt.Errorf("%q must not contain consecutive hyphens", k))
	}
	if strings.HasSuffix(val, "-") {
		es = append(es, fmt.Errorf("%q must not end in a hyphen", k))
	}

	return
}

// validSecretsManagerSecretID validates a Secrets Manager secret identifier as
// accepted by the DMS SecretsManagerSecretId parameter: a full ARN, a partial
// ARN, or the friendly name of the secret.
func validSecretsManagerSecretID(v any, k string) (ws []string, es []error) {
	val, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if strings.HasPrefix(val, "arn:") {
		return verify.ValidARN(v, k)
	}

	if len(val) < 1 || len(val) > 512 {
		es = append(es, fmt.Errorf("%q must be between 1 and 512 characters, got %d", k, len(val)))
	}
	if !regexache.MustCompile(`^[0-9A-Za-z/_+=.@-]*$`).MatchString(val) {
		es = append(es, fmt.Errorf("%q must be a valid ARN or a secret name containing only alphanumeric characters and /_+=.@- special characters", k))
	}

	return
}
