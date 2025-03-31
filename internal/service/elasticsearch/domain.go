// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	elasticsearch "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elasticsearch_domain", name="Domain")
// @Tags(identifierAttribute="id")
func resourceDomain() *schema.Resource {
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
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf("elasticsearch_version", func(ctx context.Context, d *schema.ResourceDiff, meta any) bool {
				newVersion := d.Get("elasticsearch_version").(string)
				domainName := d.Get(names.AttrDomainName).(string)

				conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)
				resp, err := conn.GetCompatibleElasticsearchVersions(ctx, &elasticsearch.GetCompatibleElasticsearchVersionsInput{
					DomainName: aws.String(domainName),
				})
				if err != nil {
					log.Printf("[ERROR] Failed to get compatible Elasticsearch versions %s", domainName)
					return false
				}
				if len(resp.CompatibleElasticsearchVersions) != 1 {
					return true
				}
				return !slices.Contains(resp.CompatibleElasticsearchVersions[0].TargetVersions, newVersion)
			}),
			customdiff.ForceNewIf("encrypt_at_rest.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				// cannot disable (at all) or enable if < 6.7 without forcenew
				o, n := d.GetChange("encrypt_at_rest.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get("elasticsearch_version").(string))
			}),
			customdiff.ForceNewIf("node_to_node_encryption.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta any) bool {
				o, n := d.GetChange("node_to_node_encryption.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get("elasticsearch_version").(string))
			}),
		),

		Schema: map[string]*schema.Schema{
			"access_policies": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
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
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
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
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AutoTuneDesiredState](),
						},
						"maintenance_schedule": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
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
							DiffSuppressFunc: isDedicatedMasterDisabled,
						},
						"dedicated_master_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"dedicated_master_type": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: isDedicatedMasterDisabled,
						},
						names.AttrInstanceCount: {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.ESPartitionInstanceTypeM3MediumElasticsearch,
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
							ValidateFunc: validation.StringInSlice(warmPartitionInstanceType_Values(), false),
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
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.5",
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
			"kibana_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
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
						},
						names.AttrSecurityGroupIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceDomainCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state.
	name := d.Get(names.AttrDomainName).(string)
	_, err := findDomainByName(ctx, conn, name)

	if err == nil {
		return sdkdiag.AppendErrorf(diags, "Elasticsearch Domain (%s) already exists", name)
	}

	input := &elasticsearch.CreateElasticsearchDomainInput{
		DomainName:           aws.String(name),
		ElasticsearchVersion: aws.String(d.Get("elasticsearch_version").(string)),
		TagList:              getTagsIn(ctx),
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

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]any)) > 0 {
		input.AutoTuneOptions = expandAutoTuneOptionsInput(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside cluster_config")
			}

			input.ElasticsearchClusterConfig = expandElasticsearchClusterConfig(v[0].(map[string]any))
		}
	}

	if v, ok := d.GetOk("cognito_options"); ok {
		input.CognitoOptions = expandCognitoOptions(v.([]any))
	}

	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		input.DomainEndpointOptions = expandDomainEndpointOptions(v.([]any))
	}

	if v, ok := d.GetOk("ebs_options"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside ebs_options")
			}

			input.EBSOptions = expandEBSOptions(v[0].(map[string]any))
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside encrypt_at_rest")
			}

			input.EncryptionAtRestOptions = expandEncryptAtRestOptions(v[0].(map[string]any))
		}
	}

	if v, ok := d.GetOk("log_publishing_options"); ok {
		input.LogPublishingOptions = expandLogPublishingOptions(v.(*schema.Set))
	}

	if v, ok := d.GetOk("node_to_node_encryption"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside node_to_node_encryption")
			}

			input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(v[0].(map[string]any))
		}
	}

	if v, ok := d.GetOk("snapshot_options"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside snapshot_options")
			}

			tfMap := v[0].(map[string]any)

			input.SnapshotOptions = &awstypes.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int32(int32(tfMap["automated_snapshot_start_hour"].(int))),
			}
		}
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		if v := v.([]any); len(v) == 1 {
			if v[0] == nil {
				return sdkdiag.AppendErrorf(diags, "At least one field is expected inside vpc_options")
			}

			input.VPCOptions = expandVPCOptions(v[0].(map[string]any))
		}
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (any, error) {
			return conn.CreateElasticsearchDomain(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.InvalidTypeException](err, "Error setting policy") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "enable a service-linked role to give Amazon ES permissions") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Domain is still being deleted") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Amazon Elasticsearch must be allowed to use the passed role") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "The passed role has not propagated yet") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Authentication error") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Unauthorized Operation: Elasticsearch must be authorised to describe") ||
				errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "The passed role must authorize Amazon Elasticsearch to describe") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Elasticsearch Domain (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*elasticsearch.CreateElasticsearchDomainOutput).DomainStatus.ARN))

	if _, err := waitDomainCreated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]any)) > 0 {
		input := &elasticsearch.UpdateElasticsearchDomainConfigInput{
			AutoTuneOptions: expandAutoTuneOptions(v.([]any)[0].(map[string]any)),
			DomainName:      aws.String(name),
		}

		_, err = conn.UpdateElasticsearchDomainConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elasticsearch Domain (%s) Config: %s", d.Id(), err)
		}

		if _, err := waitDomainConfigUpdated(ctx, conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	name := d.Get(names.AttrDomainName).(string)
	ds, err := findDomainByName(ctx, conn, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain (%s): %s", d.Id(), err)
	}

	output, err := conn.DescribeElasticsearchDomainConfig(ctx, &elasticsearch.DescribeElasticsearchDomainConfigInput{
		DomainName: aws.String(name),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Elasticsearch Domain (%s) config: %s", d.Id(), err)
	}

	dc := output.DomainConfig

	if v := aws.ToString(ds.AccessPolicies); v != "" {
		policies, err := verify.PolicyToSet(d.Get("access_policies").(string), v)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		d.Set("access_policies", policies)
	}
	if err := d.Set("advanced_options", advancedOptionsIgnoreDefault(d.Get("advanced_options").(map[string]any), flex.FlattenStringValueMap(ds.AdvancedOptions))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting advanced_options: %s", err)
	}
	// Populate AdvancedSecurityOptions with values returned from
	// DescribeElasticsearchDomainConfig, if enabled, else use
	// values from resource; additionally, append MasterUserOptions
	// from resource as they are not returned from the API
	if ds.AdvancedSecurityOptions != nil {
		advSecOpts := flattenAdvancedSecurityOptions(ds.AdvancedSecurityOptions)
		if !aws.ToBool(ds.AdvancedSecurityOptions.Enabled) {
			advSecOpts[0]["internal_user_database_enabled"] = getUserDBEnabled(d)
		}
		advSecOpts[0]["master_user_options"] = getMasterUserOptions(d)

		if err := d.Set("advanced_security_options", advSecOpts); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting advanced_security_options: %s", err)
		}
	}
	d.Set(names.AttrARN, ds.ARN)
	if v := dc.AutoTuneOptions; v != nil {
		if err := d.Set("auto_tune_options", []any{flattenAutoTuneOptions(v.Options)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting auto_tune_options: %s", err)
		}
	}
	if err := d.Set("cluster_config", flattenElasticsearchClusterConfig(ds.ElasticsearchClusterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cluster_config: %s", err)
	}
	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cognito_options: %s", err)
	}
	if err := d.Set("domain_endpoint_options", flattenDomainEndpointOptions(ds.DomainEndpointOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_endpoint_options: %s", err)
	}
	d.Set("domain_id", ds.DomainId)
	d.Set(names.AttrDomainName, ds.DomainName)
	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_options: %s", err)
	}
	d.Set("elasticsearch_version", ds.ElasticsearchVersion)
	if err := d.Set("encrypt_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encrypt_at_rest: %s", err)
	}
	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_publishing_options: %s", err)
	}
	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting node_to_node_encryption: %s", err)
	}
	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting snapshot_options: %s", err)
	}
	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", []any{flattenVPCDerivedInfo(ds.VPCOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_options: %s", err)
		}

		endpoints := flex.FlattenStringValueMap(ds.Endpoints)
		d.Set(names.AttrEndpoint, endpoints["vpc"])

		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return sdkdiag.AppendErrorf(diags, "%q: Elasticsearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set(names.AttrEndpoint, ds.Endpoint)
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return sdkdiag.AppendErrorf(diags, "%q: Elasticsearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	return diags
}

func resourceDomainUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		name := d.Get(names.AttrDomainName).(string)
		input := &elasticsearch.UpdateElasticsearchDomainConfigInput{
			DomainName: aws.String(name),
		}

		if d.HasChange("access_policies") {
			o, n := d.GetChange("access_policies")

			if equivalent, err := awspolicy.PoliciesAreEquivalent(o.(string), n.(string)); err != nil || !equivalent {
				input.AccessPolicies = aws.String(d.Get("access_policies").(string))
			}
		}

		if d.HasChange("advanced_options") {
			input.AdvancedOptions = flex.ExpandStringValueMap(d.Get("advanced_options").(map[string]any))
		}

		if d.HasChange("advanced_security_options") {
			input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(d.Get("advanced_security_options").([]any))
		}

		if d.HasChange("auto_tune_options") {
			input.AutoTuneOptions = expandAutoTuneOptions(d.Get("auto_tune_options").([]any)[0].(map[string]any))
		}

		if d.HasChange("domain_endpoint_options") {
			input.DomainEndpointOptions = expandDomainEndpointOptions(d.Get("domain_endpoint_options").([]any))
		}

		if d.HasChanges("ebs_options", "cluster_config") {
			if v := d.Get("ebs_options").([]any); len(v) == 1 {
				input.EBSOptions = expandEBSOptions(v[0].(map[string]any))
			}

			if d.HasChange("cluster_config") {
				if v := d.Get("cluster_config").([]any); len(v) == 1 {
					input.ElasticsearchClusterConfig = expandElasticsearchClusterConfig(v[0].(map[string]any))

					// Work around "ValidationException: Your domain's Elasticsearch version does not support cold storage options. Upgrade to Elasticsearch 7.9 or later.".
					if semver.LessThan(d.Get("elasticsearch_version").(string), "7.9") {
						input.ElasticsearchClusterConfig.ColdStorageOptions = nil
					}
				}
			}
		}

		if d.HasChange("encrypt_at_rest") {
			input.EncryptionAtRestOptions = nil
			if v, ok := d.GetOk("encrypt_at_rest"); ok {
				v := v.([]any)
				if v[0] == nil {
					return sdkdiag.AppendErrorf(diags, "At least one field is expected inside encrypt_at_rest")
				}

				input.EncryptionAtRestOptions = expandEncryptAtRestOptions(v[0].(map[string]any))
			}
		}

		if d.HasChange("node_to_node_encryption") {
			input.NodeToNodeEncryptionOptions = nil
			if v, ok := d.GetOk("node_to_node_encryption"); ok {
				v := v.([]any)
				input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(v[0].(map[string]any))
			}
		}

		if d.HasChange("snapshot_options") {
			if v := d.Get("snapshot_options").([]any); len(v) == 1 {
				tfMap := v[0].(map[string]any)

				snapshotOptions := &awstypes.SnapshotOptions{
					AutomatedSnapshotStartHour: aws.Int32(int32(tfMap["automated_snapshot_start_hour"].(int))),
				}

				input.SnapshotOptions = snapshotOptions
			}
		}

		if d.HasChange("vpc_options") {
			v := d.Get("vpc_options").([]any)
			input.VPCOptions = expandVPCOptions(v[0].(map[string]any))
		}

		if d.HasChange("cognito_options") {
			input.CognitoOptions = expandCognitoOptions(d.Get("cognito_options").([]any))
		}

		if d.HasChange("log_publishing_options") {
			input.LogPublishingOptions = expandLogPublishingOptions(d.Get("log_publishing_options").(*schema.Set))
		}

		_, err := conn.UpdateElasticsearchDomainConfig(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Elasticsearch Domain (%s) Config: %s", d.Id(), err)
		}

		if _, err := waitDomainConfigUpdated(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) Config update: %s", d.Id(), err)
		}

		if d.HasChange("elasticsearch_version") {
			input := &elasticsearch.UpgradeElasticsearchDomainInput{
				DomainName:    aws.String(name),
				TargetVersion: aws.String(d.Get("elasticsearch_version").(string)),
			}

			_, err := conn.UpgradeElasticsearchDomain(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "upgrading Elasticsearch Domain (%s): %s", d.Id(), err)
			}

			if _, err := waitUpgradeSucceeded(ctx, conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) upgrade: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDomainRead(ctx, d, meta)...)
}

func resourceDomainDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	name := d.Get(names.AttrDomainName).(string)

	log.Printf("[DEBUG] Deleting Elasticsearch Domain: %s", d.Id())
	_, err := conn.DeleteElasticsearchDomain(ctx, &elasticsearch.DeleteElasticsearchDomainInput{
		DomainName: aws.String(name),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Elasticsearch Domain (%s): %s", d.Id(), err)
	}

	if _, err := waitDomainDeleted(ctx, conn, name, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Elasticsearch Domain (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func resourceDomainImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).ElasticsearchClient(ctx)

	d.Set(names.AttrDomainName, d.Id())

	ds, err := findDomainByName(ctx, conn, d.Get(names.AttrDomainName).(string))

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(ds.ARN))

	return []*schema.ResourceData{d}, nil
}

func findDomainByName(ctx context.Context, conn *elasticsearch.Client, name string) (*awstypes.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(name),
	}
	output, err := findDomain(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.Deleted) {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findDomain(ctx context.Context, conn *elasticsearch.Client, input *elasticsearch.DescribeElasticsearchDomainInput) (*awstypes.ElasticsearchDomainStatus, error) {
	output, err := conn.DescribeElasticsearchDomain(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findDomainConfigByName(ctx context.Context, conn *elasticsearch.Client, name string) (*awstypes.ElasticsearchDomainConfig, error) {
	input := &elasticsearch.DescribeElasticsearchDomainConfigInput{
		DomainName: aws.String(name),
	}

	return findDomainConfig(ctx, conn, input)
}

func findDomainConfig(ctx context.Context, conn *elasticsearch.Client, input *elasticsearch.DescribeElasticsearchDomainConfigInput) (*awstypes.ElasticsearchDomainConfig, error) {
	output, err := conn.DescribeElasticsearchDomainConfig(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DomainConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DomainConfig, nil
}

func findDomainUpgradeStatusByName(ctx context.Context, conn *elasticsearch.Client, name string) (*elasticsearch.GetUpgradeStatusOutput, error) {
	input := &elasticsearch.GetUpgradeStatusInput{
		DomainName: aws.String(name),
	}

	return findDomainUpgradeStatus(ctx, conn, input)
}

func findDomainUpgradeStatus(ctx context.Context, conn *elasticsearch.Client, input *elasticsearch.GetUpgradeStatusInput) (*elasticsearch.GetUpgradeStatusOutput, error) {
	output, err := conn.GetUpgradeStatus(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusDomainProcessing(ctx context.Context, conn *elasticsearch.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		// Don't call findDomainByName here as the domain's Deleted flag will be set while DomainProcessingStatus is "Deleting".
		input := &elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(name),
		}
		output, err := findDomain(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DomainProcessingStatus), nil
	}
}

func statusDomainUpgrade(ctx context.Context, conn *elasticsearch.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDomainUpgradeStatusByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// Elasticsearch upgrades consist of multiple steps:
		// https://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/es-version-migration.html
		// Prevent false positive completion where the UpgradeStep is not the final UPGRADE step.
		status := output.StepStatus
		if status == awstypes.UpgradeStatusSucceeded && output.UpgradeStep != awstypes.UpgradeStepUpgrade {
			status = awstypes.UpgradeStatusInProgress
		}

		return output, string(status), nil
	}
}

func waitUpgradeSucceeded(ctx context.Context, conn *elasticsearch.Client, name string, timeout time.Duration) (*elasticsearch.GetUpgradeStatusOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.UpgradeStatusInProgress),
		Target:     enum.Slice(awstypes.UpgradeStatusSucceeded, awstypes.UpgradeStatusSucceededWithIssues),
		Refresh:    statusDomainUpgrade(ctx, conn, name),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elasticsearch.GetUpgradeStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDomainCreated(ctx context.Context, conn *elasticsearch.Client, domainName string, timeout time.Duration) (*awstypes.ElasticsearchDomainStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainProcessingStatusTypeCreating),
		Target:  enum.Slice(awstypes.DomainProcessingStatusTypeActive),
		Refresh: statusDomainProcessing(ctx, conn, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ElasticsearchDomainStatus); ok {
		if v := output.ChangeProgressDetails; v != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", v.ConfigChangeStatus, aws.ToString(v.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDomainConfigUpdated(ctx context.Context, conn *elasticsearch.Client, domainName string, timeout time.Duration) (*awstypes.ElasticsearchDomainStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainProcessingStatusTypeModifying),
		Target:  enum.Slice(awstypes.DomainProcessingStatusTypeActive),
		Refresh: statusDomainProcessing(ctx, conn, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ElasticsearchDomainStatus); ok {
		if v := output.ChangeProgressDetails; v != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", v.ConfigChangeStatus, aws.ToString(v.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDomainDeleted(ctx context.Context, conn *elasticsearch.Client, domainName string, timeout time.Duration) (*awstypes.ElasticsearchDomainStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainProcessingStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusDomainProcessing(ctx, conn, domainName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ElasticsearchDomainStatus); ok {
		if v := output.ChangeProgressDetails; v != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", v.ConfigChangeStatus, aws.ToString(v.Message)))
		}

		return output, err
	}

	return nil, err
}

func warmPartitionInstanceType_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.ESWarmPartitionInstanceType](), "ultrawarm1.xlarge.elasticsearch")
}

// inPlaceEncryptionEnableVersion returns true if, based on version, encryption
// can be enabled in place (without ForceNew)
func inPlaceEncryptionEnableVersion(version string) bool {
	return semver.GreaterThanOrEqual(version, "6.7")
}

func suppressEquivalentKMSKeyIDs(k, old, new string, d *schema.ResourceData) bool {
	// The Elasticsearch API accepts a short KMS key id but always returns the ARN of the key.
	// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'.
	// These should be treated as equivalent.
	return strings.Contains(old, new)
}

func getKibanaEndpoint(d *schema.ResourceData) string {
	return d.Get(names.AttrEndpoint).(string) + "/_plugin/kibana/"
}

func isDedicatedMasterDisabled(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.GetOk("cluster_config"); ok {
		tfMap := v.([]any)[0].(map[string]any)
		return !tfMap["dedicated_master_enabled"].(bool)
	}
	return false
}

func isCustomEndpointDisabled(k, old, new string, d *schema.ResourceData) bool {
	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		tfMap := v.([]any)[0].(map[string]any)
		return !tfMap["custom_endpoint_enabled"].(bool)
	}
	return false
}

func expandNodeToNodeEncryptionOptions(tfMap map[string]any) *awstypes.NodeToNodeEncryptionOptions {
	apiObject := &awstypes.NodeToNodeEncryptionOptions{}

	if v, ok := tfMap[names.AttrEnabled]; ok {
		apiObject.Enabled = aws.Bool(v.(bool))
	}

	return apiObject
}

func flattenNodeToNodeEncryptionOptions(apiObject *awstypes.NodeToNodeEncryptionOptions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{}

	if apiObject.Enabled != nil {
		tfMap[names.AttrEnabled] = aws.ToBool(apiObject.Enabled)
	}

	return []any{tfMap}
}

func expandElasticsearchClusterConfig(tfMap map[string]any) *awstypes.ElasticsearchClusterConfig { // nosemgrep:ci.elasticsearch-in-func-name
	apiObject := &awstypes.ElasticsearchClusterConfig{}

	if v, ok := tfMap["cold_storage_options"].([]any); ok && len(v) > 0 {
		apiObject.ColdStorageOptions = expandColdStorageOptions(v[0].(map[string]any))
	}

	if v, ok := tfMap["dedicated_master_enabled"]; ok {
		isEnabled := v.(bool)
		apiObject.DedicatedMasterEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := tfMap["dedicated_master_count"]; ok && v.(int) > 0 {
				apiObject.DedicatedMasterCount = aws.Int32(int32(v.(int)))
			}
			if v, ok := tfMap["dedicated_master_type"]; ok && v.(string) != "" {
				apiObject.DedicatedMasterType = awstypes.ESPartitionInstanceType(v.(string))
			}
		}
	}

	if v, ok := tfMap[names.AttrInstanceCount]; ok {
		apiObject.InstanceCount = aws.Int32(int32(v.(int)))
	}
	if v, ok := tfMap[names.AttrInstanceType]; ok {
		apiObject.InstanceType = awstypes.ESPartitionInstanceType(v.(string))
	}

	if v, ok := tfMap["zone_awareness_enabled"]; ok {
		isEnabled := v.(bool)
		apiObject.ZoneAwarenessEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := tfMap["zone_awareness_config"]; ok {
				apiObject.ZoneAwarenessConfig = expandZoneAwarenessConfig(v.([]any))
			}
		}
	}

	if v, ok := tfMap["warm_enabled"]; ok {
		isEnabled := v.(bool)
		apiObject.WarmEnabled = aws.Bool(isEnabled)

		if isEnabled {
			if v, ok := tfMap["warm_count"]; ok {
				apiObject.WarmCount = aws.Int32(int32(v.(int)))
			}

			if v, ok := tfMap["warm_type"]; ok {
				apiObject.WarmType = awstypes.ESWarmPartitionInstanceType(v.(string))
			}
		}
	}

	return apiObject
}

func expandColdStorageOptions(tfMap map[string]any) *awstypes.ColdStorageOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ColdStorageOptions{}

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandZoneAwarenessConfig(tfList []any) *awstypes.ZoneAwarenessConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)
	apiObject := &awstypes.ZoneAwarenessConfig{}

	if v, ok := tfMap["availability_zone_count"]; ok && v.(int) > 0 {
		apiObject.AvailabilityZoneCount = aws.Int32(int32(v.(int)))
	}

	return apiObject
}

func flattenElasticsearchClusterConfig(apiObject *awstypes.ElasticsearchClusterConfig) []any { // nosemgrep:ci.elasticsearch-in-func-name
	tfMap := map[string]any{
		"zone_awareness_config":  flattenZoneAwarenessConfig(apiObject.ZoneAwarenessConfig),
		"zone_awareness_enabled": aws.ToBool(apiObject.ZoneAwarenessEnabled),
	}

	if apiObject.ColdStorageOptions != nil {
		tfMap["cold_storage_options"] = flattenColdStorageOptions(apiObject.ColdStorageOptions)
	}
	if apiObject.DedicatedMasterCount != nil {
		tfMap["dedicated_master_count"] = aws.ToInt32(apiObject.DedicatedMasterCount)
	}
	if apiObject.DedicatedMasterEnabled != nil {
		tfMap["dedicated_master_enabled"] = aws.ToBool(apiObject.DedicatedMasterEnabled)
	}
	tfMap["dedicated_master_type"] = apiObject.DedicatedMasterType
	if apiObject.InstanceCount != nil {
		tfMap[names.AttrInstanceCount] = aws.ToInt32(apiObject.InstanceCount)
	}
	tfMap[names.AttrInstanceType] = apiObject.InstanceType
	if apiObject.WarmEnabled != nil {
		tfMap["warm_enabled"] = aws.ToBool(apiObject.WarmEnabled)
	}
	if apiObject.WarmCount != nil {
		tfMap["warm_count"] = aws.ToInt32(apiObject.WarmCount)
	}
	tfMap["warm_type"] = string(apiObject.WarmType)

	return []any{tfMap}
}

func flattenColdStorageOptions(apiObject *awstypes.ColdStorageOptions) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		names.AttrEnabled: aws.ToBool(apiObject.Enabled),
	}

	return []any{tfMap}
}

func flattenZoneAwarenessConfig(apiObject *awstypes.ZoneAwarenessConfig) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"availability_zone_count": aws.ToInt32(apiObject.AvailabilityZoneCount),
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

// ebsVolumeTypePermitsIopsInput returns true if the volume type supports the Iops input
//
// This check prevents a ValidationException when updating EBS volume types from a value
// that supports IOPS (ex. gp3) to one that doesn't (ex. gp2).
func ebsVolumeTypePermitsIopsInput(volumeType string) bool {
	permittedTypes := enum.Slice(awstypes.VolumeTypeGp3, awstypes.VolumeTypeIo1)
	return slices.Contains(permittedTypes, volumeType)
}

// ebsVolumeTypePermitsThroughputInput returns true if the volume type supports the Throughput input
//
// This check prevents a ValidationException when updating EBS volume types from a value
// that supports Throughput (ex. gp3) to one that doesn't (ex. gp2).
func ebsVolumeTypePermitsThroughputInput(volumeType string) bool {
	permittedTypes := enum.Slice(awstypes.VolumeTypeGp3)
	return slices.Contains(permittedTypes, volumeType)
}
