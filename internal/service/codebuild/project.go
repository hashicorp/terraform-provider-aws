package codebuild

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codebuild.BucketOwnerAccess_Values(), false),
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == d.Get("name") && new == "" {
									return true
								}
								return false
							},
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace_type": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if d.Get("artifacts.0.type") == codebuild.ArtifactsTypeS3 {
									return old == codebuild.ArtifactNamespaceNone && new == ""
								}
								return false
							},
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactNamespace_Values(), false),
						},
						"packaging": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								switch d.Get("artifacts.0.type") {
								case codebuild.ArtifactsTypeCodepipeline:
									return new == ""
								case codebuild.ArtifactsTypeS3:
									return old == codebuild.ArtifactPackagingNone && new == ""
								}
								return false
							},
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactPackaging_Values(), false),
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactsType_Values(), false),
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
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
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(codebuild.ComputeType_Values(), false),
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
						"service_role": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"timeout_in_mins": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(5, 480),
						},
					},
				},
			},
			"cache": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      codebuild.CacheTypeNoCache,
							ValidateFunc: validation.StringInSlice(codebuild.CacheType_Values(), false),
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"modes": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(codebuild.CacheMode_Values(), false),
							},
						},
					},
				},
			},
			"concurrent_build_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"description": {
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
			"environment": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ComputeType_Values(), false),
						},
						"environment_variable": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(codebuild.EnvironmentVariableType_Values(), false),
										Default:      codebuild.EnvironmentVariableTypePlaintext,
									},
								},
							},
						},
						"image": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.EnvironmentType_Values(), false),
						},
						"image_pull_credentials_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      codebuild.ImagePullCredentialsTypeCodebuild,
							ValidateFunc: validation.StringInSlice(codebuild.ImagePullCredentialsType_Values(), false),
						},
						"privileged_mode": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"certificate": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`\.(pem|zip)$`), "must end in .pem or .zip"),
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
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codebuild.CredentialProviderType_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"file_system_locations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"identifier": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
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
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      codebuild.FileSystemTypeEfs,
							ValidateFunc: validation.StringInSlice(codebuild.FileSystemType_Values(), false),
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
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      codebuild.LogsConfigStatusTypeEnabled,
										ValidateFunc: validation.StringInSlice(codebuild.LogsConfigStatusType_Values(), false),
									},
									"group_name": {
										Type:     schema.TypeString,
										Optional: true,
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(codebuild.BucketOwnerAccess_Values(), false),
									},
									"status": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      codebuild.LogsConfigStatusTypeDisabled,
										ValidateFunc: validation.StringInSlice(codebuild.LogsConfigStatusType_Values(), false),
									},
									"location": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validProjectS3LogsLocation,
									},
									"encryption_disabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidProjectName,
			},
			"secondary_artifacts": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 12,
				Set:      resourceProjectArtifactsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bucket_owner_access": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codebuild.BucketOwnerAccess_Values(), false),
						},
						"encryption_disabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"namespace_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactNamespace_Values(), false),
							Default:      codebuild.ArtifactNamespaceNone,
						},
						"override_artifact_name": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"packaging": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactPackaging_Values(), false),
							Default:      codebuild.ArtifactPackagingNone,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"artifact_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.ArtifactsType_Values(), false),
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
						"auth": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource": {
										Type:       schema.TypeString,
										Sensitive:  true,
										Optional:   true,
										Deprecated: "Use the aws_codebuild_source_credential resource instead",
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codebuild.SourceAuthType_Values(), false),
										Deprecated:   "Use the aws_codebuild_source_credential resource instead",
									},
								},
							},
							Deprecated: "Use the aws_codebuild_source_credential resource instead",
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.SourceType_Values(), false),
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
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"source_identifier": {
							Type:     schema.TypeString,
							Required: true,
						},
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
			"service_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"source": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource": {
										Type:       schema.TypeString,
										Sensitive:  true,
										Optional:   true,
										Deprecated: "Use the aws_codebuild_source_credential resource instead",
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codebuild.SourceAuthType_Values(), false),
										Deprecated:   "Use the aws_codebuild_source_credential resource instead",
									},
								},
							},
							Deprecated: "Use the aws_codebuild_source_credential resource instead",
						},
						"buildspec": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"location": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(codebuild.SourceType_Values(), false),
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
						"report_build_status": {
							Type:     schema.TypeBool,
							Optional: true,
						},
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
					},
				},
			},
			"source_version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"project_visibility": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      codebuild.ProjectVisibilityTypePrivate,
				ValidateFunc: validation.StringInSlice(codebuild.ProjectVisibilityType_Values(), false),
			},
			"public_project_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_access_role": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"build_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(5, 480),
			},
			"queued_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      480,
				ValidateFunc: validation.IntBetween(5, 480),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"subnets": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MaxItems: 16,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							MaxItems: 5,
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// Plan time validation for cache location
				cacheType, cacheTypeOk := diff.GetOk("cache.0.type")
				if !cacheTypeOk || cacheType.(string) == codebuild.CacheTypeNoCache || cacheType.(string) == codebuild.CacheTypeLocal {
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

func resourceProjectCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	projectEnv := expandProjectEnvironment(d)
	projectSource := expandProjectSource(d)
	projectArtifacts := expandProjectArtifacts(d)
	projectSecondaryArtifacts := expandProjectSecondaryArtifacts(d)
	projectSecondarySources := expandProjectSecondarySources(d)
	projectLogsConfig := expandProjectLogsConfig(d)
	projectBatchConfig := expandCodeBuildBuildBatchConfig(d)
	projectFileSystemLocations := expandProjectFileSystemLocations(d)

	if aws.StringValue(projectSource.Type) == codebuild.SourceTypeNoSource {
		if aws.StringValue(projectSource.Buildspec) == "" {
			return fmt.Errorf("`buildspec` must be set when source's `type` is `NO_SOURCE`")
		}

		if aws.StringValue(projectSource.Location) != "" {
			return fmt.Errorf("`location` must be empty when source's `type` is `NO_SOURCE`")
		}
	}

	params := &codebuild.CreateProjectInput{
		Environment:         projectEnv,
		Name:                aws.String(d.Get("name").(string)),
		Source:              &projectSource,
		Artifacts:           &projectArtifacts,
		SecondaryArtifacts:  projectSecondaryArtifacts,
		SecondarySources:    projectSecondarySources,
		LogsConfig:          projectLogsConfig,
		BuildBatchConfig:    projectBatchConfig,
		FileSystemLocations: projectFileSystemLocations,
		Tags:                Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("cache"); ok {
		params.Cache = expandProjectCache(v.([]interface{}))
	}

	if v, ok := d.GetOk("concurrent_build_limit"); ok {
		params.ConcurrentBuildLimit = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_key"); ok {
		params.EncryptionKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_role"); ok {
		params.ServiceRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_version"); ok {
		params.SourceVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("build_timeout"); ok {
		params.TimeoutInMinutes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("queued_timeout"); ok {
		params.QueuedTimeoutInMinutes = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("vpc_config"); ok {
		params.VpcConfig = expandCodeBuildVpcConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("badge_enabled"); ok {
		params.BadgeEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("secondary_source_version"); ok && v.(*schema.Set).Len() > 0 {
		params.SecondarySourceVersions = expandProjectSecondarySourceVersions(v.(*schema.Set))
	}

	var resp *codebuild.CreateProjectOutput
	// Handle IAM eventual consistency
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		var err error

		resp, err = conn.CreateProject(params)
		if err != nil {
			// InvalidInputException: CodeBuild is not authorized to perform
			// InvalidInputException: Not authorized to perform DescribeSecurityGroups
			if tfawserr.ErrMessageContains(err, codebuild.ErrCodeInvalidInputException, "ot authorized to perform") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.CreateProject(params)
	}
	if err != nil {
		return fmt.Errorf("Error creating CodeBuild project: %w", err)
	}

	d.SetId(aws.StringValue(resp.Project.Arn))

	if v, ok := d.GetOk("project_visibility"); ok && v.(string) != codebuild.ProjectVisibilityTypePrivate {

		visInput := &codebuild.UpdateProjectVisibilityInput{
			ProjectArn:        aws.String(d.Id()),
			ProjectVisibility: aws.String(v.(string)),
		}

		if v, ok := d.GetOk("resource_access_role"); ok {
			visInput.ResourceAccessRole = aws.String(v.(string))
		}

		_, err = conn.UpdateProjectVisibility(visInput)
		if err != nil {
			return fmt.Errorf("Error updating CodeBuild project (%s) visibility: %w", d.Id(), err)
		}
	}
	return resourceProjectRead(d, meta)
}

func expandProjectSecondarySourceVersions(ssv *schema.Set) []*codebuild.ProjectSourceVersion {
	sourceVersions := make([]*codebuild.ProjectSourceVersion, 0)

	rawSourceVersions := ssv.List()
	if len(rawSourceVersions) == 0 {
		return nil
	}

	for _, config := range rawSourceVersions {
		sourceVersion := expandProjectSourceVersion(config.(map[string]interface{}))
		sourceVersions = append(sourceVersions, &sourceVersion)
	}

	return sourceVersions
}

func expandProjectSourceVersion(data map[string]interface{}) codebuild.ProjectSourceVersion {

	sourceVersion := codebuild.ProjectSourceVersion{
		SourceIdentifier: aws.String(data["source_identifier"].(string)),
		SourceVersion:    aws.String(data["source_version"].(string)),
	}

	return sourceVersion
}

func expandProjectFileSystemLocations(d *schema.ResourceData) []*codebuild.ProjectFileSystemLocation {
	fileSystemLocations := make([]*codebuild.ProjectFileSystemLocation, 0)

	configsList := d.Get("file_system_locations").(*schema.Set).List()

	if len(configsList) == 0 {
		return nil
	}

	for _, config := range configsList {
		art := expandProjectFileSystemLocation(config.(map[string]interface{}))
		fileSystemLocations = append(fileSystemLocations, &art)
	}

	return fileSystemLocations
}

func expandProjectFileSystemLocation(data map[string]interface{}) codebuild.ProjectFileSystemLocation {
	projectFileSystemLocation := codebuild.ProjectFileSystemLocation{
		Type: aws.String(data["type"].(string)),
	}

	if data["identifier"].(string) != "" {
		projectFileSystemLocation.Identifier = aws.String(data["identifier"].(string))
	}

	if data["location"].(string) != "" {
		projectFileSystemLocation.Location = aws.String(data["location"].(string))
	}

	if data["mount_options"].(string) != "" {
		projectFileSystemLocation.MountOptions = aws.String(data["mount_options"].(string))
	}

	if data["mount_point"].(string) != "" {
		projectFileSystemLocation.MountPoint = aws.String(data["mount_point"].(string))
	}

	return projectFileSystemLocation
}

func expandProjectSecondaryArtifacts(d *schema.ResourceData) []*codebuild.ProjectArtifacts {
	artifacts := make([]*codebuild.ProjectArtifacts, 0)

	configsList := d.Get("secondary_artifacts").(*schema.Set).List()

	if len(configsList) == 0 {
		return nil
	}

	for _, config := range configsList {
		art := expandProjectArtifactData(config.(map[string]interface{}))
		artifacts = append(artifacts, &art)
	}

	return artifacts
}

func expandProjectArtifacts(d *schema.ResourceData) codebuild.ProjectArtifacts {
	configs := d.Get("artifacts").([]interface{})
	data := configs[0].(map[string]interface{})

	return expandProjectArtifactData(data)
}

func expandProjectArtifactData(data map[string]interface{}) codebuild.ProjectArtifacts {
	artifactType := data["type"].(string)

	projectArtifacts := codebuild.ProjectArtifacts{
		Type: aws.String(artifactType),
	}

	// Only valid for S3 and CODEPIPELINE artifacts types
	// InvalidInputException: Invalid artifacts: artifact type NO_ARTIFACTS should have null encryptionDisabled
	if artifactType == codebuild.ArtifactsTypeS3 || artifactType == codebuild.ArtifactsTypeCodepipeline {
		projectArtifacts.EncryptionDisabled = aws.Bool(data["encryption_disabled"].(bool))
	}

	if v, ok := data["artifact_identifier"].(string); ok && v != "" {
		projectArtifacts.ArtifactIdentifier = aws.String(v)
	}

	if v, ok := data["location"].(string); ok && v != "" {
		projectArtifacts.Location = aws.String(v)
	}

	if v, ok := data["name"].(string); ok && v != "" {
		projectArtifacts.Name = aws.String(v)
	}

	if v, ok := data["namespace_type"].(string); ok && v != "" {
		projectArtifacts.NamespaceType = aws.String(v)
	}

	if v, ok := data["override_artifact_name"]; ok {
		projectArtifacts.OverrideArtifactName = aws.Bool(v.(bool))
	}

	if v, ok := data["packaging"].(string); ok && v != "" {
		projectArtifacts.Packaging = aws.String(v)
	}

	if v, ok := data["path"].(string); ok && v != "" {
		projectArtifacts.Path = aws.String(v)
	}

	if v, ok := data["bucket_owner_access"].(string); ok && v != "" {
		projectArtifacts.BucketOwnerAccess = aws.String(v)
	}

	return projectArtifacts
}

func expandProjectCache(s []interface{}) *codebuild.ProjectCache {
	var projectCache *codebuild.ProjectCache

	data := s[0].(map[string]interface{})

	projectCache = &codebuild.ProjectCache{
		Type: aws.String(data["type"].(string)),
	}

	if v, ok := data["location"]; ok {
		projectCache.Location = aws.String(v.(string))
	}

	if cacheType := data["type"]; cacheType == codebuild.CacheTypeLocal {
		if modes, modesOk := data["modes"]; modesOk {
			modesStrings := modes.([]interface{})
			projectCache.Modes = flex.ExpandStringList(modesStrings)
		}
	}

	return projectCache
}

func expandProjectEnvironment(d *schema.ResourceData) *codebuild.ProjectEnvironment {
	configs := d.Get("environment").([]interface{})

	envConfig := configs[0].(map[string]interface{})

	projectEnv := &codebuild.ProjectEnvironment{
		PrivilegedMode: aws.Bool(envConfig["privileged_mode"].(bool)),
	}

	if v := envConfig["compute_type"]; v != nil {
		projectEnv.ComputeType = aws.String(v.(string))
	}

	if v := envConfig["image"]; v != nil {
		projectEnv.Image = aws.String(v.(string))
	}

	if v := envConfig["type"]; v != nil {
		projectEnv.Type = aws.String(v.(string))
	}

	if v, ok := envConfig["certificate"]; ok && v.(string) != "" {
		projectEnv.Certificate = aws.String(v.(string))
	}

	if v := envConfig["image_pull_credentials_type"]; v != nil {
		projectEnv.ImagePullCredentialsType = aws.String(v.(string))
	}

	if v, ok := envConfig["registry_credential"]; ok && len(v.([]interface{})) > 0 {
		config := v.([]interface{})[0].(map[string]interface{})

		projectRegistryCredential := &codebuild.RegistryCredential{}

		if v, ok := config["credential"]; ok && v.(string) != "" {
			projectRegistryCredential.Credential = aws.String(v.(string))
		}

		if v, ok := config["credential_provider"]; ok && v.(string) != "" {
			projectRegistryCredential.CredentialProvider = aws.String(v.(string))
		}

		projectEnv.RegistryCredential = projectRegistryCredential
	}

	if v := envConfig["environment_variable"]; v != nil {
		envVariables := v.([]interface{})
		if len(envVariables) > 0 {
			projectEnvironmentVariables := make([]*codebuild.EnvironmentVariable, 0, len(envVariables))

			for _, envVariablesConfig := range envVariables {
				config := envVariablesConfig.(map[string]interface{})

				projectEnvironmentVar := &codebuild.EnvironmentVariable{}

				if v := config["name"].(string); v != "" {
					projectEnvironmentVar.Name = &v
				}

				if v, ok := config["value"].(string); ok {
					projectEnvironmentVar.Value = &v
				}

				if v := config["type"].(string); v != "" {
					projectEnvironmentVar.Type = &v
				}

				projectEnvironmentVariables = append(projectEnvironmentVariables, projectEnvironmentVar)
			}

			projectEnv.EnvironmentVariables = projectEnvironmentVariables
		}
	}

	return projectEnv
}

func expandProjectLogsConfig(d *schema.ResourceData) *codebuild.LogsConfig {
	logsConfig := &codebuild.LogsConfig{}

	if v, ok := d.GetOk("logs_config"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		configList := v.([]interface{})
		data := configList[0].(map[string]interface{})

		if v, ok := data["cloudwatch_logs"]; ok {
			logsConfig.CloudWatchLogs = expandCodeBuildCloudWatchLogsConfig(v.([]interface{}))
		}

		if v, ok := data["s3_logs"]; ok {
			logsConfig.S3Logs = expandCodeBuildS3LogsConfig(v.([]interface{}))
		}
	}

	if logsConfig.CloudWatchLogs == nil {
		logsConfig.CloudWatchLogs = &codebuild.CloudWatchLogsConfig{
			Status: aws.String(codebuild.LogsConfigStatusTypeEnabled),
		}
	}

	if logsConfig.S3Logs == nil {
		logsConfig.S3Logs = &codebuild.S3LogsConfig{
			Status: aws.String(codebuild.LogsConfigStatusTypeDisabled),
		}
	}

	return logsConfig
}

func expandCodeBuildBuildBatchConfig(d *schema.ResourceData) *codebuild.ProjectBuildBatchConfig {
	configs, ok := d.Get("build_batch_config").([]interface{})
	if !ok || len(configs) == 0 || configs[0] == nil {
		return nil
	}

	data := configs[0].(map[string]interface{})

	projectBuildBatchConfig := &codebuild.ProjectBuildBatchConfig{
		Restrictions: expandCodeBuildBatchRestrictions(data),
		ServiceRole:  aws.String(data["service_role"].(string)),
	}

	if v, ok := data["combine_artifacts"]; ok {
		projectBuildBatchConfig.CombineArtifacts = aws.Bool(v.(bool))
	}

	if v, ok := data["timeout_in_mins"]; ok && v != 0 {
		projectBuildBatchConfig.TimeoutInMins = aws.Int64(int64(v.(int)))
	}

	return projectBuildBatchConfig
}

func expandCodeBuildBatchRestrictions(data map[string]interface{}) *codebuild.BatchRestrictions {
	if v, ok := data["restrictions"]; !ok || len(v.([]interface{})) == 0 || v.([]interface{})[0] == nil {
		return nil
	}

	restrictionsData := data["restrictions"].([]interface{})[0].(map[string]interface{})

	restrictions := &codebuild.BatchRestrictions{}
	if v, ok := restrictionsData["compute_types_allowed"]; ok && len(v.([]interface{})) != 0 {
		restrictions.ComputeTypesAllowed = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := restrictionsData["maximum_builds_allowed"]; ok && v != 0 {
		restrictions.MaximumBuildsAllowed = aws.Int64(int64(v.(int)))
	}

	return restrictions
}

func expandCodeBuildCloudWatchLogsConfig(configList []interface{}) *codebuild.CloudWatchLogsConfig {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	status := data["status"].(string)

	cloudWatchLogsConfig := &codebuild.CloudWatchLogsConfig{
		Status: aws.String(status),
	}

	if v, ok := data["group_name"]; ok {
		groupName := v.(string)
		if len(groupName) > 0 {
			cloudWatchLogsConfig.GroupName = aws.String(groupName)
		}
	}

	if v, ok := data["stream_name"]; ok {
		streamName := v.(string)
		if len(streamName) > 0 {
			cloudWatchLogsConfig.StreamName = aws.String(streamName)
		}
	}

	return cloudWatchLogsConfig
}

func expandCodeBuildS3LogsConfig(configList []interface{}) *codebuild.S3LogsConfig {
	if len(configList) == 0 || configList[0] == nil {
		return nil
	}

	data := configList[0].(map[string]interface{})

	status := data["status"].(string)

	s3LogsConfig := &codebuild.S3LogsConfig{
		Status: aws.String(status),
	}

	if v, ok := data["location"].(string); ok && v != "" {
		s3LogsConfig.Location = aws.String(v)
	}

	if v, ok := data["bucket_owner_access"].(string); ok && v != "" {
		s3LogsConfig.BucketOwnerAccess = aws.String(v)
	}

	s3LogsConfig.EncryptionDisabled = aws.Bool(data["encryption_disabled"].(bool))

	return s3LogsConfig
}

func expandCodeBuildVpcConfig(rawVpcConfig []interface{}) *codebuild.VpcConfig {
	vpcConfig := codebuild.VpcConfig{}
	if len(rawVpcConfig) == 0 || rawVpcConfig[0] == nil {
		return &vpcConfig
	}

	data := rawVpcConfig[0].(map[string]interface{})
	vpcConfig.VpcId = aws.String(data["vpc_id"].(string))
	vpcConfig.Subnets = flex.ExpandStringSet(data["subnets"].(*schema.Set))
	vpcConfig.SecurityGroupIds = flex.ExpandStringSet(data["security_group_ids"].(*schema.Set))

	return &vpcConfig
}

func expandProjectSecondarySources(d *schema.ResourceData) []*codebuild.ProjectSource {
	configs := d.Get("secondary_sources").(*schema.Set).List()

	if len(configs) == 0 {
		return nil
	}

	sources := make([]*codebuild.ProjectSource, 0)

	for _, config := range configs {
		source := expandProjectSourceData(config.(map[string]interface{}))
		sources = append(sources, &source)
	}

	return sources
}

func expandProjectSource(d *schema.ResourceData) codebuild.ProjectSource {
	configs := d.Get("source").([]interface{})

	data := configs[0].(map[string]interface{})
	return expandProjectSourceData(data)
}

func expandProjectSourceData(data map[string]interface{}) codebuild.ProjectSource {
	sourceType := data["type"].(string)

	projectSource := codebuild.ProjectSource{
		Buildspec:     aws.String(data["buildspec"].(string)),
		GitCloneDepth: aws.Int64(int64(data["git_clone_depth"].(int))),
		InsecureSsl:   aws.Bool(data["insecure_ssl"].(bool)),
		Type:          aws.String(sourceType),
	}

	if data["source_identifier"] != nil {
		projectSource.SourceIdentifier = aws.String(data["source_identifier"].(string))
	}

	if data["location"].(string) != "" {
		projectSource.Location = aws.String(data["location"].(string))
	}

	// Only valid for BITBUCKET, GITHUB, and GITHUB_ENTERPRISE source types, e.g.
	// InvalidInputException: Source type NO_SOURCE does not support ReportBuildStatus
	if sourceType == codebuild.SourceTypeBitbucket || sourceType == codebuild.SourceTypeGithub || sourceType == codebuild.SourceTypeGithubEnterprise {
		projectSource.ReportBuildStatus = aws.Bool(data["report_build_status"].(bool))
	}

	// Probe data for auth details (max of 1 auth per ProjectSource object)
	if v, ok := data["auth"]; ok && len(v.([]interface{})) > 0 {
		if auths := v.([]interface{}); auths[0] != nil {
			auth := auths[0].(map[string]interface{})
			projectSource.Auth = &codebuild.SourceAuth{
				Type:     aws.String(auth["type"].(string)),
				Resource: aws.String(auth["resource"].(string)),
			}
		}
	}

	// Only valid for CODECOMMIT, GITHUB, GITHUB_ENTERPRISE, BITBUCKET source types.
	if sourceType == codebuild.SourceTypeCodecommit || sourceType == codebuild.SourceTypeGithub || sourceType == codebuild.SourceTypeGithubEnterprise || sourceType == codebuild.SourceTypeBitbucket {
		if v, ok := data["git_submodules_config"]; ok && len(v.([]interface{})) > 0 {
			config := v.([]interface{})[0].(map[string]interface{})

			gitSubmodulesConfig := &codebuild.GitSubmodulesConfig{}

			if v, ok := config["fetch_submodules"]; ok {
				gitSubmodulesConfig.FetchSubmodules = aws.Bool(v.(bool))
			}

			projectSource.GitSubmodulesConfig = gitSubmodulesConfig
		}
	}

	// Only valid for BITBUCKET, GITHUB, GITHUB_ENTERPRISE source types.
	if sourceType == codebuild.SourceTypeBitbucket || sourceType == codebuild.SourceTypeGithub || sourceType == codebuild.SourceTypeGithubEnterprise {
		if v, ok := data["build_status_config"]; ok && len(v.([]interface{})) > 0 {
			config := v.([]interface{})[0].(map[string]interface{})

			buildStatusConfig := &codebuild.BuildStatusConfig{}

			if v, ok := config["context"]; ok {
				buildStatusConfig.Context = aws.String(v.(string))
			}
			if v, ok := config["target_url"]; ok {
				buildStatusConfig.TargetUrl = aws.String(v.(string))
			}

			projectSource.BuildStatusConfig = buildStatusConfig
		}
	}

	return projectSource
}

func resourceProjectRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	project, err := FindProjectByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodeBuild Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error Listing CodeBuild Projects: %w", err)
	}

	if err := d.Set("artifacts", flattenProjectArtifacts(project.Artifacts)); err != nil {
		return fmt.Errorf("error setting artifacts: %w", err)
	}

	if err := d.Set("environment", flattenProjectEnvironment(project.Environment)); err != nil {
		return fmt.Errorf("error setting environment: %w", err)
	}

	if err := d.Set("file_system_locations", flattenProjectFileSystemLocations(project.FileSystemLocations)); err != nil {
		return fmt.Errorf("error setting file_system_locations: %w", err)
	}

	if err := d.Set("cache", flattenProjectCache(project.Cache)); err != nil {
		return fmt.Errorf("error setting cache: %w", err)
	}

	if err := d.Set("logs_config", flattenLogsConfig(project.LogsConfig)); err != nil {
		return fmt.Errorf("error setting logs_config: %w", err)
	}

	if err := d.Set("secondary_artifacts", flattenProjectSecondaryArtifacts(project.SecondaryArtifacts)); err != nil {
		return fmt.Errorf("error setting secondary_artifacts: %w", err)
	}

	if err := d.Set("secondary_sources", flattenProjectSecondarySources(project.SecondarySources)); err != nil {
		return fmt.Errorf("error setting secondary_sources: %w", err)
	}

	if err := d.Set("secondary_source_version", flattenAwsCodeBuildProjectSecondarySourceVersions(project.SecondarySourceVersions)); err != nil {
		return fmt.Errorf("error setting secondary_source_version: %w", err)
	}

	if err := d.Set("source", flattenProjectSource(project.Source)); err != nil {
		return fmt.Errorf("error setting source: %w", err)
	}

	if err := d.Set("vpc_config", flattenVPCConfig(project.VpcConfig)); err != nil {
		return fmt.Errorf("error setting vpc_config: %w", err)
	}

	if err := d.Set("build_batch_config", flattenBuildBatchConfig(project.BuildBatchConfig)); err != nil {
		return fmt.Errorf("error setting build_batch_config: %w", err)
	}

	d.Set("arn", project.Arn)
	d.Set("concurrent_build_limit", project.ConcurrentBuildLimit)
	d.Set("description", project.Description)
	d.Set("encryption_key", project.EncryptionKey)
	d.Set("name", project.Name)
	d.Set("service_role", project.ServiceRole)
	d.Set("source_version", project.SourceVersion)
	d.Set("build_timeout", project.TimeoutInMinutes)
	d.Set("project_visibility", project.ProjectVisibility)
	d.Set("public_project_alias", project.PublicProjectAlias)
	d.Set("resource_access_role", project.ResourceAccessRole)
	d.Set("queued_timeout", project.QueuedTimeoutInMinutes)
	if project.Badge != nil {
		d.Set("badge_enabled", project.Badge.BadgeEnabled)
		d.Set("badge_url", project.Badge.BadgeRequestUrl)
	} else {
		d.Set("badge_enabled", false)
		d.Set("badge_url", "")
	}

	tags := KeyValueTags(project.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceProjectUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if d.HasChanges("project_visibility", "resource_access_role") {

		visInput := &codebuild.UpdateProjectVisibilityInput{
			ProjectArn:        aws.String(d.Id()),
			ProjectVisibility: aws.String(d.Get("project_visibility").(string)),
		}

		if v, ok := d.GetOk("resource_access_role"); ok {
			visInput.ResourceAccessRole = aws.String(v.(string))
		}

		_, err := conn.UpdateProjectVisibility(visInput)
		if err != nil {
			return fmt.Errorf("Error updating CodeBuild project (%s) visibility: %w", d.Id(), err)
		}
	}

	if d.HasChangesExcept("project_visibility", "resource_access_role") {

		params := &codebuild.UpdateProjectInput{
			Name: aws.String(d.Get("name").(string)),
		}

		if d.HasChange("environment") {
			projectEnv := expandProjectEnvironment(d)
			params.Environment = projectEnv
		}

		if d.HasChange("file_system_locations") {
			projectFileSystemLocations := expandProjectFileSystemLocations(d)
			params.FileSystemLocations = projectFileSystemLocations
		}

		if d.HasChange("source") {
			projectSource := expandProjectSource(d)
			params.Source = &projectSource
		}

		if d.HasChange("artifacts") {
			projectArtifacts := expandProjectArtifacts(d)
			params.Artifacts = &projectArtifacts
		}

		if d.HasChange("secondary_sources") {
			_, n := d.GetChange("secondary_sources")

			if n.(*schema.Set).Len() > 0 {
				projectSecondarySources := expandProjectSecondarySources(d)
				params.SecondarySources = projectSecondarySources
			} else {
				params.SecondarySources = []*codebuild.ProjectSource{}
			}
		}

		if d.HasChange("secondary_source_version") {
			_, n := d.GetChange("secondary_source_version")

			psv := d.Get("secondary_source_version").(*schema.Set)

			if n.(*schema.Set).Len() > 0 {
				params.SecondarySourceVersions = expandProjectSecondarySourceVersions(psv)
			} else {
				params.SecondarySourceVersions = []*codebuild.ProjectSourceVersion{}
			}
		}

		if d.HasChange("secondary_artifacts") {
			_, n := d.GetChange("secondary_artifacts")

			if n.(*schema.Set).Len() > 0 {
				projectSecondaryArtifacts := expandProjectSecondaryArtifacts(d)
				params.SecondaryArtifacts = projectSecondaryArtifacts
			} else {
				params.SecondaryArtifacts = []*codebuild.ProjectArtifacts{}
			}
		}

		if d.HasChange("vpc_config") {
			params.VpcConfig = expandCodeBuildVpcConfig(d.Get("vpc_config").([]interface{}))
		}

		if d.HasChange("logs_config") {
			logsConfig := expandProjectLogsConfig(d)
			params.LogsConfig = logsConfig
		}

		if d.HasChange("build_batch_config") {
			params.BuildBatchConfig = expandCodeBuildBuildBatchConfig(d)
		}

		if d.HasChange("cache") {
			if v, ok := d.GetOk("cache"); ok {
				params.Cache = expandProjectCache(v.([]interface{}))
			} else {
				params.Cache = &codebuild.ProjectCache{
					Type: aws.String("NO_CACHE"),
				}
			}
		}

		if d.HasChange("concurrent_build_limit") {
			params.ConcurrentBuildLimit = aws.Int64(int64(d.Get("concurrent_build_limit").(int)))
		}

		if d.HasChange("description") {
			params.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("encryption_key") {
			params.EncryptionKey = aws.String(d.Get("encryption_key").(string))
		}

		if d.HasChange("service_role") {
			params.ServiceRole = aws.String(d.Get("service_role").(string))
		}

		if d.HasChange("source_version") {
			params.SourceVersion = aws.String(d.Get("source_version").(string))
		}

		if d.HasChange("build_timeout") {
			params.TimeoutInMinutes = aws.Int64(int64(d.Get("build_timeout").(int)))
		}

		if d.HasChange("queued_timeout") {
			params.QueuedTimeoutInMinutes = aws.Int64(int64(d.Get("queued_timeout").(int)))
		}

		if d.HasChange("badge_enabled") {
			params.BadgeEnabled = aws.Bool(d.Get("badge_enabled").(bool))
		}

		// The documentation clearly says "The replacement set of tags for this build project."
		// But its a slice of pointers so if not set for every update, they get removed.
		params.Tags = Tags(tags.IgnoreAWS())

		// Handle IAM eventual consistency
		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			_, err := conn.UpdateProject(params)
			if err != nil {
				// InvalidInputException: CodeBuild is not authorized to perform
				// InvalidInputException: Not authorized to perform DescribeSecurityGroups
				if tfawserr.ErrMessageContains(err, codebuild.ErrCodeInvalidInputException, "ot authorized to perform") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = conn.UpdateProject(params)
		}
		if err != nil {
			return fmt.Errorf("[ERROR] Error updating CodeBuild project (%s): %w", d.Id(), err)
		}
	}

	return resourceProjectRead(d, meta)
}

func resourceProjectDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeBuildConn

	_, err := conn.DeleteProject(&codebuild.DeleteProjectInput{
		Name: aws.String(d.Id()),
	})
	return err
}

func flattenProjectFileSystemLocations(apiObjects []*codebuild.ProjectFileSystemLocation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenProjectFileSystemLocation(apiObject))
	}

	return tfList
}

func flattenProjectFileSystemLocation(apiObject *codebuild.ProjectFileSystemLocation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Identifier; v != nil {
		tfMap["identifier"] = aws.StringValue(v)
	}

	if v := apiObject.Location; v != nil {
		tfMap["location"] = aws.StringValue(v)
	}

	if v := apiObject.MountOptions; v != nil {
		tfMap["mount_options"] = aws.StringValue(v)
	}

	if v := apiObject.MountPoint; v != nil {
		tfMap["mount_point"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLogsConfig(logsConfig *codebuild.LogsConfig) []interface{} {
	if logsConfig == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if v := logsConfig.CloudWatchLogs; v != nil {
		values["cloudwatch_logs"] = flattenCloudWatchLogs(v)
	}

	if v := logsConfig.S3Logs; v != nil {
		values["s3_logs"] = flattenS3Logs(v)
	}

	return []interface{}{values}
}

func flattenCloudWatchLogs(cloudWatchLogsConfig *codebuild.CloudWatchLogsConfig) []interface{} {
	values := map[string]interface{}{}

	if cloudWatchLogsConfig == nil {
		values["status"] = codebuild.LogsConfigStatusTypeDisabled
	} else {
		values["status"] = aws.StringValue(cloudWatchLogsConfig.Status)
		values["group_name"] = aws.StringValue(cloudWatchLogsConfig.GroupName)
		values["stream_name"] = aws.StringValue(cloudWatchLogsConfig.StreamName)
	}

	return []interface{}{values}
}

func flattenS3Logs(s3LogsConfig *codebuild.S3LogsConfig) []interface{} {
	values := map[string]interface{}{}

	if s3LogsConfig == nil {
		values["status"] = codebuild.LogsConfigStatusTypeDisabled
	} else {
		values["status"] = aws.StringValue(s3LogsConfig.Status)
		values["location"] = aws.StringValue(s3LogsConfig.Location)
		values["encryption_disabled"] = aws.BoolValue(s3LogsConfig.EncryptionDisabled)
		values["bucket_owner_access"] = aws.StringValue(s3LogsConfig.BucketOwnerAccess)
	}

	return []interface{}{values}
}

func flattenProjectSecondaryArtifacts(artifactsList []*codebuild.ProjectArtifacts) *schema.Set {
	artifactSet := schema.Set{
		F: resourceProjectArtifactsHash,
	}

	for _, artifacts := range artifactsList {
		artifactSet.Add(flattenProjectArtifactsData(*artifacts))
	}
	return &artifactSet
}

func flattenProjectArtifacts(artifacts *codebuild.ProjectArtifacts) []interface{} {
	return []interface{}{flattenProjectArtifactsData(*artifacts)}
}

func flattenProjectArtifactsData(artifacts codebuild.ProjectArtifacts) map[string]interface{} {
	values := map[string]interface{}{}

	values["type"] = aws.StringValue(artifacts.Type)

	if artifacts.ArtifactIdentifier != nil {
		values["artifact_identifier"] = aws.StringValue(artifacts.ArtifactIdentifier)
	}

	if artifacts.EncryptionDisabled != nil {
		values["encryption_disabled"] = aws.BoolValue(artifacts.EncryptionDisabled)
	}

	if artifacts.OverrideArtifactName != nil {
		values["override_artifact_name"] = aws.BoolValue(artifacts.OverrideArtifactName)
	}

	if artifacts.Location != nil {
		values["location"] = aws.StringValue(artifacts.Location)
	}

	if artifacts.Name != nil {
		values["name"] = aws.StringValue(artifacts.Name)
	}

	if artifacts.NamespaceType != nil {
		values["namespace_type"] = aws.StringValue(artifacts.NamespaceType)
	}

	if artifacts.Packaging != nil {
		values["packaging"] = aws.StringValue(artifacts.Packaging)
	}

	if artifacts.Path != nil {
		values["path"] = aws.StringValue(artifacts.Path)
	}

	if artifacts.BucketOwnerAccess != nil {
		values["bucket_owner_access"] = aws.StringValue(artifacts.BucketOwnerAccess)
	}
	return values
}

func flattenProjectCache(cache *codebuild.ProjectCache) []interface{} {
	if cache == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"location": aws.StringValue(cache.Location),
		"type":     aws.StringValue(cache.Type),
		"modes":    aws.StringValueSlice(cache.Modes),
	}

	return []interface{}{values}
}

func flattenProjectEnvironment(environment *codebuild.ProjectEnvironment) []interface{} {
	envConfig := map[string]interface{}{}

	envConfig["type"] = aws.StringValue(environment.Type)
	envConfig["compute_type"] = aws.StringValue(environment.ComputeType)
	envConfig["image"] = aws.StringValue(environment.Image)
	envConfig["certificate"] = aws.StringValue(environment.Certificate)
	envConfig["privileged_mode"] = aws.BoolValue(environment.PrivilegedMode)
	envConfig["image_pull_credentials_type"] = aws.StringValue(environment.ImagePullCredentialsType)

	envConfig["registry_credential"] = flattenRegistryCredential(environment.RegistryCredential)

	if environment.EnvironmentVariables != nil {
		envConfig["environment_variable"] = environmentVariablesToMap(environment.EnvironmentVariables)
	}

	return []interface{}{envConfig}
}

func flattenRegistryCredential(registryCredential *codebuild.RegistryCredential) []interface{} {
	if registryCredential == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"credential":          aws.StringValue(registryCredential.Credential),
		"credential_provider": aws.StringValue(registryCredential.CredentialProvider),
	}

	return []interface{}{values}
}

func flattenProjectSecondarySources(sourceList []*codebuild.ProjectSource) []interface{} {
	l := make([]interface{}, 0)

	for _, source := range sourceList {
		l = append(l, flattenProjectSourceData(source))
	}

	return l
}

func flattenProjectSource(source *codebuild.ProjectSource) []interface{} {
	l := make([]interface{}, 1)

	l[0] = flattenProjectSourceData(source)

	return l
}

func flattenProjectSourceData(source *codebuild.ProjectSource) interface{} {
	m := map[string]interface{}{
		"buildspec":           aws.StringValue(source.Buildspec),
		"location":            aws.StringValue(source.Location),
		"git_clone_depth":     int(aws.Int64Value(source.GitCloneDepth)),
		"insecure_ssl":        aws.BoolValue(source.InsecureSsl),
		"report_build_status": aws.BoolValue(source.ReportBuildStatus),
		"type":                aws.StringValue(source.Type),
	}

	m["git_submodules_config"] = flattenProjectGitSubmodulesConfig(source.GitSubmodulesConfig)

	m["build_status_config"] = flattenProjectBuildStatusConfig(source.BuildStatusConfig)

	if source.Auth != nil {
		m["auth"] = []interface{}{sourceAuthToMap(source.Auth)}
	}
	if source.SourceIdentifier != nil {
		m["source_identifier"] = aws.StringValue(source.SourceIdentifier)
	}

	return m
}

func flattenAwsCodeBuildProjectSecondarySourceVersions(sourceVersions []*codebuild.ProjectSourceVersion) []interface{} {
	l := make([]interface{}, 0)

	for _, sourceVersion := range sourceVersions {
		l = append(l, flattenAwsCodeBuildProjectsourceVersionsData(sourceVersion))
	}
	return l
}

func flattenAwsCodeBuildProjectsourceVersionsData(sourceVersion *codebuild.ProjectSourceVersion) map[string]interface{} {
	values := map[string]interface{}{}

	if sourceVersion.SourceIdentifier != nil {
		values["source_identifier"] = aws.StringValue(sourceVersion.SourceIdentifier)
	}

	if sourceVersion.SourceVersion != nil {
		values["source_version"] = aws.StringValue(sourceVersion.SourceVersion)
	}

	return values
}

func flattenProjectGitSubmodulesConfig(config *codebuild.GitSubmodulesConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"fetch_submodules": aws.BoolValue(config.FetchSubmodules),
	}

	return []interface{}{values}
}

func flattenProjectBuildStatusConfig(config *codebuild.BuildStatusConfig) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"context":    aws.StringValue(config.Context),
		"target_url": aws.StringValue(config.TargetUrl),
	}

	return []interface{}{values}
}

func flattenVPCConfig(vpcConfig *codebuild.VpcConfig) []interface{} {
	if vpcConfig != nil {
		values := map[string]interface{}{}

		values["vpc_id"] = aws.StringValue(vpcConfig.VpcId)
		values["subnets"] = flex.FlattenStringSet(vpcConfig.Subnets)
		values["security_group_ids"] = flex.FlattenStringSet(vpcConfig.SecurityGroupIds)

		return []interface{}{values}
	}
	return nil
}

func flattenBuildBatchConfig(buildBatchConfig *codebuild.ProjectBuildBatchConfig) []interface{} {
	if buildBatchConfig == nil {
		return nil
	}

	values := map[string]interface{}{}

	values["service_role"] = aws.StringValue(buildBatchConfig.ServiceRole)

	if buildBatchConfig.CombineArtifacts != nil {
		values["combine_artifacts"] = aws.BoolValue(buildBatchConfig.CombineArtifacts)
	}

	if buildBatchConfig.Restrictions != nil {
		values["restrictions"] = flattenBuildBatchRestrictionsConfig(buildBatchConfig.Restrictions)
	}

	if buildBatchConfig.TimeoutInMins != nil {
		values["timeout_in_mins"] = aws.Int64Value(buildBatchConfig.TimeoutInMins)
	}

	return []interface{}{values}
}

func flattenBuildBatchRestrictionsConfig(restrictions *codebuild.BatchRestrictions) []interface{} {
	if restrictions == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{
		"compute_types_allowed":  aws.StringValueSlice(restrictions.ComputeTypesAllowed),
		"maximum_builds_allowed": aws.Int64Value(restrictions.MaximumBuildsAllowed),
	}

	return []interface{}{values}
}

func resourceProjectArtifactsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	if v, ok := m["artifact_identifier"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["encryption_disabled"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := m["location"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["namespace_type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["override_artifact_name"]; ok {
		buf.WriteString(fmt.Sprintf("%t-", v.(bool)))
	}

	if v, ok := m["packaging"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["path"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["bucket_owner_access"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func environmentVariablesToMap(environmentVariables []*codebuild.EnvironmentVariable) []interface{} {

	envVariables := []interface{}{}
	if len(environmentVariables) > 0 {
		for _, env := range environmentVariables {
			item := map[string]interface{}{}
			item["name"] = aws.StringValue(env.Name)
			item["value"] = aws.StringValue(env.Value)
			if env.Type != nil {
				item["type"] = aws.StringValue(env.Type)
			}
			envVariables = append(envVariables, item)
		}
	}

	return envVariables
}

func sourceAuthToMap(sourceAuth *codebuild.SourceAuth) map[string]interface{} {

	auth := map[string]interface{}{}
	auth["type"] = aws.StringValue(sourceAuth.Type)

	if sourceAuth.Resource != nil {
		auth["resource"] = aws.StringValue(sourceAuth.Resource)
	}

	return auth
}

func ValidProjectName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[A-Za-z0-9]`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"first character of %q must be a letter or number", value))
	}

	if !regexp.MustCompile(`^[A-Za-z0-9\-_]+$`).MatchString(value) {
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

	simplePattern := `^[a-z0-9][^/]*\/(.+)$`
	if !regexp.MustCompile(simplePattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q does not match pattern (%q): %q",
			k, simplePattern, value))
	}

	return
}
