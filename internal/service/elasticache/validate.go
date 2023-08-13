// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

func validReplicationGroupAuthToken(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if (len(value) < 16) || (len(value) > 128) {
		errors = append(errors, fmt.Errorf(
			"%q must contain from 16 to 128 alphanumeric characters or symbols (excluding @, \", and /)", k))
	}
	if !regexp.MustCompile(`^[^@"\/]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters or symbols (excluding @, \", and /) allowed in %q", k))
	}
	return
}

func validNodeGroupSlotsFormat(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^\d+-\d+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q contains wrong format for keyspace, must be startkey-endkey", k))
		return
	}
	keyspaces := strings.Split(value, "-")
	intKeyspaces := []int{}
	for _, ks := range keyspaces {
		i, _ := strconv.Atoi(ks)
		intKeyspaces = append(intKeyspaces, i)
		if i > 16383 {
			errors = append(errors, fmt.Errorf("keyspace %q is outside maximum keyspace (16383)", ks))
		}
	}
	if intKeyspaces[0] > intKeyspaces[1] {
		errors = append(errors, fmt.Errorf("%q contains wrong format for keyspace, startkey must not be greater than startkey", k))
	}
	return
}

func validNodeGroupConfiguration(numNodeGroups *int64, configs []*elasticache.NodeGroupConfiguration) (errorMessages []string) {
	if !validNumberOfNodeGroupConfiguration(numNodeGroups, configs) {
		errorMessages = append(errorMessages, `number of "node_group_configuration" must be match with "num_node_groups"`)
	}

	if !validReplicaCount(configs) {
		errorMessages = append(errorMessages, `number of "replica_availability_zones" must be match with "replica_count"`)
	}

	if !validSlotsCoverage(configs) {
		errorMessages = append(errorMessages, `"slots" must be specified for none or all covering 0-16383 keyspaces without intersections`)
	}

	return errorMessages
}

func validNumberOfNodeGroupConfiguration(numNodeGroups *int64, configs []*elasticache.NodeGroupConfiguration) bool {
	return numNodeGroups == nil || len(configs) == int(aws.Int64Value(numNodeGroups))
}

func validReplicaCount(configs []*elasticache.NodeGroupConfiguration) bool {
	for _, c := range configs {
		if c.ReplicaCount != nil &&
			c.ReplicaAvailabilityZones != nil &&
			int(aws.Int64Value(c.ReplicaCount)) != len(c.ReplicaAvailabilityZones) {
			return false
		}
	}
	return true
}

func validSlotsCoverage(configs []*elasticache.NodeGroupConfiguration) bool {
	numOccupiedKeySpaces := 0
	for _, c := range configs {
		if c.Slots != nil {
			keyspaces := strings.Split(*c.Slots, "-")
			startKey, _ := strconv.Atoi(keyspaces[0])
			endKey, _ := strconv.Atoi(keyspaces[1])
			numOccupiedKeySpaces += endKey - startKey + 1
		}
	}
	return numOccupiedKeySpaces == 16384 || numOccupiedKeySpaces == 0
}
