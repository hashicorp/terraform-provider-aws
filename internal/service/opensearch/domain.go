// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package opensearch

import (
	"context"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearch/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_domain", name="Domain")
// @Tags(identifierAttribute="id")
func resourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

				name := d.Id()
				ds, err := findDomainByName(ctx, conn, name)
				if err != nil {
					return nil, fmt.Errorf("reading OpenSearch Domain (%s): %w", name, err)
				}

				d.SetId(aws.ToString(ds.ARN))
				d.Set(names.AttrDomainName, name)

				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(120 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf(names.AttrEngineVersion, func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
				newVersion := d.Get(names.AttrEngineVersion).(string)
				domainName := d.Get(names.AttrDomainName).(string)

				conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)
				resp, err := conn.GetCompatibleVersions(ctx, &opensearch.GetCompatibleVersionsInput{
					DomainName: aws.String(domainName),
				})
				if err != nil {
					log.Printf("[ERROR] Failed to get compatible OpenSearch versions %s", domainName)
					return false
				}
				if len(resp.CompatibleVersions) != 1 {
					return true
				}
				return !slices.Contains(resp.CompatibleVersions[0].TargetVersions, newVersion)
			}),
			validateJWTOptionsVersion,
			customdiff.ForceNewIf("encrypt_at_rest.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				o, n := d.GetChange("encrypt_at_rest.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get(names.AttrEngineVersion).(string))
			}),
			customdiff.ForceNewIf("node_to_node_encryption.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				o, n := d.GetChange("node_to_node_encryption.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get(names.AttrEngineVersion).(string))
			}),
			customdiff.ForceNewIf("advanced_security_options.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				o, n := d.GetChange("advanced_security_options.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return false
			}),
			customdiff.ForceNewIfChange(names.AttrIPAddressType, func(_ context.Context, old, new, meta any) bool {
				return (old.(string) == string(awstypes.IPAddressTypeDualstack)) && old.(string) != new.(string)
			}),
		),

		Schema: map[string]*schema.Schema{
			"access_policies": sdkv2.IAMPolicyDocumentSchemaOptionalComputed(),
			"advanced_options": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"advanced_security_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"anonymous_auth_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						"internal_user_database_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"jwt_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									names.AttrPublicKey: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										DiffSuppressFunc: suppressPublicKeyDiff,
									},
									"roles_key": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
									"subject_key": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(1, 64),
									},
								},
							},
						},
						"master_user_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"master_user_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"master_user_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"master_user_password": {
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
									},
								},
							},
						},
					},
				},
			},
			"aiml_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"natural_language_query_generation_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"desired_state": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.NaturalLanguageQueryGenerationDesiredState](),
									},
								},
							},
						},
						"s3_vectors_engine": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
								},
							},
						},
						"serverless_vector_acceleration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Computed: true,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_tune_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"desired_state": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AutoTuneDesiredState](),
						},
						"maintenance_schedule": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cron_expression_for_recurrence": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrDuration: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrUnit: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[awstypes.TimeUnit](),
												},
												names.AttrValue: {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"start_at": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},
								},
							},
						},
						"rollback_on_disable": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RollbackOnDisable](),
						},
						"use_off_peak_window": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cold_storage_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"dedicated_master_count": {
							Type:             schema.TypeInt,
							Optional:         true,
							DiffSuppressFunc: suppressComputedDedicatedMaster,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"dedicated_master_type": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressComputedDedicatedMaster,
						},
						names.AttrInstanceCount: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.OpenSearchPartitionInstanceTypeM3MediumSearch,
						},
						"multi_az_with_standby_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"node_options": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_config": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"count": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												names.AttrEnabled: {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
												names.AttrType: {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"node_type": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.NodeOptionsNodeType](),
									},
								},
							},
						},
						"warm_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 150),
						},
						"warm_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"warm_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.OpenSearchWarmPartitionInstanceType](),
						},
						"zone_awareness_config": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone_count": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      2,
										ValidateFunc: validation.IntInSlice([]int{2, 3}),
									},
								},
							},
						},
						"zone_awareness_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"cognito_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrUserPoolID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"dashboard_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_endpoint_v2": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_endpoint_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"custom_endpoint": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: isCustomEndpointDisabled,
						},
						"custom_endpoint_certificate_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     verify.ValidARN,
							DiffSuppressFunc: isCustomEndpointDisabled,
						},
						"custom_endpoint_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"enforce_https": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"tls_security_policy": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TLSSecurityPolicy](),
						},
					},
				},
			},
			"domain_endpoint_v2_hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z][0-9a-z\-]{2,27}$`),
					"must start with a lowercase alphabet and be at least 3 and no more than 28 characters long."+
						" Valid characters are a-z (lowercase letters), 0-9, and - (hyphen)."),
			},
			"ebs_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ebs_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						names.AttrIOPS: {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						names.AttrThroughput: {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(125),
						},
						names.AttrVolumeSize: {
							Type:     schema.TypeInt,
							Optional: true,
						},
						names.AttrVolumeType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
						},
					},
				},
			},
			"encrypt_at_rest": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
						names.AttrKMSKeyID: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressEquivalentKMSKeyIDs,
						},
					},
				},
			},
			names.AttrEndpoint: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_v2": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"identity_center_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled_api_access": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"identity_center_instance_arn": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateFunc:     verify.ValidARN,
							DiffSuppressFunc: suppressDiffIfIdentityCenterOptionsDisabled,
						},
						"roles_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RolesKeyIdCOption](),
							DiffSuppressFunc: suppressDiffIfIdentityCenterOptionsDisabled,
						},
						"subject_key": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SubjectKeyIdCOption](),
							DiffSuppressFunc: suppressDiffIfIdentityCenterOptionsDisabled,
						},
					},
				},
			},
			names.AttrIPAddressType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.IPAddressType](),
			},
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCloudWatchLogGroupARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"log_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LogType](),
						},
					},
				},
			},
			"node_to_node_encryption": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"off_peak_window_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"off_peak_window": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"window_start_time": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hours": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
												"minutes": {
													Type:     schema.TypeInt,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"snapshot_options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"automated_snapshot_start_hour": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"software_update_options": {
				Type:             schema.TypeList,
				Optional:         true,
				Computed:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_software_update_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func suppressDiffIfIdentityCenterOptionsDisabled(_, _, _ string, d *schema.ResourceData) bool {
	// `!ok` means the attribute is not set, or the attribute is set to false
	if _, ok := d.GetOk("identity_center_options.0.enabled_api_access"); !ok {
		return true
	}
	return false
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state.
	name := d.Get(names.AttrDomainName).(string)
	_, err := findDomainByName(ctx, conn, name)

	if err == nil {
		return sdkdiag.AppendErrorf(diags, "OpenSearch Domain (%s) already exists", name)
	}

	input := opensearch.CreateDomainInput{
		DomainName: aws.String(name),
		TagList:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_policies"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input.AccessPolicies = aws.String(policy)
	}

	if v, ok := d.GetOk("advanced_options"); ok {
		input.AdvancedOptions = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("advanced_security_options"); ok {
		input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(v.([]any))
	}

	if v, ok := d.GetOk("aiml_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.AIMLOptions = expandAIMLOptionsInput(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]any)) > 0 {
		input.AutoTuneOptions = expandAutoTuneOptionsInput(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		config := v.([]any)

		if len(config) == 1 {
			if config[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside cluster_config")
			}
			m := config[0].(map[string]any)
			input.ClusterConfig = expandClusterConfig(m)
		}
	}

	if v, ok := d.GetOk("cognito_options"); ok {
		input.CognitoOptions = expandCognitoOptions(v.([]any))
	}

	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		input.DomainEndpointOptions = expandDomainEndpointOptions(v.([]any))
	}

	if v, ok := d.GetOk("ebs_options"); ok {
		options := v.([]any)

		if len(options) == 1 {
			if options[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside ebs_options")
			}

			s := options[0].(map[string]any)
			input.EBSOptions = expandEBSOptions(s)
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		options := v.([]any)
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside encrypt_at_rest")
		}

		s := options[0].(map[string]any)
		input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		input.IPAddressType = awstypes.IPAddressType(v.(string))
	}

	if v, ok := d.GetOk("log_publishing_options"); ok && v.(*schema.Set).Len() > 0 {
		input.LogPublishingOptions = expandLogPublishingOptions(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("node_to_node_encryption"); ok {
		options := v.([]any)

		s := options[0].(map[string]any)
		input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
	}

	if v, ok := d.GetOk("off_peak_window_options"); ok && len(v.([]any)) > 0 {
		input.OffPeakWindowOptions = expandOffPeakWindowOptions(v.([]any)[0].(map[string]any))

		// This option is only available when modifying a domain created prior to February 16, 2023, not when creating a new domain.
		// An off-peak window is required for a domain and cannot be disabled.
		if input.OffPeakWindowOptions != nil {
			input.OffPeakWindowOptions.Enabled = aws.Bool(true)
		}
	}

	if v, ok := d.GetOk("snapshot_options"); ok {
		options := v.([]any)

		if len(options) == 1 {
			if options[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside snapshot_options")
			}

			o := options[0].(map[string]any)

			snapshotOptions := awstypes.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int32(int32(o["automated_snapshot_start_hour"].(int))),
			}

			input.SnapshotOptions = &snapshotOptions
		}
	}

	if v, ok := d.GetOk("software_update_options"); ok {
		input.SoftwareUpdateOptions = expandSoftwareUpdateOptions(v.([]any))
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]any)
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]any)
		input.VPCOptions = expandVPCOptions(s)
	}

	// IAM Roles can take some time to propagate if set in AccessPolicies and created in the same terraform
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func(ctx context.Context) (any, error) {
			return conn.CreateDomain(ctx, &input)
		},
		domainErrorRetryable,
	)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Domain (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*opensearch.CreateDomainOutput).DomainStatus.ARN))

	if err := waitForDomainCreation(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]any)) > 0 {
		input := opensearch.UpdateDomainConfigInput{
			AutoTuneOptions: expandAutoTuneOptions(v.([]any)[0].(map[string]any)),
			DomainName:      aws.String(name),
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func(ctx context.Context) (any, error) {
				return conn.UpdateDomainConfig(ctx, &input)
			},
			domainErrorRetryable,
		)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s) Config: %s", d.Id(), err)
		}

		if err := waitForDomainUpdate(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) update: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("identity_center_options"); ok {
		input := opensearch.UpdateDomainConfigInput{
			IdentityCenterOptions: expandIdentityCenterOptions(v.([]any)),
			DomainName:            aws.String(name),
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func(ctx context.Context) (any, error) {
				return conn.UpdateDomainConfig(ctx, &input)
			},
			domainErrorRetryable,
		)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s) Config: %s", d.Id(), err)
		}

		if err := waitForDomainUpdate(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	name := d.Get(names.AttrDomainName).(string)
	ds, err := findDomainByName(ctx, conn, name)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] OpenSearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s): %s", d.Id(), err)
	}

	output, err := conn.DescribeDomainConfig(ctx, &opensearch.DescribeDomainConfigInput{
		DomainName: aws.String(name),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s) Config: %s", d.Id(), err)
	}

	dc := output.DomainConfig

	if v := aws.ToString(ds.AccessPolicies); v != "" {
		policies, err := verify.PolicyToSet(d.Get("access_policies").(string), v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("access_policies", policies)
	}
	if err = d.Set("advanced_options", advancedOptionsIgnoreDefault(d.Get("advanced_options").(map[string]any), flex.FlattenStringValueMap(ds.AdvancedOptions))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_options: %s", err)
	}
	// Populate AdvancedSecurityOptions with values returned from
	// DescribeDomainConfig, if enabled, else use
	// values from resource; additionally, append MasterUserOptions
	// from resource as they are not returned from the API
	if v := ds.AdvancedSecurityOptions; v != nil && aws.ToBool(v.Enabled) {
		advSecOpts := flattenAdvancedSecurityOptions(v)
		advSecOpts[0]["master_user_options"] = getMasterUserOptions(d)

		if err := d.Set("advanced_security_options", advSecOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting advanced_security_options: %s", err)
		}
	}
	if ds.AIMLOptions != nil {
		if err := d.Set("aiml_options", []any{flattenAIMLOptionsOutput(ds.AIMLOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting aiml_options: %s", err)
		}
	} else {
		d.Set("aiml_options", nil)
	}
	d.Set(names.AttrARN, ds.ARN)
	if v := dc.AutoTuneOptions; v != nil {
		if err := d.Set("auto_tune_options", []any{flattenAutoTuneOptions(v.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_tune_options: %s", err)
		}
	}
	if err := d.Set("cluster_config", flattenClusterConfig(ds.ClusterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_config: %s", err)
	}
	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_options: %s", err)
	}
	if err := d.Set("domain_endpoint_options", flattenDomainEndpointOptions(ds.DomainEndpointOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_endpoint_options: %s", err)
	}
	d.Set("domain_endpoint_v2_hosted_zone_id", ds.DomainEndpointV2HostedZoneId)
	d.Set("domain_id", ds.DomainId)
	d.Set(names.AttrDomainName, ds.DomainName)
	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_options: %s", err)
	}
	if err := d.Set("encrypt_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encrypt_at_rest: %s", err)
	}
	d.Set(names.AttrEngineVersion, ds.EngineVersion)
	if ds.IdentityCenterOptions != nil {
		if err := d.Set("identity_center_options", flattenIdentityCenterOptions(ds.IdentityCenterOptions)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting identity_center_options: %s", err)
		}
	}
	d.Set(names.AttrIPAddressType, ds.IPAddressType)
	// Remove any disabled log types that aren't in state.
	var inStateLogTypes []string
	if v := d.GetRawState(); !v.IsNull() {
		if v := v.GetAttr("log_publishing_options"); !v.IsNull() {
			for _, v := range v.AsValueSet().Values() {
				if !v.IsNull() {
					for k, v := range v.AsValueMap() {
						if k == "log_type" && !v.IsNull() {
							inStateLogTypes = append(inStateLogTypes, v.AsString())
						}
					}
				}
			}
		}
	}
	maps.DeleteFunc(ds.LogPublishingOptions, func(k string, v awstypes.LogPublishingOption) bool {
		return !aws.ToBool(v.Enabled) && !slices.Contains(inStateLogTypes, k)
	})
	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_publishing_options: %s", err)
	}
	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_to_node_encryption: %s", err)
	}
	if ds.OffPeakWindowOptions != nil {
		if err := d.Set("off_peak_window_options", []any{flattenOffPeakWindowOptions(ds.OffPeakWindowOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting off_peak_window_options: %s", err)
		}
	} else {
		d.Set("off_peak_window_options", nil)
	}
	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_options: %s", err)
	}
	if err := d.Set("software_update_options", flattenSoftwareUpdateOptions(ds.SoftwareUpdateOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting software_update_options: %s", err)
	}
	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", []any{flattenVPCDerivedInfo(ds.VPCOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}

		endpoints := flex.FlattenStringValueMap(ds.Endpoints)
		d.Set(names.AttrEndpoint, endpoints["vpc"])
		d.Set("dashboard_endpoint", getDashboardEndpoint(d.Get(names.AttrEndpoint).(string)))
		if endpoints["vpcv2"] != nil {
			d.Set("endpoint_v2", endpoints["vpcv2"])
			d.Set("dashboard_endpoint_v2", getDashboardEndpoint(d.Get("endpoint_v2").(string)))
		}
		if ds.Endpoint != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch Domain in VPC expected to have null Endpoint value", d.Id())
		}
		if ds.EndpointV2 != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch Domain in VPC expected to have null EndpointV2 value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set(names.AttrEndpoint, ds.Endpoint)
			d.Set("dashboard_endpoint", getDashboardEndpoint(d.Get(names.AttrEndpoint).(string)))
		}
		if ds.EndpointV2 != nil {
			d.Set("endpoint_v2", ds.EndpointV2)
			d.Set("dashboard_endpoint_v2", getDashboardEndpoint(d.Get("endpoint_v2").(string)))
		}
		if ds.Endpoints != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch Domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrDomainName).(string)
		input := opensearch.UpdateDomainConfigInput{
			DomainName: aws.String(name),
		}

		if d.HasChange("access_policies") {
			o, n := d.GetChange("access_policies")

			if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
				policy, err := structure.NormalizeJsonString(n.(string))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
				}
				input.AccessPolicies = aws.String(policy)
			}
		}

		if d.HasChange("advanced_options") {
			input.AdvancedOptions = flex.ExpandStringValueMap(d.Get("advanced_options").(map[string]any))
		}

		if d.HasChange("advanced_security_options") {
			input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(d.Get("advanced_security_options").([]any))

			// When jwt_options block is removed from config, explicitly disable JWT authentication
			if input.AdvancedSecurityOptions.JWTOptions == nil {
				if oldRaw, _ := d.GetChange("advanced_security_options"); len(oldRaw.([]any)) > 0 && oldRaw.([]any)[0] != nil {
					oldMap := oldRaw.([]any)[0].(map[string]any)
					if oldJwt, ok := oldMap["jwt_options"].([]any); ok && len(oldJwt) > 0 && oldJwt[0] != nil {
						input.AdvancedSecurityOptions.JWTOptions = &awstypes.JWTOptionsInput{
							Enabled: aws.Bool(false),
						}
					}
				}
			}
		}

		if d.HasChange("aiml_options") {
			if v, ok := d.GetOk("aiml_options"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				input.AIMLOptions = expandAIMLOptionsInput(v.([]any)[0].(map[string]any))
			}
		}

		if d.HasChange("auto_tune_options") {
			input.AutoTuneOptions = expandAutoTuneOptions(d.Get("auto_tune_options").([]any)[0].(map[string]any))
		}

		if d.HasChange("cognito_options") {
			input.CognitoOptions = expandCognitoOptions(d.Get("cognito_options").([]any))
		}

		if d.HasChange("domain_endpoint_options") {
			input.DomainEndpointOptions = expandDomainEndpointOptions(d.Get("domain_endpoint_options").([]any))
		}

		if d.HasChanges("ebs_options", "cluster_config") {
			options := d.Get("ebs_options").([]any)

			if len(options) == 1 {
				s := options[0].(map[string]any)
				input.EBSOptions = expandEBSOptions(s)
			}

			if d.HasChange("cluster_config") {
				config := d.Get("cluster_config").([]any)

				if len(config) == 1 {
					m := config[0].(map[string]any)
					input.ClusterConfig = expandClusterConfig(m)

					// Work around "ValidationException: Your domain's Elasticsearch version does not support cold storage options. Upgrade to Elasticsearch 7.9 or later.".
					if engineType, version, err := parseEngineVersion(d.Get(names.AttrEngineVersion).(string)); err == nil {
						switch engineType {
						case string(awstypes.EngineTypeElasticsearch):
							if semver.LessThan(version, "7.9") {
								input.ClusterConfig.ColdStorageOptions = nil
							}
						case string(awstypes.EngineTypeOpenSearch):
							// All OpenSearch versions support cold storage options.
						default:
							log.Printf("[WARN] unknown engine type: %s", engineType)
						}
					} else {
						log.Printf("[WARN] %s", err)
					}
				}
			}
		}

		if d.HasChange("encrypt_at_rest") {
			input.EncryptionAtRestOptions = nil
			if v, ok := d.GetOk("encrypt_at_rest"); ok {
				options := v.([]any)
				if options[0] == nil {
					return sdkdiag.AppendErrorf(diags, "at least one field is expected inside encrypt_at_rest")
				}

				s := options[0].(map[string]any)
				input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
			}
		}

		if d.HasChange("identity_center_options") {
			if v, ok := d.GetOk("identity_center_options"); ok && len(v.([]any)) > 0 {
				input.IdentityCenterOptions = expandIdentityCenterOptions(d.Get("identity_center_options").([]any))
			} else {
				// Identity Center Options is disabled when empty object is provided.
				input.IdentityCenterOptions = &awstypes.IdentityCenterOptionsInput{}
			}
		}

		if d.HasChange(names.AttrIPAddressType) {
			input.IPAddressType = awstypes.IPAddressType(d.Get(names.AttrIPAddressType).(string))
		}

		if d.HasChange("log_publishing_options") {
			o, n := d.GetChange("log_publishing_options")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			input.LogPublishingOptions = expandLogPublishingOptions(ns.List())

			// Explicitly disable removed log types.
			oldTypes := tfslices.ApplyToAll(os.List(), func(v any) string {
				return v.(map[string]any)["log_type"].(string)
			})
			newTypes := tfslices.ApplyToAll(ns.List(), func(v any) string {
				return v.(map[string]any)["log_type"].(string)
			})
			_, remove, _ := flex.DiffSlices(oldTypes, newTypes, func(s1, s2 string) bool { return s1 == s2 })
			for _, logType := range remove {
				input.LogPublishingOptions[logType] = awstypes.LogPublishingOption{
					Enabled: aws.Bool(false),
				}
			}
		}

		if d.HasChange("node_to_node_encryption") {
			input.NodeToNodeEncryptionOptions = nil
			if v, ok := d.GetOk("node_to_node_encryption"); ok {
				options := v.([]any)

				s := options[0].(map[string]any)
				input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
			}
		}

		if d.HasChange("off_peak_window_options") {
			input.OffPeakWindowOptions = expandOffPeakWindowOptions(d.Get("off_peak_window_options").([]any)[0].(map[string]any))
		}

		if d.HasChange("snapshot_options") {
			options := d.Get("snapshot_options").([]any)

			if len(options) == 1 {
				o := options[0].(map[string]any)

				snapshotOptions := awstypes.SnapshotOptions{
					AutomatedSnapshotStartHour: aws.Int32(int32(o["automated_snapshot_start_hour"].(int))),
				}

				input.SnapshotOptions = &snapshotOptions
			}
		}

		if d.HasChange("software_update_options") {
			input.SoftwareUpdateOptions = expandSoftwareUpdateOptions(d.Get("software_update_options").([]any))
		}

		if d.HasChange("vpc_options") {
			options := d.Get("vpc_options").([]any)
			s := options[0].(map[string]any)
			input.VPCOptions = expandVPCOptions(s)
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout,
			func(ctx context.Context) (any, error) {
				return conn.UpdateDomainConfig(ctx, &input)
			},
			domainErrorRetryable,
		)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s) Config: %s", d.Id(), err)
		}

		if err := waitForDomainUpdate(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) Config update: %s", d.Id(), err)
		}

		if d.HasChange(names.AttrEngineVersion) {
			input := opensearch.UpgradeDomainInput{
				DomainName:    aws.String(name),
				TargetVersion: aws.String(d.Get(names.AttrEngineVersion).(string)),
			}

			_, err := conn.UpgradeDomain(ctx, &input)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "upgrading OpenSearch Domain (%s): %s", d.Id(), err)
			}

			if _, err := waitUpgradeSucceeded(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) upgrade: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchClient(ctx)

	log.Printf("[DEBUG] Deleting OpenSearch Domain: %s", d.Id())
	name := d.Get(names.AttrDomainName).(string)
	input := opensearch.DeleteDomainInput{
		DomainName: aws.String(name),
	}

	_, err := conn.DeleteDomain(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Domain (%s): %s", d.Id(), err)
	}

	if err := waitForDomainDelete(ctx, conn, name, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findDomainByName(ctx context.Context, conn *opensearch.Client, name string) (*awstypes.DomainStatus, error) {
	input := &opensearch.DescribeDomainInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomain(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainStatus == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.DomainStatus, nil
}

// inPlaceEncryptionEnableVersion returns true if, based on version, encryption
// can be enabled in place (without ForceNew)
func inPlaceEncryptionEnableVersion(version string) bool {
	if engineType, version, err := parseEngineVersion(version); err == nil {
		switch engineType {
		case string(awstypes.EngineTypeElasticsearch):
			if semver.GreaterThanOrEqual(version, "6.7") {
				return true
			}
		case string(awstypes.EngineTypeOpenSearch):
			// All OpenSearch versions support enabling encryption in-place.
			return true
		}
	}

	return false
}

// validateJWTOptionsVersion validates that JWT options are only used with OpenSearch 2.11 or later.
func validateJWTOptionsVersion(_ context.Context, d *schema.ResourceDiff, _ any) error {
	if v, ok := d.GetOk("advanced_security_options"); ok {
		options := v.([]any)
		if len(options) > 0 && options[0] != nil {
			m := options[0].(map[string]any)
			if jwtOptions, ok := m["jwt_options"].([]any); ok && len(jwtOptions) > 0 && jwtOptions[0] != nil {
				jwtMap := jwtOptions[0].(map[string]any)
				if enabled, ok := jwtMap[names.AttrEnabled].(bool); ok && enabled {
					engineVersion := d.Get(names.AttrEngineVersion).(string)
					if engineType, version, err := parseEngineVersion(engineVersion); err == nil {
						switch engineType {
						case string(awstypes.EngineTypeElasticsearch):
							return fmt.Errorf("jwt_options is not supported with Elasticsearch. Use OpenSearch 2.11 or later")
						case string(awstypes.EngineTypeOpenSearch):
							if semver.LessThan(version, "2.11") {
								return fmt.Errorf("jwt_options requires OpenSearch 2.11 or later, got %s", engineVersion)
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func suppressEquivalentKMSKeyIDs(k, old, new string, d *schema.ResourceData) bool {
	// The OpenSearch API accepts a short KMS key id but always returns the ARN of the key.
	// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'.
	// These should be treated as equivalent.
	return strings.Contains(old, new)
}

func suppressPublicKeyDiff(k, old, new string, d *schema.ResourceData) bool {
	// AWS returns the public key without newlines, but users may provide it with newlines.
	// Normalize both values by removing newlines before comparison.
	oldNormalized := strings.ReplaceAll(old, "\n", "")
	newNormalized := strings.ReplaceAll(new, "\n", "")
	return oldNormalized == newNormalized
}

func getDashboardEndpoint(endpoint string) string {
	return endpoint + "/_dashboards"
}

func suppressComputedDedicatedMaster(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("cluster_config")
	if ok {
		clusterConfig := v.([]any)[0].(map[string]any)
		return !clusterConfig["dedicated_master_enabled"].(bool)
	}
	return false
}

func isCustomEndpointDisabled(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("domain_endpoint_options")
	if ok {
		domainEndpointOptions := v.([]any)[0].(map[string]any)
		return !domainEndpointOptions["custom_endpoint_enabled"].(bool)
	}
	return false
}

func expandNodeToNodeEncryptionOptions(s map[string]any) *awstypes.NodeToNodeEncryptionOptions {
	options := awstypes.NodeToNodeEncryptionOptions{}

	if v, ok := s[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	return &options
}

func flattenNodeToNodeEncryptionOptions(o *awstypes.NodeToNodeEncryptionOptions) []map[string]any {
	if o == nil {
		return []map[string]any{}
	}

	m := map[string]any{}
	if o.Enabled != nil {
		m[names.AttrEnabled] = aws.ToBool(o.Enabled)
	}

	return []map[string]any{m}
}

func expandClusterConfig(m map[string]any) *awstypes.ClusterConfig {
	config := awstypes.ClusterConfig{}

	if v, ok := m["cold_storage_options"]; ok {
		config.ColdStorageOptions = expandColdStorageOptions(v.([]any))
	}

	if v, ok := m["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		config.DedicatedMasterEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["dedicated_master_count"]; ok && v.(int) > 0 {
				config.DedicatedMasterCount = aws.Int32(int32(v.(int)))
			}
			if v, ok := m["dedicated_master_type"]; ok && v.(string) != "" {
				config.DedicatedMasterType = awstypes.OpenSearchPartitionInstanceType(v.(string))
			}
		}
	}

	if v, ok := m[names.AttrInstanceCount]; ok {
		config.InstanceCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := m[names.AttrInstanceType]; ok {
		config.InstanceType = awstypes.OpenSearchPartitionInstanceType(v.(string))
	}

	if v, ok := m["multi_az_with_standby_enabled"]; ok {
		config.MultiAZWithStandbyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["node_options"]; ok {
		config.NodeOptions = expandNodeOptions(v.([]any))
	}

	if v, ok := m["warm_enabled"]; ok {
		isEnabled := v.(bool)
		config.WarmEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["warm_count"]; ok {
				config.WarmCount = aws.Int32(int32(v.(int)))
			}

			if v, ok := m["warm_type"]; ok {
				config.WarmType = awstypes.OpenSearchWarmPartitionInstanceType(v.(string))
			}
		}
	}

	if v, ok := m["zone_awareness_enabled"]; ok {
		isEnabled := v.(bool)
		config.ZoneAwarenessEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["zone_awareness_config"]; ok {
				config.ZoneAwarenessConfig = expandZoneAwarenessConfig(v.([]any))
			}
		}
	}

	return &config
}

func expandZoneAwarenessConfig(l []any) *awstypes.ZoneAwarenessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	zoneAwarenessConfig := &awstypes.ZoneAwarenessConfig{}

	if v, ok := m["availability_zone_count"]; ok && v.(int) > 0 {
		zoneAwarenessConfig.AvailabilityZoneCount = aws.Int32(int32(v.(int)))
	}

	return zoneAwarenessConfig
}

func expandColdStorageOptions(l []any) *awstypes.ColdStorageOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	ColdStorageOptions := &awstypes.ColdStorageOptions{}

	if v, ok := m[names.AttrEnabled]; ok {
		ColdStorageOptions.Enabled = aws.Bool(v.(bool))
	}

	return ColdStorageOptions
}

func expandNodeOptions(tfList []any) []awstypes.NodeOption {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := make([]awstypes.NodeOption, 0)
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := awstypes.NodeOption{
			NodeType: awstypes.NodeOptionsNodeType(tfMap["node_type"].(string)),
		}

		if v, ok := tfMap["node_config"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.NodeConfig = expandNodeConfig(v[0].(map[string]any))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandNodeConfig(tfMap map[string]any) *awstypes.NodeConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NodeConfig{}

	isEnabled := tfMap[names.AttrEnabled].(bool)
	apiObject.Enabled = aws.Bool(isEnabled)

	if isEnabled {
		apiObject.Count = aws.Int32(int32(tfMap["count"].(int)))
		apiObject.Type = awstypes.OpenSearchPartitionInstanceType(tfMap[names.AttrType].(string))
	}

	return apiObject
}

func flattenClusterConfig(c *awstypes.ClusterConfig) []map[string]any {
	m := map[string]any{
		"zone_awareness_config":  flattenZoneAwarenessConfig(c.ZoneAwarenessConfig),
		"zone_awareness_enabled": aws.ToBool(c.ZoneAwarenessEnabled),
	}

	if c.ColdStorageOptions != nil {
		m["cold_storage_options"] = flattenColdStorageOptions(c.ColdStorageOptions)
	}
	if c.DedicatedMasterCount != nil {
		m["dedicated_master_count"] = aws.ToInt32(c.DedicatedMasterCount)
	}
	if c.DedicatedMasterEnabled != nil {
		m["dedicated_master_enabled"] = aws.ToBool(c.DedicatedMasterEnabled)
	}

	m["dedicated_master_type"] = c.DedicatedMasterType

	if c.InstanceCount != nil {
		m[names.AttrInstanceCount] = aws.ToInt32(c.InstanceCount)
	}

	m[names.AttrInstanceType] = c.InstanceType

	if c.MultiAZWithStandbyEnabled != nil {
		m["multi_az_with_standby_enabled"] = aws.ToBool(c.MultiAZWithStandbyEnabled)
	}

	if len(c.NodeOptions) > 0 {
		m["node_options"] = flattenNodeOptions(c.NodeOptions)
	}

	if c.WarmEnabled != nil {
		m["warm_enabled"] = aws.ToBool(c.WarmEnabled)
	}
	if c.WarmCount != nil {
		m["warm_count"] = aws.ToInt32(c.WarmCount)
	}

	m["warm_type"] = c.WarmType

	return []map[string]any{m}
}

func flattenZoneAwarenessConfig(zoneAwarenessConfig *awstypes.ZoneAwarenessConfig) []any {
	if zoneAwarenessConfig == nil {
		return []any{}
	}

	m := map[string]any{
		"availability_zone_count": aws.ToInt32(zoneAwarenessConfig.AvailabilityZoneCount),
	}

	return []any{m}
}

func flattenColdStorageOptions(coldStorageOptions *awstypes.ColdStorageOptions) []any {
	if coldStorageOptions == nil {
		return []any{}
	}

	m := map[string]any{
		names.AttrEnabled: aws.ToBool(coldStorageOptions.Enabled),
	}

	return []any{m}
}

func flattenNodeOptions(apiObjects []awstypes.NodeOption) []any {
	if len(apiObjects) == 0 {
		return []any{}
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}
		tfMap["node_config"] = flattenNodeConfig(apiObject.NodeConfig)
		tfMap["node_type"] = apiObject.NodeType
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenNodeConfig(apiObject *awstypes.NodeConfig) []any {
	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	if apiObject.Count != nil {
		tfMap["count"] = aws.ToInt32(apiObject.Count)
	}

	if apiObject.Type != "" {
		tfMap[names.AttrType] = apiObject.Type
	}

	return []any{tfMap}
}

// advancedOptionsIgnoreDefault checks for defaults in the n map and, if
// they don't exist in the o map, it deletes them. AWS returns default advanced
// options that cause perpetual diffs.
func advancedOptionsIgnoreDefault(o map[string]any, n map[string]any) map[string]any {
	for k, v := range n {
		switch fmt.Sprintf("%s=%s", k, v) {
		case "override_main_response_version=false":
			if _, ok := o[k]; !ok {
				delete(n, "override_main_response_version")
			}
		case "rest.action.multi.allow_explicit_index=true":
			if _, ok := o[k]; !ok {
				delete(n, "rest.action.multi.allow_explicit_index")
			}
		}
	}

	return n
}

// parseEngineVersion parses a domain's engine version string into engine type and semver string.
// engine_version is a string of format Elasticsearch_X.Y or OpenSearch_X.Y.
func parseEngineVersion(engineVersion string) (string, string, error) {
	parts := strings.Split(engineVersion, "_")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for engine version (%s)", engineVersion)
	}

	return parts[0], parts[1], nil
}

// EBSVolumeTypePermitsIopsInput returns true if the volume type supports the Iops input
//
// This check prevents a ValidationException when updating EBS volume types from a value
// that supports IOPS (ex. gp3) to one that doesn't (ex. gp2).
func ebsVolumeTypePermitsIopsInput(volumeType awstypes.VolumeType) bool {
	permittedTypes := enum.Slice(awstypes.VolumeTypeGp3, awstypes.VolumeTypeIo1)
	return slices.Contains(permittedTypes, string(volumeType))
}

// EBSVolumeTypePermitsThroughputInput returns true if the volume type supports the Throughput input
//
// This check prevents a ValidationException when updating EBS volume types from a value
// that supports Throughput (ex. gp3) to one that doesn't (ex. gp2).
func ebsVolumeTypePermitsThroughputInput(volumeType awstypes.VolumeType) bool {
	permittedTypes := enum.Slice(awstypes.VolumeTypeGp3)
	return slices.Contains(permittedTypes, string(volumeType))
}

func domainErrorRetryable(err error) (bool, error) {
	switch {
	case errs.IsAErrorMessageContains[*awstypes.InvalidTypeException](err, "Error setting policy"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "enable a service-linked role to give Amazon ES permissions"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Domain is still being deleted"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Amazon OpenSearch Service must be allowed to use the passed role"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "The passed role has not propagated yet"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Authentication error"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Unauthorized Operation: OpenSearch Service must be authorised to describe"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "The passed role must authorize Amazon OpenSearch Service to describe"),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "A change/update is in progress. Please wait for it to complete before requesting another change."),
		errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "The Resource Access Policy specified for the CloudWatch Logs log group"):
		return true, err

	default:
		return false, err
	}
}
