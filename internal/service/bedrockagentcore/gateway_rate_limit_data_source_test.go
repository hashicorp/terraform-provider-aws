// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGatewayRateLimitDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_bedrockagentcore_gateway_rate_limit.test"
	resourceName := "aws_bedrockagentcore_gateway_rate_limit.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimitDataSource/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Values shared with the resource must agree.
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New("rate_limit_id"),
						resourceName, tfjsonpath.New("rate_limit_id"),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New(names.AttrDescription),
						resourceName, tfjsonpath.New(names.AttrDescription),
						compare.ValuesSame(),
					),
					statecheck.CompareValuePairs(
						dataSourceName, tfjsonpath.New("dimension_keys"),
						resourceName, tfjsonpath.New("dimension_keys"),
						compare.ValuesSame(),
					),

					// entries is a computed list here but a set on the resource, so
					// the two cannot be compared directly.
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("entries"), knownvalue.ListSizeExact(2)),

					// Attributes the resource deliberately does not expose.
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrStatus), knownvalue.StringExact("ACTIVE")),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCreatedAt), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
				},
			},
		},
	})
}

// A data source must error when nothing matches, rather than returning empty.
func TestAccBedrockAgentCoreGatewayRateLimitDataSource_notFound(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/GatewayRateLimitDataSource/not_found/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				// The finder wraps the API's ResourceNotFoundException in a
				// retry.NotFoundError, so the raw exception name never reaches
				// the diagnostic - match what smerr actually renders. The ID
				// line pins the two-part key so the gateway identifier is
				// guaranteed to appear alongside the rate limit id.
				ExpectError: regexache.MustCompile(`ID: [^,\n]+,does-not-exist`),
			},
		},
	})
}
