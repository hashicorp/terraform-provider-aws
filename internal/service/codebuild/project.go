// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_project", name="Project")
// @Tags
func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectUpdate,
		DeleteWithoutTimeout: resourceProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifacts": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"artifact_identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bucket_owner_access": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.BucketOwnerAccess](),
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == d.Get(names.AttrName) && new == "" {
									return true
								}
								return false
							},
						},
						"namespace_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if artifactType := types.ArtifactsType(d.Get("artifacts.0.type").(string)); artifactType == types.ArtifactsTypeS3 {
									return types.ArtifactNamespace(old) == types.ArtifactNamespaceNone && new == ""
								}
								return old == new
							},
							ValidateDiagFunc: enum.Validate[types.ArtifactNamespace](),
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"packaging": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								switch artifactType := types.ArtifactsType(d.Get("artifacts.0.type").(string)); artifactType {
								case types.ArtifactsTypeCodepipeline:
									return new == ""
								case types.ArtifactsTypeS3:
									return types.ArtifactPackaging(old) == types.ArtifactPackagingNone && new == ""
								default:
									return old == new
								}
							},
							ValidateDiagFunc: enum.Validate[types.ArtifactPackaging](),
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ArtifactsType](),
						},
					},
				},
			},
			"badge_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"badge_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"build_batch_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"combine_artifacts": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"restrictions": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compute_types_allowed": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: enum.Validate[types.ComputeType](),
										},
									},
									"maximum_builds_allowed": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 100),
									},
								},
							},
						},
						names.AttrServiceRole: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"timeout_in_mins": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(5, 2160),
						},
					},
				},
			},
			"build_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(5, 2160),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					switch environmentType := types.EnvironmentType(d.Get("environment.0.type").(string)); environmentType {
					case types.EnvironmentTypeArmLambdaContainer, types.EnvironmentTypeLinuxLambdaContainer:
						return true
					default:
						return old == new
					}
				},
			},
			"cache": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"modes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.CacheMode](),
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.CacheTypeNoCache,
							ValidateDiagFunc: enum.Validate[types.CacheType](),
						},
					},
				},
			},
			"concurrent_build_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"encryption_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrEnvironment: {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCertificate: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`\.(pem|zip)$`), "must end in .pem or .zip"),
						},
						"compute_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ComputeType](),
						},
						"environment_variable": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.EnvironmentVariableTypePlaintext,
										ValidateDiagFunc: enum.Validate[types.EnvironmentVariableType](),
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
						},
						"image_pull_credentials_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ImagePullCredentialsTypeCodebuild,
							ValidateDiagFunc: enum.Validate[types.ImagePullCredentialsType](),
						},
						"privileged_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"registry_credential": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"credential": {
										Type:     schema.TypeString,
										Required: true,
									},
									"credential_provider": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.CredentialProviderType](),
									},
								},
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.EnvironmentType](),
						},
					},
				},
			},
			"file_system_locations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIdentifier: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mount_options": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mount_point": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.FileSystemTypeEfs,
							ValidateDiagFunc: enum.Validate[types.FileSystemType](),
						},
					},
				},
			},
			"logs_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCloudWatchLogs: {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrGroupName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.LogsConfigStatusTypeEnabled,
										ValidateDiagFunc: enum.Validate[types.LogsConfigStatusType](),
									},
									"stream_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
						"s3_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_owner_access": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.BucketOwnerAccess](),
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									names.AttrLocation: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validProjectS3LogsLocation,
									},
									names.AttrStatus: {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.LogsConfigStatusTypeDisabled,
										ValidateDiagFunc: enum.Validate[types.LogsConfigStatusType](),
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidProjectName,
			},
			"project_visibility": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          types.ProjectVisibilityTypePrivate,
				ValidateDiagFunc: enum.Validate[types.ProjectVisibilityType](),
			},
			"public_project_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"queued_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      480,
				ValidateFunc: validation.IntBetween(5, 480),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					switch environmentType := types.EnvironmentType(d.Get("environment.0.type").(string)); environmentType {
					case types.EnvironmentTypeArmLambdaContainer, types.EnvironmentTypeLinuxLambdaContainer:
						return true
					default:
						return old == new
					}
				},
			},
			"resource_access_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"secondary_artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 12,
				Set:      resourceProjectArtifactsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"artifact_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket_owner_access": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[types.BucketOwnerAccess](),
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace_type": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ArtifactNamespaceNone,
							ValidateDiagFunc: enum.Validate[types.ArtifactNamespace](),
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"packaging": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.ArtifactPackagingNone,
							ValidateDiagFunc: enum.Validate[types.ArtifactPackaging](),
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ArtifactsType](),
						},
					},
				},
			},
			"secondary_sources": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 12,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"build_status_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"context": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"target_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"git_clone_depth": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"git_submodules_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fetch_submodules": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"insecure_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"source_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.SourceType](),
						},
					},
				},
			},
			"secondary_source_version": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 12,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"source_version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrServiceRole: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrSource: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"build_status_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"context": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"target_url": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"git_clone_depth": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"git_submodules_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"fetch_submodules": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"insecure_ssl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						names.AttrLocation: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.SourceType](),
						},
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"source_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCConfig: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MaxItems: 16,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MaxItems: 5,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for cache location
				cacheType, cacheTypeOk := diff.GetOk("cache.0.type")
				if !cacheTypeOk || types.CacheType(cacheType.(string)) == types.CacheTypeNoCache || types.CacheType(cacheType.(string)) == types.CacheTypeLocal {
					return nil
				}
				if v, ok := diff.GetOk("cache.0.location"); ok && v.(string) != "" {
					return nil
				}
				if !diff.NewValueKnown("cache.0.location") {
					// value may be computed - don't assume it isn't set
					return nil
				}
				return fmt.Errorf(`cache location is required when cache type is %q`, cacheType.(string))
			},
			verify.SetTagsDiff,
		),
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	var projectSource *types.ProjectSource
	if v, ok := d.GetOk(names.AttrSource); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		projectSource = expandProjectSource(v.([]interface{})[0].(map[string]interface{}))
	}

	if projectSource != nil && projectSource.Type == types.SourceTypeNoSource {
		if aws.ToString(projectSource.Buildspec) == "" {
			return sdkdiag.AppendErrorf(diags, "`buildspec` must be set when source's `type` is `NO_SOURCE`")
		}

		if aws.ToString(projectSource.Location) != "" {
			return sdkdiag.AppendErrorf(diags, "`location` must be empty when source's `type` is `NO_SOURCE`")
		}
	}

	name := d.Get(names.AttrName).(string)
	input := &codebuild.CreateProjectInput{
		LogsConfig: expandProjectLogsConfig(d.Get("logs_config")),
		Name:       aws.String(name),
		Source:     projectSource,
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("artifacts"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Artifacts = expandProjectArtifacts(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("badge_enabled"); ok {
		input.BadgeEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("build_batch_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.BuildBatchConfig = expandProjectBuildBatchConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cache"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Cache = expandProjectCache(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("concurrent_build_limit"); ok {
		input.ConcurrentBuildLimit = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		input.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEnvironment); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Environment = expandProjectEnvironment(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("file_system_locations"); ok && v.(*schema.Set).Len() > 0 {
		input.FileSystemLocations = expandProjectFileSystemLocations(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("queued_timeout"); ok {
		input.QueuedTimeoutInMinutes = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("secondary_artifacts"); ok && v.(*schema.Set).Len() > 0 {
		input.SecondaryArtifacts = expandProjectSecondaryArtifacts(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("secondary_sources"); ok && v.(*schema.Set).Len() > 0 {
		input.SecondarySources = expandProjectSecondarySources(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("secondary_source_version"); ok && v.(*schema.Set).Len() > 0 {
		input.SecondarySourceVersions = expandProjectSecondarySourceVersions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrServiceRole); ok {
		input.ServiceRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_version"); ok {
		input.SourceVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_timeout"); ok {
		input.TimeoutInMinutes = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.VpcConfig = expandVPCConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	// InvalidInputException: CodeBuild is not authorized to perform
	// InvalidInputException: Not authorized to perform DescribeSecurityGroups
	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidInputException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateProject(ctx, input)
	}, "ot authorized to perform")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Project (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*codebuild.CreateProjectOutput).Project.Arn))

	if v, ok := d.GetOk("project_visibility"); ok {
		if v := types.ProjectVisibilityType(v.(string)); v != types.ProjectVisibilityTypePrivate {
			input := &codebuild.UpdateProjectVisibilityInput{
				ProjectArn:        aws.String(d.Id()),
				ProjectVisibility: v,
			}

			if v, ok := d.GetOk("resource_access_role"); ok {
				input.ResourceAccessRole = aws.String(v.(string))
			}

			_, err = conn.UpdateProjectVisibility(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating CodeBuild Project (%s) visibility: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	project, err := findProjectByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CodeBuild Project (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, project.Arn)
	if project.Artifacts != nil {
		if err := d.Set("artifacts", []interface{}{flattenProjectArtifacts(project.Artifacts)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting artifacts: %s", err)
		}
	} else {
		d.Set("artifacts", nil)
	}
	if project.Badge != nil {
		d.Set("badge_enabled", project.Badge.BadgeEnabled)
		d.Set("badge_url", project.Badge.BadgeRequestUrl)
	} else {
		d.Set("badge_enabled", false)
		d.Set("badge_url", "")
	}
	if err := d.Set("build_batch_config", flattenBuildBatchConfig(project.BuildBatchConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting build_batch_config: %s", err)
	}
	d.Set("build_timeout", project.TimeoutInMinutes)
	if err := d.Set("cache", flattenProjectCache(project.Cache)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache: %s", err)
	}
	d.Set("concurrent_build_limit", project.ConcurrentBuildLimit)
	d.Set(names.AttrDescription, project.Description)
	d.Set("encryption_key", project.EncryptionKey)
	if err := d.Set(names.AttrEnvironment, flattenProjectEnvironment(project.Environment)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting environment: %s", err)
	}
	if err := d.Set("file_system_locations", flattenProjectFileSystemLocations(project.FileSystemLocations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting file_system_locations: %s", err)
	}
	if err := d.Set("logs_config", flattenLogsConfig(project.LogsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logs_config: %s", err)
	}
	d.Set(names.AttrName, project.Name)
	if v := project.ProjectVisibility; v != "" {
		d.Set("project_visibility", project.ProjectVisibility)
	} else {
		d.Set("project_visibility", types.ProjectVisibilityTypePrivate)
	}
	d.Set("public_project_alias", project.PublicProjectAlias)
	d.Set("resource_access_role", project.ResourceAccessRole)
	d.Set("queued_timeout", project.QueuedTimeoutInMinutes)
	if err := d.Set("secondary_artifacts", flattenProjectSecondaryArtifacts(project.SecondaryArtifacts)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting secondary_artifacts: %s", err)
	}
	if err := d.Set("secondary_sources", flattenProjectSecondarySources(project.SecondarySources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting secondary_sources: %s", err)
	}
	if err := d.Set("secondary_source_version", flattenProjectSecondarySourceVersions(project.SecondarySourceVersions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting secondary_source_version: %s", err)
	}
	d.Set(names.AttrServiceRole, project.ServiceRole)
	if project.Source != nil {
		if err := d.Set(names.AttrSource, []interface{}{flattenProjectSource(project.Source)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting source: %s", err)
		}
	} else {
		d.Set(names.AttrSource, nil)
	}
	d.Set("source_version", project.SourceVersion)
	if err := d.Set(names.AttrVPCConfig, flattenVPCConfig(project.VpcConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting vpc_config: %s", err)
	}

	setTagsOut(ctx, project.Tags)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	if d.HasChanges("project_visibility", "resource_access_role") {
		input := &codebuild.UpdateProjectVisibilityInput{
			ProjectArn:        aws.String(d.Id()),
			ProjectVisibility: types.ProjectVisibilityType(d.Get("project_visibility").(string)),
		}

		if v, ok := d.GetOk("resource_access_role"); ok {
			input.ResourceAccessRole = aws.String(v.(string))
		}

		_, err := conn.UpdateProjectVisibility(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeBuild Project (%s) visibility: %s", d.Id(), err)
		}
	}

	if d.HasChangesExcept("project_visibility", "resource_access_role") {
		input := &codebuild.UpdateProjectInput{
			Name: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange("artifacts") {
			if v, ok := d.GetOk("artifacts"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Artifacts = expandProjectArtifacts(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("badge_enabled") {
			input.BadgeEnabled = aws.Bool(d.Get("badge_enabled").(bool))
		}

		if d.HasChange("build_batch_config") {
			if v, ok := d.GetOk("build_batch_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.BuildBatchConfig = expandProjectBuildBatchConfig(v.([]interface{})[0].(map[string]interface{}))
			} else {
				input.BuildBatchConfig = &types.ProjectBuildBatchConfig{}
			}
		}

		if d.HasChange("cache") {
			if v, ok := d.GetOk("cache"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Cache = expandProjectCache(v.([]interface{})[0].(map[string]interface{}))
			} else {
				input.Cache = &types.ProjectCache{
					Type: types.CacheTypeNoCache,
				}
			}
		}

		if d.HasChange("concurrent_build_limit") {
			if v := int32(d.Get("concurrent_build_limit").(int)); v != 0 {
				input.ConcurrentBuildLimit = aws.Int32(v)
			} else {
				input.ConcurrentBuildLimit = aws.Int32(-1)
			}
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("encryption_key") {
			input.EncryptionKey = aws.String(d.Get("encryption_key").(string))
		}

		if d.HasChange(names.AttrEnvironment) {
			if v, ok := d.GetOk(names.AttrEnvironment); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Environment = expandProjectEnvironment(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("file_system_locations") {
			if v, ok := d.GetOk("file_system_locations"); ok && v.(*schema.Set).Len() > 0 {
				input.FileSystemLocations = expandProjectFileSystemLocations(v.(*schema.Set).List())
			}
		}

		if d.HasChange("logs_config") {
			input.LogsConfig = expandProjectLogsConfig(d.Get("logs_config"))
		}

		if d.HasChange("queued_timeout") {
			input.QueuedTimeoutInMinutes = aws.Int32(int32(d.Get("queued_timeout").(int)))
		}

		if d.HasChange("secondary_artifacts") {
			if v, ok := d.GetOk("secondary_artifacts"); ok && v.(*schema.Set).Len() > 0 {
				input.SecondaryArtifacts = expandProjectSecondaryArtifacts(v.(*schema.Set).List())
			} else {
				input.SecondaryArtifacts = []types.ProjectArtifacts{}
			}
		}

		if d.HasChange("secondary_sources") {
			if v, ok := d.GetOk("secondary_sources"); ok && v.(*schema.Set).Len() > 0 {
				input.SecondarySources = expandProjectSecondarySources(v.(*schema.Set).List())
			} else {
				input.SecondarySources = []types.ProjectSource{}
			}
		}

		if d.HasChange("secondary_source_version") {
			if v, ok := d.GetOk("secondary_source_version"); ok && v.(*schema.Set).Len() > 0 {
				input.SecondarySourceVersions = expandProjectSecondarySourceVersions(v.(*schema.Set).List())
			} else {
				input.SecondarySourceVersions = []types.ProjectSourceVersion{}
			}
		}

		if d.HasChange(names.AttrServiceRole) {
			input.ServiceRole = aws.String(d.Get(names.AttrServiceRole).(string))
		}

		if d.HasChange(names.AttrSource) {
			if v, ok := d.GetOk(names.AttrSource); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.Source = expandProjectSource(v.([]interface{})[0].(map[string]interface{}))
			}
		}

		if d.HasChange("source_version") {
			input.SourceVersion = aws.String(d.Get("source_version").(string))
		}

		if d.HasChange("build_timeout") {
			input.TimeoutInMinutes = aws.Int32(int32(d.Get("build_timeout").(int)))
		}

		if d.HasChange(names.AttrVPCConfig) {
			if v, ok := d.GetOk(names.AttrVPCConfig); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.VpcConfig = expandVPCConfig(v.([]interface{})[0].(map[string]interface{}))
			} else {
				input.VpcConfig = &types.VpcConfig{}
			}
		}

		// The documentation clearly says "The replacement set of tags for this build project."
		// But its a slice of pointers so if not set for every update, they get removed.
		input.Tags = getTagsIn(ctx)

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidInputException](ctx, propagationTimeout, func() (interface{}, error) {
			return conn.UpdateProject(ctx, input)
		}, "ot authorized to perform")

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CodeBuild Project (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildClient(ctx)

	log.Printf("[INFO] Deleting CodeBuild Project: %s", d.Id())
	_, err := conn.DeleteProject(ctx, &codebuild.DeleteProjectInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Project (%s): %s", d.Id(), err)
	}

	return diags
}

func findProjectByNameOrARN(ctx context.Context, conn *codebuild.Client, nameOrARN string) (*types.Project, error) {
	input := &codebuild.BatchGetProjectsInput{
		Names: []string{nameOrARN},
	}

	return findProject(ctx, conn, input)
}

func findProject(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetProjectsInput) (*types.Project, error) {
	output, err := findProjects(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findProjects(ctx context.Context, conn *codebuild.Client, input *codebuild.BatchGetProjectsInput) ([]types.Project, error) {
	output, err := conn.BatchGetProjects(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Projects, nil
}

func expandProjectSecondarySourceVersions(tfList []interface{}) []types.ProjectSourceVersion {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.ProjectSourceVersion, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandProjectSourceVersion(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandProjectSourceVersion(tfMap map[string]interface{}) *types.ProjectSourceVersion {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ProjectSourceVersion{
		SourceIdentifier: aws.String(tfMap["source_identifier"].(string)),
		SourceVersion:    aws.String(tfMap["source_version"].(string)),
	}

	return apiObject
}

func expandProjectFileSystemLocations(tfList []interface{}) []types.ProjectFileSystemLocation {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.ProjectFileSystemLocation, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandProjectFileSystemLocation(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandProjectFileSystemLocation(tfMap map[string]interface{}) *types.ProjectFileSystemLocation {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ProjectFileSystemLocation{
		Type: types.FileSystemType(tfMap[names.AttrType].(string)),
	}

	if tfMap[names.AttrIdentifier].(string) != "" {
		apiObject.Identifier = aws.String(tfMap[names.AttrIdentifier].(string))
	}

	if tfMap[names.AttrLocation].(string) != "" {
		apiObject.Location = aws.String(tfMap[names.AttrLocation].(string))
	}

	if tfMap["mount_options"].(string) != "" {
		apiObject.MountOptions = aws.String(tfMap["mount_options"].(string))
	}

	if tfMap["mount_point"].(string) != "" {
		apiObject.MountPoint = aws.String(tfMap["mount_point"].(string))
	}

	return apiObject
}

func expandProjectSecondaryArtifacts(tfList []interface{}) []types.ProjectArtifacts {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.ProjectArtifacts, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandProjectArtifacts(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandProjectArtifacts(tfMap map[string]interface{}) *types.ProjectArtifacts {
	if tfMap == nil {
		return nil
	}

	artifactType := types.ArtifactsType(tfMap[names.AttrType].(string))
	apiObject := &types.ProjectArtifacts{
		Type: artifactType,
	}

	// Only valid for S3 and CODEPIPELINE artifacts types
	// InvalidInputException: Invalid artifacts: artifact type NO_ARTIFACTS should have null encryptionDisabled
	if artifactType == types.ArtifactsTypeS3 || artifactType == types.ArtifactsTypeCodepipeline {
		apiObject.EncryptionDisabled = aws.Bool(tfMap["encryption_disabled"].(bool))
	}

	if v, ok := tfMap["artifact_identifier"].(string); ok && v != "" {
		apiObject.ArtifactIdentifier = aws.String(v)
	}

	if v, ok := tfMap["bucket_owner_access"].(string); ok && v != "" {
		apiObject.BucketOwnerAccess = types.BucketOwnerAccess(v)
	}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["namespace_type"].(string); ok && v != "" {
		apiObject.NamespaceType = types.ArtifactNamespace(v)
	}

	if v, ok := tfMap["override_artifact_name"]; ok {
		apiObject.OverrideArtifactName = aws.Bool(v.(bool))
	}

	if v, ok := tfMap["packaging"].(string); ok && v != "" {
		apiObject.Packaging = types.ArtifactPackaging(v)
	}

	if v, ok := tfMap[names.AttrPath].(string); ok && v != "" {
		apiObject.Path = aws.String(v)
	}

	return apiObject
}

func expandProjectCache(tfMap map[string]interface{}) *types.ProjectCache {
	if tfMap == nil {
		return nil
	}

	cacheType := types.CacheType(tfMap[names.AttrType].(string))
	apiObject := &types.ProjectCache{
		Type: cacheType,
	}

	if v, ok := tfMap[names.AttrLocation]; ok {
		apiObject.Location = aws.String(v.(string))
	}

	if cacheType == types.CacheTypeLocal {
		if v, ok := tfMap["modes"].([]interface{}); ok && len(v) > 0 {
			apiObject.Modes = flex.ExpandStringyValueList[types.CacheMode](v)
		}
	}

	return apiObject
}

func expandProjectEnvironment(tfMap map[string]interface{}) *types.ProjectEnvironment {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ProjectEnvironment{
		PrivilegedMode: aws.Bool(tfMap["privileged_mode"].(bool)),
	}

	if v, ok := tfMap[names.AttrCertificate].(string); ok && v != "" {
		apiObject.Certificate = aws.String(v)
	}

	if v, ok := tfMap["compute_type"].(string); ok && v != "" {
		apiObject.ComputeType = types.ComputeType(v)
	}

	if v, ok := tfMap["image"].(string); ok && v != "" {
		apiObject.Image = aws.String(v)
	}

	if v, ok := tfMap["image_pull_credentials_type"].(string); ok && v != "" {
		apiObject.ImagePullCredentialsType = types.ImagePullCredentialsType(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = types.EnvironmentType(v)
	}

	if v, ok := tfMap["registry_credential"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]interface{})

		projectRegistryCredential := &types.RegistryCredential{}

		if v, ok := tfMap["credential"]; ok && v.(string) != "" {
			projectRegistryCredential.Credential = aws.String(v.(string))
		}

		if v, ok := tfMap["credential_provider"]; ok && v.(string) != "" {
			projectRegistryCredential.CredentialProvider = types.CredentialProviderType(v.(string))
		}

		apiObject.RegistryCredential = projectRegistryCredential
	}

	if v, ok := tfMap["environment_variable"].([]interface{}); ok && len(v) > 0 {
		projectEnvironmentVariables := make([]types.EnvironmentVariable, 0)

		for _, tfMapRaw := range v {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			projectEnvironmentVar := types.EnvironmentVariable{}

			if v := tfMap[names.AttrName].(string); v != "" {
				projectEnvironmentVar.Name = aws.String(v)
			}

			if v := tfMap[names.AttrType].(string); v != "" {
				projectEnvironmentVar.Type = types.EnvironmentVariableType(v)
			}

			if v, ok := tfMap[names.AttrValue].(string); ok {
				projectEnvironmentVar.Value = aws.String(v)
			}

			projectEnvironmentVariables = append(projectEnvironmentVariables, projectEnvironmentVar)
		}

		apiObject.EnvironmentVariables = projectEnvironmentVariables
	}

	return apiObject
}

func expandProjectLogsConfig(v interface{}) *types.LogsConfig {
	apiObject := &types.LogsConfig{}

	if v, ok := v.([]interface{}); ok && len(v) > 0 && v[0] != nil {
		if tfMap := v[0].(map[string]interface{}); tfMap != nil {
			if v, ok := tfMap[names.AttrCloudWatchLogs].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				apiObject.CloudWatchLogs = expandCloudWatchLogsConfig(v[0].(map[string]interface{}))
			}

			if v, ok := tfMap["s3_logs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
				apiObject.S3Logs = expandS3LogsConfig(v[0].(map[string]interface{}))
			}
		}
	}

	if apiObject.CloudWatchLogs == nil {
		apiObject.CloudWatchLogs = &types.CloudWatchLogsConfig{
			Status: types.LogsConfigStatusTypeEnabled,
		}
	}

	if apiObject.S3Logs == nil {
		apiObject.S3Logs = &types.S3LogsConfig{
			Status: types.LogsConfigStatusTypeDisabled,
		}
	}

	return apiObject
}

func expandCloudWatchLogsConfig(tfMap map[string]interface{}) *types.CloudWatchLogsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.CloudWatchLogsConfig{
		Status: types.LogsConfigStatusType(tfMap[names.AttrStatus].(string)),
	}

	if v, ok := tfMap[names.AttrGroupName].(string); ok && v != "" {
		apiObject.GroupName = aws.String(v)
	}

	if v, ok := tfMap["stream_name"].(string); ok && v != "" {
		apiObject.StreamName = aws.String(v)
	}

	return apiObject
}

func expandS3LogsConfig(tfMap map[string]interface{}) *types.S3LogsConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.S3LogsConfig{
		EncryptionDisabled: aws.Bool(tfMap["encryption_disabled"].(bool)),
		Status:             types.LogsConfigStatusType(tfMap[names.AttrStatus].(string)),
	}

	if v, ok := tfMap["bucket_owner_access"].(string); ok && v != "" {
		apiObject.BucketOwnerAccess = types.BucketOwnerAccess(v)
	}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	return apiObject
}

func expandProjectBuildBatchConfig(tfMap map[string]interface{}) *types.ProjectBuildBatchConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.ProjectBuildBatchConfig{
		ServiceRole: aws.String(tfMap[names.AttrServiceRole].(string)),
	}

	if v, ok := tfMap["combine_artifacts"].(bool); ok {
		apiObject.CombineArtifacts = aws.Bool(v)
	}

	if v, ok := tfMap["restrictions"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Restrictions = expandBatchRestrictions(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["timeout_in_mins"].(int); ok && v != 0 {
		apiObject.TimeoutInMins = aws.Int32(int32(v))
	}

	return apiObject
}

func expandBatchRestrictions(tfMap map[string]interface{}) *types.BatchRestrictions {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.BatchRestrictions{}

	if v, ok := tfMap["compute_types_allowed"].([]interface{}); ok && len(v) > 0 {
		apiObject.ComputeTypesAllowed = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["maximum_builds_allowed"].(int); ok && v != 0 {
		apiObject.MaximumBuildsAllowed = aws.Int32(int32(v))
	}

	return apiObject
}

func expandVPCConfig(tfMap map[string]interface{}) *types.VpcConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.VpcConfig{
		SecurityGroupIds: flex.ExpandStringValueSet(tfMap[names.AttrSecurityGroupIDs].(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(tfMap[names.AttrSubnets].(*schema.Set)),
		VpcId:            aws.String(tfMap[names.AttrVPCID].(string)),
	}

	return apiObject
}

func expandProjectSecondarySources(tfList []interface{}) []types.ProjectSource {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]types.ProjectSource, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandProjectSource(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandProjectSource(tfMap map[string]interface{}) *types.ProjectSource {
	if tfMap == nil {
		return nil
	}

	sourceType := types.SourceType(tfMap[names.AttrType].(string))
	apiObject := &types.ProjectSource{
		Buildspec:     aws.String(tfMap["buildspec"].(string)),
		GitCloneDepth: aws.Int32(int32(tfMap["git_clone_depth"].(int))),
		InsecureSsl:   aws.Bool(tfMap["insecure_ssl"].(bool)),
		Type:          sourceType,
	}

	if v, ok := tfMap[names.AttrLocation].(string); ok && v != "" {
		apiObject.Location = aws.String(v)
	}

	if v, ok := tfMap["source_identifier"].(string); ok && v != "" {
		apiObject.SourceIdentifier = aws.String(v)
	}

	// Only valid for BITBUCKET, GITHUB, GITHUB_ENTERPRISE, GITLAB, and GITLAB_SELF_MANAGED source types
	// e.g., InvalidInputException: Source type NO_SOURCE does not support ReportBuildStatus
	if sourceType == types.SourceTypeBitbucket || sourceType == types.SourceTypeGithub || sourceType == types.SourceTypeGithubEnterprise || sourceType == types.SourceTypeGitlab || sourceType == types.SourceTypeGitlabSelfManaged {
		apiObject.ReportBuildStatus = aws.Bool(tfMap["report_build_status"].(bool))
	}

	// Only valid for BITBUCKET, CODECOMMIT, GITHUB, and GITHUB_ENTERPRISE source types
	if sourceType == types.SourceTypeBitbucket || sourceType == types.SourceTypeCodecommit || sourceType == types.SourceTypeGithub || sourceType == types.SourceTypeGithubEnterprise {
		if v, ok := tfMap["git_submodules_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			gitSubmodulesConfig := &types.GitSubmodulesConfig{}

			if v, ok := tfMap["fetch_submodules"].(bool); ok {
				gitSubmodulesConfig.FetchSubmodules = aws.Bool(v)
			}

			apiObject.GitSubmodulesConfig = gitSubmodulesConfig
		}
	}

	// Only valid for BITBUCKET, GITHUB, GITHUB_ENTERPRISE, GITLAB, and GITLAB_SELF_MANAGED source types
	if sourceType == types.SourceTypeBitbucket || sourceType == types.SourceTypeGithub || sourceType == types.SourceTypeGithubEnterprise || sourceType == types.SourceTypeGitlab || sourceType == types.SourceTypeGitlabSelfManaged {
		if v, ok := tfMap["build_status_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})

			buildStatusConfig := &types.BuildStatusConfig{}

			if v, ok := tfMap["context"].(string); ok && v != "" {
				buildStatusConfig.Context = aws.String(v)
			}
			if v, ok := tfMap["target_url"].(string); ok && v != "" {
				buildStatusConfig.TargetUrl = aws.String(v)
			}

			apiObject.BuildStatusConfig = buildStatusConfig
		}
	}

	return apiObject
}

func flattenProjectFileSystemLocations(apiObjects []types.ProjectFileSystemLocation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenProjectFileSystemLocation(apiObject))
	}

	return tfList
}

func flattenProjectFileSystemLocation(apiObject types.ProjectFileSystemLocation) map[string]interface{} {
	tfMap := map[string]interface{}{
		names.AttrType: apiObject.Type,
	}

	if v := apiObject.Identifier; v != nil {
		tfMap[names.AttrIdentifier] = aws.ToString(v)
	}

	if v := apiObject.Location; v != nil {
		tfMap[names.AttrLocation] = aws.ToString(v)
	}

	if v := apiObject.MountOptions; v != nil {
		tfMap["mount_options"] = aws.ToString(v)
	}

	if v := apiObject.MountPoint; v != nil {
		tfMap["mount_point"] = aws.ToString(v)
	}

	return tfMap
}

func flattenLogsConfig(apiObject *types.LogsConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrCloudWatchLogs: flattenCloudWatchLogs(apiObject.CloudWatchLogs),
		"s3_logs":                flattenS3Logs(apiObject.S3Logs),
	}

	return []interface{}{tfMap}
}

func flattenCloudWatchLogs(apiObject *types.CloudWatchLogsConfig) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject == nil {
		tfMap[names.AttrStatus] = types.LogsConfigStatusTypeDisabled
	} else {
		tfMap[names.AttrGroupName] = aws.ToString(apiObject.GroupName)
		tfMap[names.AttrStatus] = apiObject.Status
		tfMap["stream_name"] = aws.ToString(apiObject.StreamName)
	}

	return []interface{}{tfMap}
}

func flattenS3Logs(apiObject *types.S3LogsConfig) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject == nil {
		tfMap[names.AttrStatus] = types.LogsConfigStatusTypeDisabled
	} else {
		tfMap["bucket_owner_access"] = apiObject.BucketOwnerAccess
		tfMap["encryption_disabled"] = aws.ToBool(apiObject.EncryptionDisabled)
		tfMap[names.AttrLocation] = aws.ToString(apiObject.Location)
		tfMap[names.AttrStatus] = apiObject.Status
	}

	return []interface{}{tfMap}
}

func flattenProjectSecondaryArtifacts(apiObjects []types.ProjectArtifacts) []interface{} {
	tfList := []interface{}{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenProjectArtifacts(&apiObject))
	}
	return tfList
}

func flattenProjectArtifacts(apiObject *types.ProjectArtifacts) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"bucket_owner_access": apiObject.BucketOwnerAccess,
		"namespace_type":      apiObject.NamespaceType,
		"packaging":           apiObject.Packaging,
		names.AttrType:        apiObject.Type,
	}

	if apiObject.ArtifactIdentifier != nil {
		tfMap["artifact_identifier"] = aws.ToString(apiObject.ArtifactIdentifier)
	}

	if apiObject.EncryptionDisabled != nil {
		tfMap["encryption_disabled"] = aws.ToBool(apiObject.EncryptionDisabled)
	}

	if apiObject.Location != nil {
		tfMap[names.AttrLocation] = aws.ToString(apiObject.Location)
	}

	if apiObject.OverrideArtifactName != nil {
		tfMap["override_artifact_name"] = aws.ToBool(apiObject.OverrideArtifactName)
	}

	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}

	if apiObject.Path != nil {
		tfMap[names.AttrPath] = aws.ToString(apiObject.Path)
	}

	return tfMap
}

func resourceProjectArtifactsHash(v interface{}) int {
	var buf bytes.Buffer
	tfMap := v.(map[string]interface{})

	if v, ok := tfMap["artifact_identifier"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap["bucket_owner_access"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap["encryption_disabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := tfMap[names.AttrLocation]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap["namespace_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap["override_artifact_name"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := tfMap["packaging"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap[names.AttrPath]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := tfMap[names.AttrType]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func flattenProjectCache(apiObject *types.ProjectCache) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrLocation: aws.ToString(apiObject.Location),
		"modes":            apiObject.Modes,
		names.AttrType:     apiObject.Type,
	}

	return []interface{}{tfMap}
}

func flattenProjectEnvironment(apiObject *types.ProjectEnvironment) []interface{} {
	tfMap := map[string]interface{}{
		"compute_type":                apiObject.ComputeType,
		"image_pull_credentials_type": apiObject.ImagePullCredentialsType,
		names.AttrType:                apiObject.Type,
	}

	tfMap["image"] = aws.ToString(apiObject.Image)
	tfMap[names.AttrCertificate] = aws.ToString(apiObject.Certificate)
	tfMap["privileged_mode"] = aws.ToBool(apiObject.PrivilegedMode)
	tfMap["registry_credential"] = flattenRegistryCredential(apiObject.RegistryCredential)

	if apiObject.EnvironmentVariables != nil {
		tfMap["environment_variable"] = flattenEnvironmentVariables(apiObject.EnvironmentVariables)
	}

	return []interface{}{tfMap}
}

func flattenRegistryCredential(apiObject *types.RegistryCredential) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"credential":          aws.ToString(apiObject.Credential),
		"credential_provider": apiObject.CredentialProvider,
	}

	return []interface{}{tfMap}
}

func flattenProjectSecondarySources(apiObject []types.ProjectSource) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObject {
		tfList = append(tfList, flattenProjectSource(&apiObject))
	}

	return tfList
}

func flattenProjectSource(apiObject *types.ProjectSource) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"buildspec":           aws.ToString(apiObject.Buildspec),
		names.AttrLocation:    aws.ToString(apiObject.Location),
		"git_clone_depth":     aws.ToInt32(apiObject.GitCloneDepth),
		"insecure_ssl":        aws.ToBool(apiObject.InsecureSsl),
		"report_build_status": aws.ToBool(apiObject.ReportBuildStatus),
		names.AttrType:        apiObject.Type,
	}

	tfMap["git_submodules_config"] = flattenProjectGitSubmodulesConfig(apiObject.GitSubmodulesConfig)

	tfMap["build_status_config"] = flattenProjectBuildStatusConfig(apiObject.BuildStatusConfig)

	if apiObject.SourceIdentifier != nil {
		tfMap["source_identifier"] = aws.ToString(apiObject.SourceIdentifier)
	}

	return tfMap
}

func flattenProjectSecondarySourceVersions(apiObjects []types.ProjectSourceVersion) []interface{} {
	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenProjectSourceVersion(apiObject))
	}
	return tfList
}

func flattenProjectSourceVersion(apiObject types.ProjectSourceVersion) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if apiObject.SourceIdentifier != nil {
		tfMap["source_identifier"] = aws.ToString(apiObject.SourceIdentifier)
	}

	if apiObject.SourceVersion != nil {
		tfMap["source_version"] = aws.ToString(apiObject.SourceVersion)
	}

	return tfMap
}

func flattenProjectGitSubmodulesConfig(apiObject *types.GitSubmodulesConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"fetch_submodules": aws.ToBool(apiObject.FetchSubmodules),
	}

	return []interface{}{tfMap}
}

func flattenProjectBuildStatusConfig(apiObject *types.BuildStatusConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"context":    aws.ToString(apiObject.Context),
		"target_url": aws.ToString(apiObject.TargetUrl),
	}

	return []interface{}{tfMap}
}

func flattenVPCConfig(apiObject *types.VpcConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrVPCID] = aws.ToString(apiObject.VpcId)
	tfMap[names.AttrSubnets] = apiObject.Subnets
	tfMap[names.AttrSecurityGroupIDs] = apiObject.SecurityGroupIds

	return []interface{}{tfMap}
}

func flattenBuildBatchConfig(apiObject *types.ProjectBuildBatchConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrServiceRole] = aws.ToString(apiObject.ServiceRole)

	if apiObject.CombineArtifacts != nil {
		tfMap["combine_artifacts"] = aws.ToBool(apiObject.CombineArtifacts)
	}

	if apiObject.Restrictions != nil {
		tfMap["restrictions"] = flattenBuildBatchRestrictionsConfig(apiObject.Restrictions)
	}

	if apiObject.TimeoutInMins != nil {
		tfMap["timeout_in_mins"] = aws.ToInt32(apiObject.TimeoutInMins)
	}

	return []interface{}{tfMap}
}

func flattenBuildBatchRestrictionsConfig(apiObject *types.BatchRestrictions) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"compute_types_allowed":  apiObject.ComputeTypesAllowed,
		"maximum_builds_allowed": aws.ToInt32(apiObject.MaximumBuildsAllowed),
	}

	return []interface{}{tfMap}
}

func flattenEnvironmentVariables(apiObjects []types.EnvironmentVariable) []interface{} {
	tfList := []interface{}{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
		tfMap[names.AttrValue] = aws.ToString(apiObject.Value)
		tfMap[names.AttrType] = apiObject.Type

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func ValidProjectName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[0-9A-Za-z]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter or number", value))
	}

	if !regexache.MustCompile(`^[0-9A-Za-z_-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"only alphanumeric characters, hyphens and underscores allowed in %q", value))
	}

	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be greater than 255 characters", value))
	}

	return
}

func validProjectS3LogsLocation(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if _, errs := verify.ValidARN(v, k); len(errs) == 0 {
		errors = append(errors, errs...)
		return
	}

	simplePattern := `^[0-9a-z][^/]*\/(.+)$`
	if !regexache.MustCompile(simplePattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q does not match pattern (%q): %q",
			k, simplePattern, value))
	}

	return
}
