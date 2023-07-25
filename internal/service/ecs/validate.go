// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validateClusterName(v interface{}, k string) (ws []string, errors []error) {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(
			regexp.MustCompile("[a-zA-Z0-9_-]+"),
			"The cluster name must consist of alphanumerics, hyphens, and underscores."),
	)(v, k)
}

// Validates that ECS Placement Constraints are set correctly
// Takes type, and expression as strings
func validPlacementConstraint(constType, constExpr string) error {
	switch constType {
	case "distinctInstance":
		// Expression can be nil for distinctInstance
		return nil
	case "memberOf":
		if constExpr == "" {
			return fmt.Errorf("Expression cannot be nil for 'memberOf' type")
		}
	default:
		return fmt.Errorf("Unknown type provided: %q", constType)
	}
	return nil
}

// Validates that an Ecs placement strategy is set correctly
// Takes type, and field as strings
func validPlacementStrategy(stratType, stratField string) error {
	switch stratType {
	case "random":
		// random requires the field attribute to be unset.
		if stratField != "" {
			return fmt.Errorf("Random type requires the field attribute to be unset. Got: %s",
				stratField)
		}
	case "spread":
		//  For the spread placement strategy, valid values are instanceId
		// (or host, which has the same effect), or any platform or custom attribute
		// that is applied to a container instance
		// stratField is already cased to a string
		return nil
	case "binpack":
		if stratField != "cpu" && stratField != "memory" {
			return fmt.Errorf("Binpack type requires the field attribute to be either 'cpu' or 'memory'. Got: %s",
				stratField)
		}
	default:
		return fmt.Errorf("Unknown type %s. Must be one of 'random', 'spread', or 'binpack'.", stratType)
	}
	return nil
}
