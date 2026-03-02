// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3Object_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"/"+rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"/"+rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccS3Object_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"/"+rName+"-0")),
					querycheck.ExpectResourceKnownValues("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("acl"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), tfknownvalue.GlobalARNNoAccountIDExact("s3", rName+"/"+rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("bucket_key_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("cache_control"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_algorithm"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_crc32"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_crc32c"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_crc64nvme"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_sha1"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("checksum_sha256"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrContent), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("content_base64"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("content_disposition"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("content_encoding"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("content_language"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrContentType), knownvalue.StringExact("application/octet-stream")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("etag"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrForceDestroy), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrKMSKeyID), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("metadata"), knownvalue.MapSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("object_lock_legal_hold_status"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("object_lock_mode"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("object_lock_retain_until_date"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("override_provider"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("server_side_encryption"), tfknownvalue.StringExact(awstypes.ServerSideEncryptionAes256)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrSource), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("source_hash"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrStorageClass), tfknownvalue.StringExact(awstypes.ObjectStorageClassStandard)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("version_id"), knownvalue.StringExact("")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("website_redirect"), knownvalue.StringExact("")),
					}),
				},
			},
		},
	})
}

func TestAccS3Object_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:   acctest.ErrorCheck(t, names.S3ServiceID),
		CheckDestroy: testAccCheckObjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Object/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.AlternateRegion())),
				},
			},

			// Step 2: Query
			{
				Query:                    true,
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				ConfigDirectory:          config.StaticDirectory("testdata/Object/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity2.Checks()),
				},
			},
		},
	})
}

func TestAccS3Object_List_directoryBucket(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(directoryBucketFullNameRegex(rName))),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringRegexp(directoryBucketFullNameRegex(rName))),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrKey), knownvalue.StringExact(rName+"-1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_directory_bucket/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+tfs3.DirectoryBucketNameSuffixRegexPattern+"/"+rName+"-0"+`$`))),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+tfs3.DirectoryBucketNameSuffixRegexPattern+"/"+rName+"-1"+`$`))),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccS3Object_List_prefix(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_s3_object.test[0]"
	resourceName2 := "aws_s3_object.test[1]"
	resourceName3 := "aws_s3_object.other[0]"
	resourceName4 := "aws_s3_object.other[1]"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()
	identity3 := tfstatecheck.Identity()
	identity4 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckObjectDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("prefix-"+rName+"-0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("prefix-"+rName+"-1")),

					identity3.GetIdentity(resourceName3),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName3, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("other-"+rName+"-0")),

					identity4.GetIdentity(resourceName4),
					statecheck.ExpectKnownValue(resourceName4, tfjsonpath.New(names.AttrBucket), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName4, tfjsonpath.New(names.AttrKey), knownvalue.StringExact("other-"+rName+"-1")),
				},
			},

			// Step 2: Query only for prefix- objects
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Object/list_prefix/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.StringExact(rName+"/prefix-"+rName+"-0")),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_s3_object.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.StringExact(rName+"/prefix-"+rName+"-1")),
					tfquerycheck.ExpectNoResourceObject("aws_s3_object.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),

					tfquerycheck.ExpectNoIdentityFunc("aws_s3_object.test", identity3.Checks()),

					tfquerycheck.ExpectNoIdentityFunc("aws_s3_object.test", identity4.Checks()),
				},
			},
		},
	})
}
