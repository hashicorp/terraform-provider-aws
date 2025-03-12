// Code generated by internal/generate/tagstests/main.go; DO NOT EDIT.

package appautoscaling_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	tfappautoscaling "github.com/hashicorp/terraform-provider-aws/internal/service/appautoscaling"
)

func expectFullResourceTags(ctx context.Context, resourceAddress string, knownValue knownvalue.Check) statecheck.StateCheck {
	return tfstatecheck.ExpectFullResourceTags(tfappautoscaling.ServicePackage(ctx), resourceAddress, knownValue)
}
