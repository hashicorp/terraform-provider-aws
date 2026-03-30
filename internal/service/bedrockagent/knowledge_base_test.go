// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagent_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagent "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagent"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccKnowledgeBase_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_S3VectorsByIndexARN(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagent.ResourceKnowledgeBase, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccKnowledgeBase_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_tags1(rName, foundationModel, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKnowledgeBaseConfig_tags2(rName, foundationModel, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccKnowledgeBaseConfig_tags1(rName, foundationModel, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

// Prerequisites:
// * psql run via null_resource/provisioner "local-exec"
// * jq for parsing output from aws cli to retrieve postgres password
func testAccKnowledgeBase_RDS_basic(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v1"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_RDS_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(0),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(0),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(0),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeRds),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(1),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccKnowledgeBase_OpenSearchServerless_basic(t *testing.T) {
	ctx := acctest.Context(t)
	collectionName := skipIfOSSCollectionNameEnvVarNotSet(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_OpenSearchServerless_basic(rName, collectionName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(0),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(0),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(1),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeOpensearchServerless),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(0),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccKnowledgeBase_update(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_description(rName, foundationModel, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKnowledgeBaseConfig_description(rName, foundationModel, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccKnowledgeBase_RDS_supplementalDataStorage(t *testing.T) {
	acctest.SkipIfExeNotOnPath(t, "psql")
	acctest.SkipIfExeNotOnPath(t, "jq")
	acctest.SkipIfExeNotOnPath(t, "aws")

	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v1"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"null": {
				Source:            "hashicorp/null",
				VersionConstraint: "3.2.2",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_RDS_supplementalDataStorage(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectPartial(map[string]knownvalue.Check{
									"supplemental_data_storage_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"storage_location": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													"s3_location":  knownvalue.ListSizeExact(1),
													names.AttrType: tfknownvalue.StringExact(awstypes.SupplementalDataStorageLocationTypeS3),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccKnowledgeBase_Kendra_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	// Index should be created with the "GEN_AI_ENTERPRISE_EDITION" edition and be "ACTIVE".
	kendraIndexARN := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_KENDRA_INDEX_ARN")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_Kendra_basic(rName, kendraIndexARN),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(1),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeKendra),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
		},
	})
}

func testAccKnowledgeBase_OpenSearchManagedCluster_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckIAMServiceLinkedRole(ctx, t, "/aws-service-role/opensearchservice.amazonaws.com")
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"opensearch": {
				Source:            "opensearch-project/opensearch",
				VersionConstraint: "~> 2.2.0",
			},
		},
		CheckDestroy: testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_OpenSearchManagedCluster_basic(rName, foundationModel),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(0),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(1),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(0),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeOpensearchManagedCluster),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(0),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
		},
	})
}

func testAccKnowledgeBase_S3Vectors_update(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"
	foundationModel := "amazon.titan-embed-text-v2:0"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_S3VectorsByIndexARN(rName, foundationModel),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(0),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(0),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(0),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeS3Vectors),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(0),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(1),
						}),
					})),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKnowledgeBaseConfig_S3VectorsByIndexName(rName, foundationModel),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(0),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(0),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(0),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeS3Vectors),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(0),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(1),
						}),
					})),
				},
			},
		},
	})
}

func testAccKnowledgeBase_StructuredDataStore_redshiftProvisioned(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_StructuredDataStore_redshiftProvisioned(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrType: tfknownvalue.StringExact(awstypes.QueryEngineTypeRedshift),
									"redshift_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"query_engine_configuration": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrType:              tfknownvalue.StringExact(awstypes.RedshiftQueryEngineTypeProvisioned),
													"provisioned_configuration": knownvalue.ListSizeExact(1),
													"serverless_configuration":  knownvalue.ListSizeExact(0),
												}),
											}),
											"query_generation_configuration": knownvalue.ListSizeExact(0),
											"storage_configuration": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrType:                   tfknownvalue.StringExact(awstypes.RedshiftQueryEngineStorageTypeRedshift),
													"aws_data_catalog_configuration": knownvalue.ListSizeExact(0),
													"redshift_configuration":         knownvalue.ListSizeExact(1),
												}),
											}),
										}),
									}),
								}),
							}),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeSql),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
		},
	})
}

func testAccKnowledgeBase_StructuredDataStore_redshiftServerless(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_StructuredDataStore_redshiftServerless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									names.AttrType: tfknownvalue.StringExact(awstypes.QueryEngineTypeRedshift),
									"redshift_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"query_engine_configuration": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrType:              tfknownvalue.StringExact(awstypes.RedshiftQueryEngineTypeServerless),
													"provisioned_configuration": knownvalue.ListSizeExact(0),
													"serverless_configuration":  knownvalue.ListSizeExact(1),
												}),
											}),
											"query_generation_configuration": knownvalue.ListSizeExact(1),
											"storage_configuration": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													names.AttrType:                   tfknownvalue.StringExact(awstypes.RedshiftQueryEngineStorageTypeRedshift),
													"aws_data_catalog_configuration": knownvalue.ListSizeExact(0),
													"redshift_configuration":         knownvalue.ListSizeExact(1),
												}),
											}),
										}),
									}),
								}),
							}),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeSql),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
		},
	})
}

func testAccKnowledgeBase_NeptuneAnalytics_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var knowledgebase awstypes.KnowledgeBase
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagent_knowledge_base.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKnowledgeBaseDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKnowledgeBaseConfig_NeptuneAnalytics_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKnowledgeBaseExists(ctx, t, resourceName, &knowledgebase),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("knowledge_base_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"kendra_knowledge_base_configuration": knownvalue.ListSizeExact(0),
							"sql_knowledge_base_configuration":    knownvalue.ListSizeExact(0),
							names.AttrType:                        tfknownvalue.StringExact(awstypes.KnowledgeBaseTypeVector),
							"vector_knowledge_base_configuration": knownvalue.ListSizeExact(1),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("storage_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.MapExact(map[string]knownvalue.Check{
							"mongo_db_atlas_configuration":             knownvalue.ListSizeExact(0),
							"neptune_analytics_configuration":          knownvalue.ListSizeExact(1),
							"opensearch_managed_cluster_configuration": knownvalue.ListSizeExact(0),
							"opensearch_serverless_configuration":      knownvalue.ListSizeExact(0),
							names.AttrType:                             tfknownvalue.StringExact(awstypes.KnowledgeBaseStorageTypeNeptuneAnalytics),
							"pinecone_configuration":                   knownvalue.ListSizeExact(0),
							"rds_configuration":                        knownvalue.ListSizeExact(0),
							"redis_enterprise_cloud_configuration":     knownvalue.ListSizeExact(0),
							"s3_vectors_configuration":                 knownvalue.ListSizeExact(0),
						}),
					})),
				},
			},
		},
	})
}

func testAccCheckKnowledgeBaseDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagent_knowledge_base" {
				continue
			}

			_, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Knowledge Base %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckKnowledgeBaseExists(ctx context.Context, t *testing.T, n string, v *awstypes.KnowledgeBase) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentClient(ctx)

		output, err := tfbedrockagent.FindKnowledgeBaseByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

// skipIfOSSCollectionNameEnvVarNotSet handles skipping tests when an environment
// variable providing a valid OSS collection name is unset
//
// This should be called in all acceptance tests currently dependent on an OpenSearch
// Serverless vector collection.
//
// To create a collection to be used with this environment variable, do the following.
//
// 1. In the AWS console, navigate to the OpenSearch service. Choose the "Collections"
// entry on the left navbar and select "Create collection" above the collections table.
// 2. Enter a collection name. Choose "Vector search" as the collection type. Choose
// "Easy create" in the security section. Click "Next" to review, then click "Submit".
// 3. Once the collection is available, select the "Indexes" tab. Click "Create vector
// index". Name the index "bedrock-knowledge-base-default-index".
// 4. In the "Vector fields" section, click "Add vector field".  Name the field
// "bedrock-knowledge-base-default-vector". Choose "faiss" for engine, "FP32" for
// precision, "1024" for dimensions, and "Euclidean" for distance metric. Click
// "Confirm" to create the field.
//  5. In the "Metadata management" section, add two fields.
//     "AMAZON_BEDROCK_METADATA" - string type, filterable is false.
//     "AMAZON_BEDROCK_TEXT_CHUNK" - string type, filterable is true.
//  6. Click "Create" to finish index creation.
//
// At this point the collection is usable with this test. Set the collection name to the
// environment variable below.
func skipIfOSSCollectionNameEnvVarNotSet(t *testing.T) string {
	t.Helper()
	return acctest.SkipIfEnvVarNotSet(t, "TF_AWS_BEDROCK_OSS_COLLECTION_NAME")
}

func testAccKnowledgeBaseConfig_tags1(rName, model, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.test.index_arn
    }
  }

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, model, tag1Key, tag1Value))
}

func testAccKnowledgeBaseConfig_tags2(rName, model, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.test.index_arn
    }
  }

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, model, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccKnowledgeBaseConfig_description(rName, model, description string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy.test,
  ]

  name        = %[1]q
  role_arn    = aws_iam_role.test.arn
  description = %[3]q

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.test.index_arn
    }
  }
}
`, rName, model, description))
}

func testAccKnowledgeBaseConfig_baseRDS(rName, model string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsEnableDNSHostnames(rName, 2), fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_iam_role" "test" {
  name               = %[1]q
  path               = "/service-role/"
  assume_role_policy = <<POLICY
{
	"Version": "2012-10-17",
	"Statement": [{
		"Action": "sts:AssumeRole",
		"Principal": {
		"Service": "bedrock.amazonaws.com"
		},
		"Effect": "Allow"
	}]
}
POLICY
}

# See https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html.
resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:ListFoundationModels",
        "bedrock:ListCustomModels"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "bedrock:InvokeModel"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      ]
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "rds_data_full_access" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/AmazonRDSDataFullAccess"
}

resource "aws_iam_role_policy_attachment" "secrets_manager_read_write" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_partition.current.partition}:policy/SecretsManagerReadWrite"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = "aurora-postgresql"
  engine_latest_version      = true
  preferred_instance_classes = ["db.serverless"]
}

resource "aws_rds_cluster" "test" {
  cluster_identifier          = %[1]q
  master_username             = "test"
  manage_master_user_password = true
  database_name               = "test"
  skip_final_snapshot         = true
  engine                      = data.aws_rds_orderable_db_instance.test.engine
  engine_version              = data.aws_rds_orderable_db_instance.test.engine_version
  enable_http_endpoint        = true
  vpc_security_group_ids      = [aws_security_group.test.id]
  db_subnet_group_name        = aws_db_subnet_group.test.name

  serverlessv2_scaling_configuration {
    max_capacity = 1.0
    min_capacity = 0.5
  }
}

resource "aws_rds_cluster_instance" "test" {
  cluster_identifier  = aws_rds_cluster.test.id
  instance_class      = "db.serverless"
  engine              = aws_rds_cluster.test.engine
  engine_version      = aws_rds_cluster.test.engine_version
  publicly_accessible = true
}

resource "aws_db_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_rds_cluster.test.master_user_secret[0].secret_arn
  version_stage = "AWSCURRENT"
  depends_on    = [aws_rds_cluster.test]
}

resource "null_resource" "db_setup" {
  depends_on = [aws_rds_cluster_instance.test, aws_rds_cluster.test, data.aws_secretsmanager_secret_version.test]

  provisioner "local-exec" {
    command = <<EOT
      sleep 60
      export PGPASSWORD=$(aws secretsmanager get-secret-value --secret-id '${aws_rds_cluster.test.master_user_secret[0].secret_arn}' --version-stage AWSCURRENT --region ${data.aws_region.current.region} --query SecretString --output text | jq -r '."password"')
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE EXTENSION IF NOT EXISTS vector;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_integration;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE SCHEMA IF NOT EXISTS bedrock_new;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE ROLE bedrock_user WITH PASSWORD '$PGPASSWORD' LOGIN;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "GRANT ALL ON SCHEMA bedrock_integration TO bedrock_user;"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE TABLE bedrock_integration.bedrock_kb (id uuid PRIMARY KEY, embedding vector(1536), chunks text, metadata json, custom_metadata jsonb);"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE INDEX ON bedrock_integration.bedrock_kb USING hnsw (embedding vector_cosine_ops);"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE INDEX ON bedrock_integration.bedrock_kb USING gin (to_tsvector('simple', chunks));"
      psql -h ${aws_rds_cluster.test.endpoint} -U ${aws_rds_cluster.test.master_username} -d ${aws_rds_cluster.test.database_name} -c "CREATE INDEX ON bedrock_integration.bedrock_kb USING gin (custom_metadata);"
    EOT
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_RDS_basic(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseRDS(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "RDS"
    rds_configuration {
      resource_arn           = aws_rds_cluster.test.arn
      credentials_secret_arn = tolist(aws_rds_cluster.test.master_user_secret)[0].secret_arn
      database_name          = aws_rds_cluster.test.database_name
      table_name             = "bedrock_integration.bedrock_kb"
      field_mapping {
        vector_field          = "embedding"
        text_field            = "chunks"
        metadata_field        = "metadata"
        primary_key_field     = "id"
        custom_metadata_field = "custom_metadata"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, null_resource.db_setup]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_RDS_supplementalDataStorage(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseRDS(rName, model), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "test_s3" {
  statement {
    effect = "Allow"
    actions = [
      "s3:DeleteObject",
      "s3:GetObject",
      "s3:ListBucket",
      "s3:PutObject",
    ]
    resources = [
      "*",
    ]
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalAccount"
      values = [
        data.aws_caller_identity.current.account_id,
      ]
    }
  }
}

resource "aws_iam_policy" "test_s3" {
  name   = "%[1]s-s3"
  policy = data.aws_iam_policy_document.test_s3.json
}

resource "aws_iam_role_policy_attachment" "test_s3" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test_s3.arn
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"

      supplemental_data_storage_configuration {
        storage_location {
          type = "S3"
          s3_location {
            uri = "s3://${aws_s3_bucket.test.bucket}"
          }
        }
      }
    }

    type = "VECTOR"
  }

  storage_configuration {
    type = "RDS"

    rds_configuration {
      resource_arn           = aws_rds_cluster.test.arn
      credentials_secret_arn = tolist(aws_rds_cluster.test.master_user_secret)[0].secret_arn
      database_name          = aws_rds_cluster.test.database_name
      table_name             = "bedrock_integration.bedrock_kb"

      field_mapping {
        vector_field          = "embedding"
        text_field            = "chunks"
        metadata_field        = "metadata"
        primary_key_field     = "id"
        custom_metadata_field = "custom_metadata"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_iam_role_policy_attachment.test_s3, null_resource.db_setup]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_baseOpenSearchServerless(rName, collectionName, model string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

# As a future enhancment, the creation and preparation of the OSS collection
# will be done within this configuration. Creation of the collection is
# possible today, but creation of the appropriate vector index must be done
# out of band via awscurl or some other mechanism.
#
# Ref: https://docs.aws.amazon.com/opensearch-service/latest/developerguide/serverless-vector-search.html
data "aws_opensearchserverless_collection" "test" {
  name = %[2]q
}

# See the Amazon Bedrock documentation for creating a service role:
# https://docs.aws.amazon.com/bedrock/latest/userguide/kb-permissions.html
data "aws_iam_policy_document" "test_trust" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:SourceAccount"
      values = [
        data.aws_caller_identity.current.account_id,
      ]
    }
    condition {
      test     = "ArnLike"
      variable = "aws:SourceArn"
      values = [
        "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
      ]
    }
  }
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"
    actions = [
      "bedrock:ListFoundationModels",
      "bedrock:ListCustomModels",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "bedrock:InvokeModel",
    ]
    resources = [
      "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:foundation-model/%[3]s",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "bedrock:RetreiveAndGenerate",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "aoss:APIAccessAll",
    ]
    resources = [
      data.aws_opensearchserverless_collection.test.arn
    ]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_trust.json
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_opensearchserverless_access_policy" "test" {
  name = %[1]q
  type = "data"
  policy = jsonencode([
    {
      Rules = [
        {
          ResourceType = "index",
          Resource = [
            "index/%[2]s/*"
          ],
          Permission = [
            "aoss:*",
          ]
        },
        {
          ResourceType = "collection",
          Resource = [
            "collection/%[2]s"
          ],
          Permission = [
            "aoss:*",
          ]
        }
      ],
      Principal = [
        data.aws_caller_identity.current.arn,
        aws_iam_role.test.arn,
      ]
    }
  ])
}
`, rName, collectionName, model)
}

func testAccKnowledgeBaseConfig_OpenSearchServerless_basic(rName, collectionName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseOpenSearchServerless(rName, collectionName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy_attachment.test,
    aws_opensearchserverless_access_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
    }
    type = "VECTOR"
  }

  storage_configuration {
    type = "OPENSEARCH_SERVERLESS"
    opensearch_serverless_configuration {
      collection_arn    = data.aws_opensearchserverless_collection.test.arn
      vector_index_name = "bedrock-knowledge-base-default-index"
      field_mapping {
        vector_field   = "bedrock-knowledge-base-default-vector"
        text_field     = "AMAZON_BEDROCK_TEXT_CHUNK"
        metadata_field = "AMAZON_BEDROCK_METADATA"
      }
    }
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_Kendra_basic(rName, kendraIndexArn string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s-bedrock"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "kendra:Retrieve",
          "kendra:DescribeIndex"
        ]
        Resource = %[2]q
      }
    ]
  })
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "KENDRA"
    kendra_knowledge_base_configuration {
      kendra_index_arn = %[2]q
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, kendraIndexArn)
}

func testAccKnowledgeBaseConfig_baseOpenSearchManagedCluster(rName, model string) string {
	// lintignore:AT004
	return fmt.Sprintf(`
terraform {
  required_providers {
    opensearch = {
      source  = "opensearch-project/opensearch"
      version = "~> 2.2.0"
    }
  }
}

data "aws_partition" "current" {}
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}
data "aws_iam_role" "opensearch" {
  name = "AWSServiceRoleForAmazonOpenSearchService"
}

resource "aws_iam_role" "bedrock_kb_role" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "bedrock.${data.aws_partition.current.dns_suffix}"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" : data.aws_caller_identity.current.account_id
          },
        }
      }
    ]
  })
}

resource "aws_iam_policy" "bedrock_models_access" {
  name        = "bedrock-%[1]s"
  description = "IAM policy for Amazon Bedrock to access embedding models"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "bedrock:RetrieveAndGenerate"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "bedrock:ListFoundationModels",
          "bedrock:ListCustomModels"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel"
        ]
        Resource = [
          "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
        ]
      }
    ]
  })
}

resource "aws_iam_policy" "opensearch_access" {
  name        = "os-%[1]s"
  description = "IAM policy for Amazon Bedrock to access OpenSearch domain"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "es:ESHttpGet",
          "es:ESHttpPost",
          "es:ESHttpPut",
          "es:ESHttpDelete"
        ]
        Resource = [
          "*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "es:DescribeDomain",
          "es:DescribeElasticsearchDomain"
        ]
        Resource = [
          "*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "bedrock_models_access" {
  role       = aws_iam_role.bedrock_kb_role.name
  policy_arn = aws_iam_policy.bedrock_models_access.arn
}

resource "aws_iam_role_policy_attachment" "opensearch_access" {
  role       = aws_iam_role.bedrock_kb_role.name
  policy_arn = aws_iam_policy.opensearch_access.arn
}

resource "aws_opensearch_domain" "knowledge_base" {
  domain_name     = substr(%[1]q, 0, 28)
  engine_version  = "OpenSearch_3.1"
  access_policies = local.opensearch_access_policy

  cluster_config {
    instance_type            = "or2.medium.search"
    instance_count           = 1
    zone_awareness_enabled   = false
    dedicated_master_enabled = false
  }

  # Configure EBS volumes for the data nodes
  ebs_options {
    ebs_enabled = true
    volume_size = 20
    volume_type = "gp3"
  }

  # Enable encryption at rest
  encrypt_at_rest {
    enabled = true
  }

  # Enable node to node encryption
  node_to_node_encryption {
    enabled = true
  }

  # Configure domain endpoint options
  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-PFS-2023-10"
  }

  # Configure advanced security options
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = true

    master_user_options {
      master_user_name     = "admin"
      master_user_password = "Barbarbarbar1!"
    }
  }

  # Auto-Tune options
  auto_tune_options {
    desired_state = "ENABLED"
  }

  # Software update options
  software_update_options {
    auto_software_update_enabled = true
  }

  # This is required to reference the existing service-linked role
  depends_on = [
    data.aws_iam_role.opensearch
  ]
}

locals {
  opensearch_access_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          AWS = "*"
        }
        Action   = "es:*"
        Resource = "arn:${data.aws_partition.current.partition}:es:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:domain/${substr(%[1]q, 0, 28)}/*"
      }
    ]
  })
}

provider "opensearch" {
  url               = "https://${aws_opensearch_domain.knowledge_base.endpoint}"
  username          = "admin"
  password          = aws_opensearch_domain.knowledge_base.advanced_security_options[0].master_user_options[0].master_user_password
  insecure          = false
  aws_region        = data.aws_region.current.region
  healthcheck       = false
  sniff             = false
  sign_aws_requests = false
}

resource "opensearch_role" "os_kb_role" {
  role_name   = "kb-%[1]s"
  description = "Knowledge Base Role"

  cluster_permissions = ["*"]

  index_permissions {
    index_patterns  = ["*"]
    allowed_actions = ["*"]
  }

  tenant_permissions {
    tenant_patterns = ["*"]
    allowed_actions = ["*"]
  }
}

resource "opensearch_roles_mapping" "mapper" {
  role_name   = opensearch_role.os_kb_role.role_name
  description = "Mapping AWS IAM roles to ES role"

  backend_roles = [
    aws_iam_role.bedrock_kb_role.arn,
    data.aws_caller_identity.current.arn
  ]
}

resource "opensearch_index" "vector_index" {
  name               = "knowledge-index"
  number_of_shards   = "5"
  number_of_replicas = "1"
  index_knn          = true

  # Mappings for Bedrock Knowledge Base compatibility
  mappings = jsonencode({
    "properties" : {
      "vector_embedding" : {
        "type" : "knn_vector",
        "dimension" : 1024,
        "space_type" : "l2",
        "method" : {
          "name" : "hnsw",
          "engine" : "faiss",
          "parameters" : {
            "ef_construction" : 128,
            "m" : 24
          }
        }
      },
      "text" : {
        "type" : "text",
        "index" : true
      },
      "metadata" : {
        "type" : "text",
        "index" : false
      }
    }
  })

  lifecycle {
    ignore_changes = [mappings]
  }

  depends_on = [aws_opensearch_domain.knowledge_base]
}
`, rName, model)
}

func testAccKnowledgeBaseConfig_OpenSearchManagedCluster_basic(rName, model string) string {
	return acctest.ConfigCompose(testAccKnowledgeBaseConfig_baseOpenSearchManagedCluster(rName, model), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.bedrock_kb_role.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
    }
  }

  storage_configuration {
    type = "OPENSEARCH_MANAGED_CLUSTER"

    opensearch_managed_cluster_configuration {
      domain_arn        = aws_opensearch_domain.knowledge_base.arn
      domain_endpoint   = "https://${aws_opensearch_domain.knowledge_base.endpoint}"
      vector_index_name = "knowledge-index"

      field_mapping {
        vector_field   = "vector_embedding"
        text_field     = "text"
        metadata_field = "metadata"
      }
    }
  }

  depends_on = [
    aws_opensearch_domain.knowledge_base,
    opensearch_index.vector_index,
    opensearch_roles_mapping.mapper,
    aws_iam_role_policy_attachment.opensearch_access,
    aws_iam_role_policy_attachment.bedrock_models_access,
  ]
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_S3VectorsByIndexARN(rName, model string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.test.index_arn
    }
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_S3VectorsByIndexName(rName, model string) string {
	return acctest.ConfigCompose(acctest.ConfigBedrockAgentKnowledgeBaseS3VectorsBase(rName), fmt.Sprintf(`
resource "aws_bedrockagent_knowledge_base" "test" {
  depends_on = [
    aws_iam_role_policy.test,
  ]

  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/%[2]s"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_name        = aws_s3vectors_index.test.index_name
      vector_bucket_arn = aws_s3vectors_vector_bucket.test.vector_bucket_arn
    }
  }
}
`, rName, model))
}

func testAccKnowledgeBaseConfig_StructuredDataStore_redshiftProvisioned(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s-bedrock"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "RedshiftDataAPIStatementPermissions"
        Effect = "Allow"
        Action = [
          "redshift-data:GetStatementResult",
          "redshift-data:DescribeStatement",
          "redshift-data:CancelStatement"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "RedshiftDataAPIExecutePermissions"
        Effect = "Allow"
        Action = [
          "redshift-data:ExecuteStatement"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "SqlWorkbenchAccess"
        Effect = "Allow"
        Action = [
          "sqlworkbench:GetSqlRecommendations",
          "sqlworkbench:PutSqlGenerationContext",
          "sqlworkbench:GetSqlGenerationContext",
          "sqlworkbench:DeleteSqlGenerationContext"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "GenerateQueryAccess"
        Effect = "Allow"
        Action = [
          "bedrock:GenerateQuery"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "GetCredentialsWithClusterCredentials"
        Effect = "Allow"
        Action = [
          "redshift:GetClusterCredentials"
        ]
        Resource = ["*"]
      }
    ]
  })
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier    = %[1]q
  database_name         = "mydb"
  master_username       = "foo_test"
  master_password       = "Mustbe8characters"
  node_type             = "ra3.large"
  allow_version_upgrade = false
  skip_final_snapshot   = true
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "SQL"

    sql_knowledge_base_configuration {
      type = "REDSHIFT"

      redshift_configuration {
        query_engine_configuration {
          type = "PROVISIONED"

          provisioned_configuration {
            cluster_identifier = aws_redshift_cluster.test.cluster_identifier

            auth_configuration {
              type          = "USERNAME"
              database_user = aws_redshift_cluster.test.master_username
            }
          }
        }

        storage_configuration {
          type = "REDSHIFT"

          redshift_configuration {
            database_name = aws_redshift_cluster.test.database_name
          }
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}

func testAccKnowledgeBaseConfig_StructuredDataStore_redshiftServerless(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s-bedrock"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "SqlWorkbenchAccess"
        Effect = "Allow"
        Action = [
          "sqlworkbench:GetSqlRecommendations",
          "sqlworkbench:PutSqlGenerationContext",
          "sqlworkbench:GetSqlGenerationContext",
          "sqlworkbench:DeleteSqlGenerationContext"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "RedshiftServerlessGetCredentials"
        Effect = "Allow"
        Action = [
          "redshift-serverless:GetClusterCredentials"
        ]
        Resource = ["*"]
      }
    ]
  })
}

resource "aws_redshiftserverless_namespace" "test" {
  namespace_name        = %[1]q
  manage_admin_password = true
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "SQL"

    sql_knowledge_base_configuration {
      type = "REDSHIFT"

      redshift_configuration {
        query_engine_configuration {
          type = "SERVERLESS"

          serverless_configuration {
            workgroup_arn = aws_redshiftserverless_workgroup.test.arn

            auth_configuration {
              type                         = "USERNAME_PASSWORD"
              username_password_secret_arn = aws_redshiftserverless_namespace.test.admin_password_secret_arn
            }
          }
        }

        storage_configuration {
          type = "REDSHIFT"

          redshift_configuration {
            database_name = aws_redshiftserverless_namespace.test.db_name
          }
        }

        query_generation_configuration {
          generation_context {
            curated_query {
              natural_language = "Find all the things"
              sql              = "SELECT * FROM things;"
            }
          }
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}

func testAccKnowledgeBaseConfig_NeptuneAnalytics_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "bedrock.amazonaws.com"
        }
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:knowledge-base/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = "%[1]s-bedrock"
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowAllPermissionsForRDS"
        Effect = "Allow"
        Action = [
          "rds:*"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "AllowDataAccessForNeptune"
        Effect = "Allow"
        Action = [
          "neptune-db:*"
        ]
        Resource = ["*"]
      },
      {
        Sid    = "AllowAllPermissionsForNeptuneGraph"
        Effect = "Allow"
        Action = [
          "neptune-graph:*"
        ]
        Resource = ["*"]
      }
    ]
  })
}

resource "aws_neptunegraph_graph" "test" {
  graph_name          = %[1]q
  provisioned_memory  = 16
  public_connectivity = false
  replica_count       = 0
  deletion_protection = false

  vector_search_configuration {
    vector_search_dimension = 1024
  }
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/cohere.embed-english-v3"
    }
  }

  storage_configuration {
    type = "NEPTUNE_ANALYTICS"

    neptune_analytics_configuration {
      graph_arn = aws_neptunegraph_graph.test.arn

      field_mapping {
        metadata_field = "metadata"
        text_field     = "text"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName)
}
