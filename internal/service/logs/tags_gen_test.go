// Code generated by internal/generate/tagstests/main.go; DO NOT EDIT.

package logs_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
)

func expectFullResourceTags(ctx context.Context, resourceAddress string, knownValue knownvalue.Check) statecheck.StateCheck {
	return tfstatecheck.ExpectFullResourceTags(tflogs.ServicePackage(ctx), resourceAddress, knownValue)
}

func expectFullDataSourceTags(ctx context.Context, resourceAddress string, knownValue knownvalue.Check) statecheck.StateCheck {
	return tfstatecheck.ExpectFullDataSourceTags(tflogs.ServicePackage(ctx), resourceAddress, knownValue)
}
