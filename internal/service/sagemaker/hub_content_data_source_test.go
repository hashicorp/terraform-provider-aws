// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccSageMakerHubContentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_sagemaker_hub_content.test"
	resourceName := "aws_sagemaker_hub_content_reference.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHubContentReferenceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentDataSourceConfig_basic(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("hub_arn"), resourceName, tfjsonpath.New("hub_arn"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("hub_content_arn"), resourceName, tfjsonpath.New("hub_content_arn"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("hub_content_version"), resourceName, tfjsonpath.New("hub_content_version"), compare.ValuesSame()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("sagemaker_public_hub_content_arn"), resourceName, tfjsonpath.New("sagemaker_public_hub_content_arn"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrCreationTime), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_status"), tfknownvalue.StringExact(awstypes.HubContentStatusAvailable)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_type"), tfknownvalue.StringExact(awstypes.HubContentTypeModelReference)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("versioned_hub_content_arn"), knownvalue.StringRegexp(regexache.MustCompile(`:hub-content/`+rName+`/ModelReference/`+rName+`/\d+\.\d+\.\d+$`))),
				},
			},
		},
	})
}

func TestAccSageMakerHubContentDataSource_publicHubModel(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_sagemaker_hub_content.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.SageMakerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccHubContentDataSourceConfig_publicHubModel(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_arn"), knownvalue.StringRegexp(regexache.MustCompile(`:hub/SageMakerPublicHub$`))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_arn"), knownvalue.StringRegexp(regexache.MustCompile(`:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-2-13b-f$`))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_document"), knownvalue.StringRegexp(regexache.MustCompile(`"HostingArtifactUri"`))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_status"), tfknownvalue.StringExact(awstypes.HubContentStatusAvailable)),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("hub_content_version"), knownvalue.StringRegexp(regexache.MustCompile(`^\d+\.\d+\.\d+$`))),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("versioned_hub_content_arn"), knownvalue.StringRegexp(regexache.MustCompile(`:aws:hub-content/SageMakerPublicHub/Model/meta-textgeneration-llama-2-13b-f/\d+\.\d+\.\d+$`))),
				},
			},
		},
	})
}

func testAccHubContentDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccHubContentReferenceConfig_basic(rName), `
data "aws_sagemaker_hub_content" "test" {
  hub_name         = aws_sagemaker_hub_content_reference.test.hub_name
  hub_content_name = aws_sagemaker_hub_content_reference.test.hub_content_name
  hub_content_type = "ModelReference"
}
`)
}

func testAccHubContentDataSourceConfig_publicHubModel() string {
	return `
data "aws_sagemaker_hub_content" "test" {
  hub_name         = "SageMakerPublicHub"
  hub_content_name = "meta-textgeneration-llama-2-13b-f"
  hub_content_type = "Model"
}
`
}
