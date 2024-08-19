// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opensearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_opensearch_domain", name="Domain")
// @Tags(identifierAttribute="id")
func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainCreate,
		ReadWithoutTimeout:   resourceDomainRead,
		UpdateWithoutTimeout: resourceDomainUpdate,
		DeleteWithoutTimeout: resourceDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDomainImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(180 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf(names.AttrEngineVersion, func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				newVersion := d.Get(names.AttrEngineVersion).(string)
				domainName := d.Get(names.AttrDomainName).(string)

				conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)
				resp, err := conn.GetCompatibleVersionsWithContext(ctx, &opensearchservice.GetCompatibleVersionsInput{
					DomainName: aws.String(domainName),
				})
				if err != nil {
					log.Printf("[ERROR] Failed to get compatible OpenSearch versions %s", domainName)
					return false
				}
				if len(resp.CompatibleVersions) != 1 {
					return true
				}
				for _, targetVersion := range resp.CompatibleVersions[0].TargetVersions {
					if aws.StringValue(targetVersion) == newVersion {
						return false
					}
				}
				return true
			}),
			customdiff.ForceNewIf("encrypt_at_rest.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				o, n := d.GetChange("encrypt_at_rest.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get(names.AttrEngineVersion).(string))
			}),
			customdiff.ForceNewIf("node_to_node_encryption.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				o, n := d.GetChange("node_to_node_encryption.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get(names.AttrEngineVersion).(string))
			}),
			customdiff.ForceNewIf("advanced_security_options.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				o, n := d.GetChange("advanced_security_options.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return false
			}),
			customdiff.ForceNewIfChange(names.AttrIPAddressType, func(_ context.Context, old, new, meta interface{}) bool {
				return (old.(string) == opensearchservice.IPAddressTypeDualstack) && old.(string) != new.(string)
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.AutoTuneDesiredState_Values(), false),
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
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(opensearchservice.TimeUnit_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.RollbackOnDisable_Values(), false),
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
							Default:  opensearchservice.OpenSearchPartitionInstanceTypeM3MediumSearch,
						},
						"multi_az_with_standby_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.OpenSearchWarmPartitionInstanceType_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.TLSSecurityPolicy_Values(), false),
						},
					},
				},
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.VolumeType_Values(), false),
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
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(opensearchservice.IPAddressType_Values(), false),
			},
			"kibana_endpoint": {
				Type:       schema.TypeString,
				Computed:   true,
				Deprecated: "use 'dashboard_endpoint' attribute instead",
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(opensearchservice.LogType_Values(), false),
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

func resourceDomainImport(ctx context.Context,
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set(names.AttrDomainName, d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state
	resp, err := FindDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))
	if err == nil {
		return sdkdiag.AppendErrorf(diags, "OpenSearch Domain %q already exists", aws.StringValue(resp.DomainName))
	}

	input := &opensearchservice.CreateDomainInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
		TagList:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrEngineVersion); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrIPAddressType); ok {
		input.IPAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("access_policies"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "policy (%s) is invalid JSON: %s", policy, err)
		}

		input.AccessPolicies = aws.String(policy)
	}

	if v, ok := d.GetOk("advanced_options"); ok {
		input.AdvancedOptions = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("advanced_security_options"); ok {
		input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]interface{})) > 0 {
		input.AutoTuneOptions = expandAutoTuneOptionsInput(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ebs_options"); ok {
		options := v.([]interface{})

		if len(options) == 1 {
			if options[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside ebs_options")
			}

			s := options[0].(map[string]interface{})
			input.EBSOptions = expandEBSOptions(s)
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside encrypt_at_rest")
		}

		s := options[0].(map[string]interface{})
		input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		config := v.([]interface{})

		if len(config) == 1 {
			if config[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside cluster_config")
			}
			m := config[0].(map[string]interface{})
			input.ClusterConfig = expandClusterConfig(m)
		}
	}

	if v, ok := d.GetOk("node_to_node_encryption"); ok {
		options := v.([]interface{})

		s := options[0].(map[string]interface{})
		input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
	}

	if v, ok := d.GetOk("snapshot_options"); ok {
		options := v.([]interface{})

		if len(options) == 1 {
			if options[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside snapshot_options")
			}

			o := options[0].(map[string]interface{})

			snapshotOptions := opensearchservice.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
			}

			input.SnapshotOptions = &snapshotOptions
		}
	}

	if v, ok := d.GetOk("software_update_options"); ok {
		input.SoftwareUpdateOptions = expandSoftwareUpdateOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return sdkdiag.AppendErrorf(diags, "At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		input.VPCOptions = expandVPCOptions(s)
	}

	if v, ok := d.GetOk("log_publishing_options"); ok {
		input.LogPublishingOptions = expandLogPublishingOptions(v.(*schema.Set))
	}

	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		input.DomainEndpointOptions = expandDomainEndpointOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("cognito_options"); ok {
		input.CognitoOptions = expandCognitoOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("off_peak_window_options"); ok && len(v.([]interface{})) > 0 {
		input.OffPeakWindowOptions = expandOffPeakWindowOptions(v.([]interface{})[0].(map[string]interface{}))

		// This option is only available when modifying a domain created prior to February 16, 2023, not when creating a new domain.
		// An off-peak window is required for a domain and cannot be disabled.
		if input.OffPeakWindowOptions != nil {
			input.OffPeakWindowOptions.Enabled = aws.Bool(true)
		}
	}

	// IAM Roles can take some time to propagate if set in AccessPolicies and created in the same terraform
	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout, func() (any, error) {
		return conn.CreateDomainWithContext(ctx, input)
	},
		domainErrorRetryable)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating OpenSearch Domain: %s", err)
	}
	out := outputRaw.(*opensearchservice.CreateDomainOutput)

	d.SetId(aws.StringValue(out.DomainStatus.ARN))

	log.Printf("[DEBUG] Waiting for OpenSearch Domain %q to be created", d.Id())
	if err := WaitForDomainCreation(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) to be created: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] OpenSearch Domain %q created", d.Id())

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]interface{})) > 0 {
		input := &opensearchservice.UpdateDomainConfigInput{
			DomainName:      aws.String(d.Get(names.AttrDomainName).(string)),
			AutoTuneOptions: expandAutoTuneOptions(v.([]interface{})[0].(map[string]interface{})),
		}

		log.Printf("[DEBUG] Updating OpenSearch Domain config: %s", input)
		_, err = conn.UpdateDomainConfigWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain config: %s", err)
		}

		if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for OpenSearch Domain (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	ds, err := FindDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] OpenSearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s): %s", d.Id(), err)
	}

	outDescribeDomainConfig, err := conn.DescribeDomainConfigWithContext(ctx, &opensearchservice.DescribeDomainConfigInput{
		DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s): %s", d.Id(), err)
	}

	dc := outDescribeDomainConfig.DomainConfig

	if ds.AccessPolicies != nil && aws.StringValue(ds.AccessPolicies) != "" {
		policies, err := verify.PolicyToSet(d.Get("access_policies").(string), aws.StringValue(ds.AccessPolicies))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s): %s", d.Id(), err)
		}

		d.Set("access_policies", policies)
	}

	options := advancedOptionsIgnoreDefault(d.Get("advanced_options").(map[string]interface{}), flex.FlattenStringMap(ds.AdvancedOptions))
	if err = d.Set("advanced_options", options); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_options %v: %s", options, err)
	}

	d.SetId(aws.StringValue(ds.ARN))
	d.Set(names.AttrARN, ds.ARN)
	d.Set("domain_id", ds.DomainId)
	d.Set(names.AttrDomainName, ds.DomainName)
	d.Set(names.AttrEngineVersion, ds.EngineVersion)
	d.Set(names.AttrIPAddressType, ds.IPAddressType)

	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_options: %s", err)
	}

	if err := d.Set("encrypt_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encrypt_at_rest: %s", err)
	}

	if err := d.Set("cluster_config", flattenClusterConfig(ds.ClusterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_config: %s", err)
	}

	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_options: %s", err)
	}

	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_to_node_encryption: %s", err)
	}

	// Populate AdvancedSecurityOptions with values returned from
	// DescribeDomainConfig, if enabled, else use
	// values from resource; additionally, append MasterUserOptions
	// from resource as they are not returned from the API
	if ds.AdvancedSecurityOptions != nil && aws.BoolValue(ds.AdvancedSecurityOptions.Enabled) {
		advSecOpts := flattenAdvancedSecurityOptions(ds.AdvancedSecurityOptions)
		advSecOpts[0]["master_user_options"] = getMasterUserOptions(d)
		if err := d.Set("advanced_security_options", advSecOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting advanced_security_options: %s", err)
		}
	}

	if v := dc.AutoTuneOptions; v != nil {
		err = d.Set("auto_tune_options", []interface{}{flattenAutoTuneOptions(v.Options)})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading OpenSearch Domain (%s): setting auto_tune_options: %s", d.Id(), err)
		}
	}

	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_options: %s", err)
	}

	if err := d.Set("software_update_options", flattenSoftwareUpdateOptions(ds.SoftwareUpdateOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting software_update_options: %s", err)
	}

	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", []interface{}{flattenVPCDerivedInfo(ds.VPCOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}

		endpoints := flex.FlattenStringMap(ds.Endpoints)
		d.Set(names.AttrEndpoint, endpoints["vpc"])
		d.Set("dashboard_endpoint", getDashboardEndpoint(d))
		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch Domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set(names.AttrEndpoint, ds.Endpoint)
			d.Set("dashboard_endpoint", getDashboardEndpoint(d))
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return sdkdiag.AppendErrorf(diags, "%q: OpenSearch Domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_publishing_options: %s", err)
	}

	if err := d.Set("domain_endpoint_options", flattenDomainEndpointOptions(ds.DomainEndpointOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_endpoint_options: %s", err)
	}

	if ds.OffPeakWindowOptions != nil {
		if err := d.Set("off_peak_window_options", []interface{}{flattenOffPeakWindowOptions(ds.OffPeakWindowOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting off_peak_window_options: %s", err)
		}
	} else {
		d.Set("off_peak_window_options", nil)
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := opensearchservice.UpdateDomainConfigInput{
			DomainName: aws.String(d.Get(names.AttrDomainName).(string)),
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
			input.AdvancedOptions = flex.ExpandStringMap(d.Get("advanced_options").(map[string]interface{}))
		}

		if d.HasChange("advanced_security_options") {
			input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(d.Get("advanced_security_options").([]interface{}))
		}

		if d.HasChange("auto_tune_options") {
			input.AutoTuneOptions = expandAutoTuneOptions(d.Get("auto_tune_options").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("cognito_options") {
			input.CognitoOptions = expandCognitoOptions(d.Get("cognito_options").([]interface{}))
		}

		if d.HasChange("domain_endpoint_options") {
			input.DomainEndpointOptions = expandDomainEndpointOptions(d.Get("domain_endpoint_options").([]interface{}))
		}

		if d.HasChange(names.AttrIPAddressType) {
			input.IPAddressType = aws.String(d.Get(names.AttrIPAddressType).(string))
		}

		if d.HasChanges("ebs_options", "cluster_config") {
			options := d.Get("ebs_options").([]interface{})

			if len(options) == 1 {
				s := options[0].(map[string]interface{})
				input.EBSOptions = expandEBSOptions(s)
			}

			if d.HasChange("cluster_config") {
				config := d.Get("cluster_config").([]interface{})

				if len(config) == 1 {
					m := config[0].(map[string]interface{})
					input.ClusterConfig = expandClusterConfig(m)

					// Work around "ValidationException: Your domain's Elasticsearch version does not support cold storage options. Upgrade to Elasticsearch 7.9 or later.".
					if engineType, version, err := ParseEngineVersion(d.Get(names.AttrEngineVersion).(string)); err == nil {
						switch engineType {
						case opensearchservice.EngineTypeElasticsearch:
							if semver.LessThan(version, "7.9") {
								input.ClusterConfig.ColdStorageOptions = nil
							}
						case opensearchservice.EngineTypeOpenSearch:
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
				options := v.([]interface{})
				if options[0] == nil {
					return sdkdiag.AppendErrorf(diags, "at least one field is expected inside encrypt_at_rest")
				}

				s := options[0].(map[string]interface{})
				input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
			}
		}

		if d.HasChange("log_publishing_options") {
			input.LogPublishingOptions = expandLogPublishingOptions(d.Get("log_publishing_options").(*schema.Set))
		}

		if d.HasChange("node_to_node_encryption") {
			input.NodeToNodeEncryptionOptions = nil
			if v, ok := d.GetOk("node_to_node_encryption"); ok {
				options := v.([]interface{})

				s := options[0].(map[string]interface{})
				input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
			}
		}

		if d.HasChange("off_peak_window_options") {
			input.OffPeakWindowOptions = expandOffPeakWindowOptions(d.Get("off_peak_window_options").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("snapshot_options") {
			options := d.Get("snapshot_options").([]interface{})

			if len(options) == 1 {
				o := options[0].(map[string]interface{})

				snapshotOptions := opensearchservice.SnapshotOptions{
					AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
				}

				input.SnapshotOptions = &snapshotOptions
			}
		}

		if d.HasChange("software_update_options") {
			input.SoftwareUpdateOptions = expandSoftwareUpdateOptions(d.Get("software_update_options").([]interface{}))
		}

		if d.HasChange("vpc_options") {
			options := d.Get("vpc_options").([]interface{})
			s := options[0].(map[string]interface{})
			input.VPCOptions = expandVPCOptions(s)
		}

		_, err := tfresource.RetryWhen(ctx, propagationTimeout, func() (any, error) {
			return conn.UpdateDomainConfigWithContext(ctx, &input)
		},
			domainErrorRetryable)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s): %s", d.Id(), err)
		}

		if err := waitForDomainUpdate(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s): waiting for completion: %s", d.Id(), err)
		}

		if d.HasChange(names.AttrEngineVersion) {
			upgradeInput := opensearchservice.UpgradeDomainInput{
				DomainName:    aws.String(d.Get(names.AttrDomainName).(string)),
				TargetVersion: aws.String(d.Get(names.AttrEngineVersion).(string)),
			}

			_, err := conn.UpgradeDomainWithContext(ctx, &upgradeInput)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s): upgrading: %s", d.Id(), err)
			}

			if _, err := waitUpgradeSucceeded(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating OpenSearch Domain (%s): upgrading: waiting for completion: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OpenSearchConn(ctx)
	domainName := d.Get(names.AttrDomainName).(string)

	log.Printf("[DEBUG] Deleting OpenSearch Domain: %q", domainName)
	_, err := conn.DeleteDomainWithContext(ctx, &opensearchservice.DeleteDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Domain (%s): %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Waiting for OpenSearch Domain %q to be deleted", domainName)
	if err := waitForDomainDelete(ctx, conn, d.Get(names.AttrDomainName).(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting OpenSearch Domain (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func FindDomainByName(ctx context.Context, conn *opensearchservice.OpenSearchService, name string) (*opensearchservice.DomainStatus, error) {
	input := &opensearchservice.DescribeDomainInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeDomainWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, opensearchservice.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainStatus == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainStatus, nil
}

// inPlaceEncryptionEnableVersion returns true if, based on version, encryption
// can be enabled in place (without ForceNew)
func inPlaceEncryptionEnableVersion(version string) bool {
	if engineType, version, err := ParseEngineVersion(version); err == nil {
		switch engineType {
		case opensearchservice.EngineTypeElasticsearch:
			if semver.GreaterThanOrEqual(version, "6.7") {
				return true
			}
		case opensearchservice.EngineTypeOpenSearch:
			// All OpenSearch versions support enabling encryption in-place.
			return true
		}
	}

	return false
}

func suppressEquivalentKMSKeyIDs(k, old, new string, d *schema.ResourceData) bool {
	// The OpenSearch API accepts a short KMS key id but always returns the ARN of the key.
	// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'.
	// These should be treated as equivalent.
	return strings.Contains(old, new)
}

func getDashboardEndpoint(d *schema.ResourceData) string {
	return d.Get(names.AttrEndpoint).(string) + "/_dashboards"
}

func getKibanaEndpoint(d *schema.ResourceData) string {
	return d.Get(names.AttrEndpoint).(string) + "/_plugin/kibana/"
}

func suppressComputedDedicatedMaster(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("cluster_config")
	if ok {
		clusterConfig := v.([]interface{})[0].(map[string]interface{})
		return !clusterConfig["dedicated_master_enabled"].(bool)
	}
	return false
}

func isCustomEndpointDisabled(k, old, new string, d *schema.ResourceData) bool {
	v, ok := d.GetOk("domain_endpoint_options")
	if ok {
		domainEndpointOptions := v.([]interface{})[0].(map[string]interface{})
		return !domainEndpointOptions["custom_endpoint_enabled"].(bool)
	}
	return false
}

func expandNodeToNodeEncryptionOptions(s map[string]interface{}) *opensearchservice.NodeToNodeEncryptionOptions {
	options := opensearchservice.NodeToNodeEncryptionOptions{}

	if v, ok := s[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	return &options
}

func flattenNodeToNodeEncryptionOptions(o *opensearchservice.NodeToNodeEncryptionOptions) []map[string]interface{} {
	if o == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	if o.Enabled != nil {
		m[names.AttrEnabled] = aws.BoolValue(o.Enabled)
	}

	return []map[string]interface{}{m}
}

func expandClusterConfig(m map[string]interface{}) *opensearchservice.ClusterConfig {
	config := opensearchservice.ClusterConfig{}

	if v, ok := m["cold_storage_options"]; ok {
		config.ColdStorageOptions = expandColdStorageOptions(v.([]interface{}))
	}

	if v, ok := m["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		config.DedicatedMasterEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["dedicated_master_count"]; ok && v.(int) > 0 {
				config.DedicatedMasterCount = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["dedicated_master_type"]; ok && v.(string) != "" {
				config.DedicatedMasterType = aws.String(v.(string))
			}
		}
	}

	if v, ok := m[names.AttrInstanceCount]; ok {
		config.InstanceCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := m[names.AttrInstanceType]; ok {
		config.InstanceType = aws.String(v.(string))
	}

	if v, ok := m["multi_az_with_standby_enabled"]; ok {
		config.MultiAZWithStandbyEnabled = aws.Bool(v.(bool))
	}

	if v, ok := m["warm_enabled"]; ok {
		isEnabled := v.(bool)
		config.WarmEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["warm_count"]; ok {
				config.WarmCount = aws.Int64(int64(v.(int)))
			}

			if v, ok := m["warm_type"]; ok {
				config.WarmType = aws.String(v.(string))
			}
		}
	}

	if v, ok := m["zone_awareness_enabled"]; ok {
		isEnabled := v.(bool)
		config.ZoneAwarenessEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := m["zone_awareness_config"]; ok {
				config.ZoneAwarenessConfig = expandZoneAwarenessConfig(v.([]interface{}))
			}
		}
	}

	return &config
}

func expandZoneAwarenessConfig(l []interface{}) *opensearchservice.ZoneAwarenessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	zoneAwarenessConfig := &opensearchservice.ZoneAwarenessConfig{}

	if v, ok := m["availability_zone_count"]; ok && v.(int) > 0 {
		zoneAwarenessConfig.AvailabilityZoneCount = aws.Int64(int64(v.(int)))
	}

	return zoneAwarenessConfig
}

func expandColdStorageOptions(l []interface{}) *opensearchservice.ColdStorageOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	ColdStorageOptions := &opensearchservice.ColdStorageOptions{}

	if v, ok := m[names.AttrEnabled]; ok {
		ColdStorageOptions.Enabled = aws.Bool(v.(bool))
	}

	return ColdStorageOptions
}

func flattenClusterConfig(c *opensearchservice.ClusterConfig) []map[string]interface{} {
	m := map[string]interface{}{
		"zone_awareness_config":  flattenZoneAwarenessConfig(c.ZoneAwarenessConfig),
		"zone_awareness_enabled": aws.BoolValue(c.ZoneAwarenessEnabled),
	}

	if c.ColdStorageOptions != nil {
		m["cold_storage_options"] = flattenColdStorageOptions(c.ColdStorageOptions)
	}
	if c.DedicatedMasterCount != nil {
		m["dedicated_master_count"] = aws.Int64Value(c.DedicatedMasterCount)
	}
	if c.DedicatedMasterEnabled != nil {
		m["dedicated_master_enabled"] = aws.BoolValue(c.DedicatedMasterEnabled)
	}
	if c.DedicatedMasterType != nil {
		m["dedicated_master_type"] = aws.StringValue(c.DedicatedMasterType)
	}
	if c.InstanceCount != nil {
		m[names.AttrInstanceCount] = aws.Int64Value(c.InstanceCount)
	}
	if c.InstanceType != nil {
		m[names.AttrInstanceType] = aws.StringValue(c.InstanceType)
	}
	if c.MultiAZWithStandbyEnabled != nil {
		m["multi_az_with_standby_enabled"] = aws.BoolValue(c.MultiAZWithStandbyEnabled)
	}
	if c.WarmEnabled != nil {
		m["warm_enabled"] = aws.BoolValue(c.WarmEnabled)
	}
	if c.WarmCount != nil {
		m["warm_count"] = aws.Int64Value(c.WarmCount)
	}
	if c.WarmType != nil {
		m["warm_type"] = aws.StringValue(c.WarmType)
	}

	return []map[string]interface{}{m}
}

func flattenZoneAwarenessConfig(zoneAwarenessConfig *opensearchservice.ZoneAwarenessConfig) []interface{} {
	if zoneAwarenessConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"availability_zone_count": aws.Int64Value(zoneAwarenessConfig.AvailabilityZoneCount),
	}

	return []interface{}{m}
}

func flattenColdStorageOptions(coldStorageOptions *opensearchservice.ColdStorageOptions) []interface{} {
	if coldStorageOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrEnabled: aws.BoolValue(coldStorageOptions.Enabled),
	}

	return []interface{}{m}
}

// advancedOptionsIgnoreDefault checks for defaults in the n map and, if
// they don't exist in the o map, it deletes them. AWS returns default advanced
// options that cause perpetual diffs.
func advancedOptionsIgnoreDefault(o map[string]interface{}, n map[string]interface{}) map[string]interface{} {
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

// ParseEngineVersion parses a domain's engine version string into engine type and semver string.
// engine_version is a string of format Elasticsearch_X.Y or OpenSearch_X.Y.
func ParseEngineVersion(engineVersion string) (string, string, error) {
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
func EBSVolumeTypePermitsIopsInput(volumeType string) bool {
	permittedTypes := []string{opensearchservice.VolumeTypeGp3, opensearchservice.VolumeTypeIo1}
	for _, t := range permittedTypes {
		if volumeType == t {
			return true
		}
	}
	return false
}

// EBSVolumeTypePermitsThroughputInput returns true if the volume type supports the Throughput input
//
// This check prevents a ValidationException when updating EBS volume types from a value
// that supports Throughput (ex. gp3) to one that doesn't (ex. gp2).
func EBSVolumeTypePermitsThroughputInput(volumeType string) bool {
	permittedTypes := []string{opensearchservice.VolumeTypeGp3}
	for _, t := range permittedTypes {
		if volumeType == t {
			return true
		}
	}
	return false
}

func domainErrorRetryable(err error) (bool, error) {
	switch {
	case tfawserr.ErrMessageContains(err, "InvalidTypeException", "Error setting policy"),
		tfawserr.ErrMessageContains(err, "ValidationException", "enable a service-linked role to give Amazon ES permissions"),
		tfawserr.ErrMessageContains(err, "ValidationException", "Domain is still being deleted"),
		tfawserr.ErrMessageContains(err, "ValidationException", "Amazon OpenSearch Service must be allowed to use the passed role"),
		tfawserr.ErrMessageContains(err, "ValidationException", "The passed role has not propagated yet"),
		tfawserr.ErrMessageContains(err, "ValidationException", "Authentication error"),
		tfawserr.ErrMessageContains(err, "ValidationException", "Unauthorized Operation: OpenSearch Service must be authorised to describe"),
		tfawserr.ErrMessageContains(err, "ValidationException", "The passed role must authorize Amazon OpenSearch Service to describe"):
		return true, err

	default:
		return false, err
	}
}
