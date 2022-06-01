package elasticsearch

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainCreate,
		Read:   resourceDomainRead,
		Update: resourceDomainUpdate,
		Delete: resourceDomainDelete,

		Importer: &schema.ResourceImporter{
			State: resourceDomainImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(90 * time.Minute),
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIf("elasticsearch_version", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				newVersion := d.Get("elasticsearch_version").(string)
				domainName := d.Get("domain_name").(string)

				conn := meta.(*conns.AWSClient).ElasticsearchConn
				resp, err := conn.GetCompatibleElasticsearchVersions(&elasticsearch.GetCompatibleElasticsearchVersionsInput{
					DomainName: aws.String(domainName),
				})
				if err != nil {
					log.Printf("[ERROR] Failed to get compatible Elasticsearch versions %s", domainName)
					return false
				}
				if len(resp.CompatibleElasticsearchVersions) != 1 {
					return true
				}
				for _, targetVersion := range resp.CompatibleElasticsearchVersions[0].TargetVersions {
					if aws.StringValue(targetVersion) == newVersion {
						return false
					}
				}
				return true
			}),
			customdiff.ForceNewIf("encrypt_at_rest.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				// cannot disable (at all) or enable if < 6.7 without forcenew
				o, n := d.GetChange("encrypt_at_rest.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get("elasticsearch_version").(string))
			}),
			customdiff.ForceNewIf("node_to_node_encryption.0.enabled", func(_ context.Context, d *schema.ResourceDiff, meta interface{}) bool {
				o, n := d.GetChange("node_to_node_encryption.0.enabled")
				if o.(bool) && !n.(bool) {
					return true
				}

				return !inPlaceEncryptionEnableVersion(d.Get("elasticsearch_version").(string))
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
						"enabled": {
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
			"arn": {
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
							ValidateFunc: validation.StringInSlice(elasticsearch.AutoTuneDesiredState_Values(), false),
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
									"duration": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"unit": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(elasticsearch.TimeUnit_Values(), false),
												},
												"value": {
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
							ValidateFunc: validation.StringInSlice(elasticsearch.RollbackOnDisable_Values(), false),
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
									"enabled": {
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
						"instance_count": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  elasticsearch.ESPartitionInstanceTypeM3MediumElasticsearch,
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
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								elasticsearch.ESWarmPartitionInstanceTypeUltrawarm1MediumElasticsearch,
								elasticsearch.ESWarmPartitionInstanceTypeUltrawarm1LargeElasticsearch,
								"ultrawarm1.xlarge.elasticsearch",
							}, false),
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
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"user_pool_id": {
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][0-9a-z\-]{2,27}$`),
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
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(elasticsearch.TLSSecurityPolicy_Values(), false),
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
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"volume_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(elasticsearch.VolumeType_Values(), false),
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
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"kms_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressEquivalentKMSKeyIDs,
						},
					},
				},
			},
			"endpoint": {
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
						"cloudwatch_log_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"log_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticsearch.LogType_Values(), false),
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
						"enabled": {
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_options": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state.
	name := d.Get("domain_name").(string)
	_, err := FindDomainByName(conn, name)

	if err == nil {
		return fmt.Errorf("Elasticsearch Domain (%s) already exists", name)
	}

	input := &elasticsearch.CreateElasticsearchDomainInput{
		DomainName:           aws.String(name),
		ElasticsearchVersion: aws.String(d.Get("elasticsearch_version").(string)),
		TagList:              Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("access_policies"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))

		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
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
				return fmt.Errorf("At least one field is expected inside ebs_options")
			}

			s := options[0].(map[string]interface{})
			input.EBSOptions = expandEBSOptions(s)
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside encrypt_at_rest")
		}

		s := options[0].(map[string]interface{})
		input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		config := v.([]interface{})

		if len(config) == 1 {
			if config[0] == nil {
				return fmt.Errorf("At least one field is expected inside cluster_config")
			}
			m := config[0].(map[string]interface{})
			input.ElasticsearchClusterConfig = expandClusterConfig(m)
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
				return fmt.Errorf("At least one field is expected inside snapshot_options")
			}

			o := options[0].(map[string]interface{})

			snapshotOptions := elasticsearch.SnapshotOptions{
				AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
			}

			input.SnapshotOptions = &snapshotOptions
		}
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside vpc_options")
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

	log.Printf("[DEBUG] Creating Elasticsearch Domain: %s", input)

	outputRaw, err := tfresource.RetryWhen(
		propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateElasticsearchDomain(input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeInvalidTypeException, "Error setting policy") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "enable a service-linked role to give Amazon ES permissions") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "Domain is still being deleted") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "Amazon Elasticsearch must be allowed to use the passed role") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "The passed role has not propagated yet") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "Authentication error") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "Unauthorized Operation: Elasticsearch must be authorised to describe") ||
				tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeValidationException, "The passed role must authorize Amazon Elasticsearch to describe") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return fmt.Errorf("error creating Elasticsearch Domain (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*elasticsearch.CreateElasticsearchDomainOutput).DomainStatus.ARN))

	if err := WaitForDomainCreation(conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain (%s) create: %w", d.Id(), err)
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]interface{})) > 0 {
		input := &elasticsearch.UpdateElasticsearchDomainConfigInput{
			AutoTuneOptions: expandAutoTuneOptions(v.([]interface{})[0].(map[string]interface{})),
			DomainName:      aws.String(name),
		}

		log.Printf("[DEBUG] Updating Elasticsearch Domain config: %s", input)
		_, err = conn.UpdateElasticsearchDomainConfig(input)

		if err != nil {
			return fmt.Errorf("error updating Elasticsearch Domain (%s) config: %w", d.Id(), err)
		}

		if err := waitForDomainUpdate(conn, name, d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("error waiting for Elasticsearch Domain (%s) update: %w", d.Id(), err)
		}
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("domain_name").(string)
	ds, err := FindDomainByName(conn, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch Domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Elasticsearch Domain (%s): %w", d.Id(), err)
	}

	output, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
		DomainName: aws.String(name),
	})

	if err != nil {
		return fmt.Errorf("error reading Elasticsearch Domain (%s) config: %w", d.Id(), err)
	}

	dc := output.DomainConfig

	if v := aws.StringValue(ds.AccessPolicies); v != "" {
		policies, err := verify.PolicyToSet(d.Get("access_policies").(string), v)

		if err != nil {
			return err
		}

		d.Set("access_policies", policies)
	}

	options := advancedOptionsIgnoreDefault(d.Get("advanced_options").(map[string]interface{}), flex.PointersMapToStringList(ds.AdvancedOptions))
	if err = d.Set("advanced_options", options); err != nil {
		return fmt.Errorf("setting advanced_options: %w", err)
	}

	d.Set("domain_id", ds.DomainId)
	d.Set("domain_name", ds.DomainName)
	d.Set("elasticsearch_version", ds.ElasticsearchVersion)

	if err := d.Set("ebs_options", flattenEBSOptions(ds.EBSOptions)); err != nil {
		return fmt.Errorf("error setting ebs_options: %w", err)
	}

	if err := d.Set("encrypt_at_rest", flattenEncryptAtRestOptions(ds.EncryptionAtRestOptions)); err != nil {
		return fmt.Errorf("error setting encrypt_at_rest: %w", err)
	}

	if err := d.Set("cluster_config", flattenClusterConfig(ds.ElasticsearchClusterConfig)); err != nil {
		return fmt.Errorf("error setting cluster_config: %w", err)
	}

	if err := d.Set("cognito_options", flattenCognitoOptions(ds.CognitoOptions)); err != nil {
		return fmt.Errorf("error setting cognito_options: %w", err)
	}

	if err := d.Set("node_to_node_encryption", flattenNodeToNodeEncryptionOptions(ds.NodeToNodeEncryptionOptions)); err != nil {
		return fmt.Errorf("error setting node_to_node_encryption: %w", err)
	}

	// Populate AdvancedSecurityOptions with values returned from
	// DescribeElasticsearchDomainConfig, if enabled, else use
	// values from resource; additionally, append MasterUserOptions
	// from resource as they are not returned from the API
	if ds.AdvancedSecurityOptions != nil {
		advSecOpts := flattenAdvancedSecurityOptions(ds.AdvancedSecurityOptions)
		if !aws.BoolValue(ds.AdvancedSecurityOptions.Enabled) {
			advSecOpts[0]["internal_user_database_enabled"] = getUserDBEnabled(d)
		}
		advSecOpts[0]["master_user_options"] = getMasterUserOptions(d)

		if err := d.Set("advanced_security_options", advSecOpts); err != nil {
			return fmt.Errorf("error setting advanced_security_options: %w", err)
		}
	}

	if v := dc.AutoTuneOptions; v != nil {
		if err := d.Set("auto_tune_options", []interface{}{flattenAutoTuneOptions(v.Options)}); err != nil {
			return fmt.Errorf("error setting auto_tune_options: %w", err)
		}
	}

	if err := d.Set("snapshot_options", flattenSnapshotOptions(ds.SnapshotOptions)); err != nil {
		return fmt.Errorf("error setting snapshot_options: %w", err)
	}

	if ds.VPCOptions != nil {
		if err := d.Set("vpc_options", flattenVPCDerivedInfo(ds.VPCOptions)); err != nil {
			return fmt.Errorf("error setting vpc_options: %w", err)
		}

		endpoints := flex.PointersMapToStringList(ds.Endpoints)
		d.Set("endpoint", endpoints["vpc"])

		d.Set("kibana_endpoint", getKibanaEndpoint(d))
		if ds.Endpoint != nil {
			return fmt.Errorf("%q: Elasticsearch domain in VPC expected to have null Endpoint value", d.Id())
		}
	} else {
		if ds.Endpoint != nil {
			d.Set("endpoint", ds.Endpoint)
			d.Set("kibana_endpoint", getKibanaEndpoint(d))
		}
		if ds.Endpoints != nil {
			return fmt.Errorf("%q: Elasticsearch domain not in VPC expected to have null Endpoints value", d.Id())
		}
	}

	if err := d.Set("log_publishing_options", flattenLogPublishingOptions(ds.LogPublishingOptions)); err != nil {
		return fmt.Errorf("error setting log_publishing_options: %w", err)
	}

	if err := d.Set("domain_endpoint_options", flattenDomainEndpointOptions(ds.DomainEndpointOptions)); err != nil {
		return fmt.Errorf("error setting domain_endpoint_options: %w", err)
	}

	d.Set("arn", ds.ARN)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for Elasticsearch Domain (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	if d.HasChangesExcept("tags", "tags_all") {
		name := d.Get("domain_name").(string)
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
			input.AdvancedOptions = flex.ExpandStringMap(d.Get("advanced_options").(map[string]interface{}))
		}

		if d.HasChange("advanced_security_options") {
			input.AdvancedSecurityOptions = expandAdvancedSecurityOptions(d.Get("advanced_security_options").([]interface{}))
		}

		if d.HasChange("auto_tune_options") {
			input.AutoTuneOptions = expandAutoTuneOptions(d.Get("auto_tune_options").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("domain_endpoint_options") {
			input.DomainEndpointOptions = expandDomainEndpointOptions(d.Get("domain_endpoint_options").([]interface{}))
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
					input.ElasticsearchClusterConfig = expandClusterConfig(m)

					// Work around "ValidationException: Your domain's Elasticsearch version does not support cold storage options. Upgrade to Elasticsearch 7.9 or later.".
					if verify.SemVerLessThan(d.Get("elasticsearch_version").(string), "7.9") {
						input.ElasticsearchClusterConfig.ColdStorageOptions = nil
					}
				}
			}
		}

		if d.HasChange("encrypt_at_rest") {
			input.EncryptionAtRestOptions = nil
			if v, ok := d.GetOk("encrypt_at_rest"); ok {
				options := v.([]interface{})
				if options[0] == nil {
					return fmt.Errorf("At least one field is expected inside encrypt_at_rest")
				}

				s := options[0].(map[string]interface{})
				input.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
			}
		}

		if d.HasChange("node_to_node_encryption") {
			input.NodeToNodeEncryptionOptions = nil
			if v, ok := d.GetOk("node_to_node_encryption"); ok {
				options := v.([]interface{})

				s := options[0].(map[string]interface{})
				input.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
			}
		}

		if d.HasChange("snapshot_options") {
			options := d.Get("snapshot_options").([]interface{})

			if len(options) == 1 {
				o := options[0].(map[string]interface{})

				snapshotOptions := elasticsearch.SnapshotOptions{
					AutomatedSnapshotStartHour: aws.Int64(int64(o["automated_snapshot_start_hour"].(int))),
				}

				input.SnapshotOptions = &snapshotOptions
			}
		}

		if d.HasChange("vpc_options") {
			options := d.Get("vpc_options").([]interface{})
			s := options[0].(map[string]interface{})
			input.VPCOptions = expandVPCOptions(s)
		}

		if d.HasChange("cognito_options") {
			options := d.Get("cognito_options").([]interface{})
			input.CognitoOptions = expandCognitoOptions(options)
		}

		if d.HasChange("log_publishing_options") {
			input.LogPublishingOptions = expandLogPublishingOptions(d.Get("log_publishing_options").(*schema.Set))
		}

		log.Printf("[DEBUG] Updating Elasticsearch Domain config: %s", input)
		_, err := conn.UpdateElasticsearchDomainConfig(input)

		if err != nil {
			return fmt.Errorf("error updating Elasticsearch Domain (%s) config: %w", d.Id(), err)
		}

		if err := waitForDomainUpdate(conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for Elasticsearch Domain (%s) update: %w", d.Id(), err)
		}

		if d.HasChange("elasticsearch_version") {
			input := &elasticsearch.UpgradeElasticsearchDomainInput{
				DomainName:    aws.String(name),
				TargetVersion: aws.String(d.Get("elasticsearch_version").(string)),
			}

			log.Printf("[DEBUG] Upgrading Elasticsearch Domain: %s", input)
			_, err := conn.UpgradeElasticsearchDomain(input)

			if err != nil {
				return fmt.Errorf("error upgrading Elasticsearch Domain (%s): %w", d.Id(), err)
			}

			if _, err := waitUpgradeSucceeded(conn, name, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("error waiting for Elasticsearch Domain (%s) upgrade: %w", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Elasticsearch Domain (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	name := d.Get("domain_name").(string)

	log.Printf("[DEBUG] Deleting Elasticsearch Domain: %s", d.Id())
	_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
		DomainName: aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, elasticsearch.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Elasticsearch Domain (%s): %w", d.Id(), err)
	}

	if err := waitForDomainDelete(conn, name, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func resourceDomainImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).ElasticsearchConn

	d.Set("domain_name", d.Id())

	ds, err := FindDomainByName(conn, d.Get("domain_name").(string))

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(ds.ARN))

	return []*schema.ResourceData{d}, nil
}

// inPlaceEncryptionEnableVersion returns true if, based on version, encryption
// can be enabled in place (without ForceNew)
func inPlaceEncryptionEnableVersion(version string) bool {
	return verify.SemVerGreaterThanOrEqual(version, "6.7")
}

func suppressEquivalentKMSKeyIDs(k, old, new string, d *schema.ResourceData) bool {
	// The Elasticsearch API accepts a short KMS key id but always returns the ARN of the key.
	// The ARN is of the format 'arn:aws:kms:REGION:ACCOUNT_ID:key/KMS_KEY_ID'.
	// These should be treated as equivalent.
	return strings.Contains(old, new)
}

func getKibanaEndpoint(d *schema.ResourceData) string {
	return d.Get("endpoint").(string) + "/_plugin/kibana/"
}

func isDedicatedMasterDisabled(k, old, new string, d *schema.ResourceData) bool {
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

func expandNodeToNodeEncryptionOptions(s map[string]interface{}) *elasticsearch.NodeToNodeEncryptionOptions {
	options := elasticsearch.NodeToNodeEncryptionOptions{}

	if v, ok := s["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	return &options
}

func flattenNodeToNodeEncryptionOptions(o *elasticsearch.NodeToNodeEncryptionOptions) []map[string]interface{} {
	if o == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}
	if o.Enabled != nil {
		m["enabled"] = aws.BoolValue(o.Enabled)
	}

	return []map[string]interface{}{m}
}

func expandClusterConfig(m map[string]interface{}) *elasticsearch.ElasticsearchClusterConfig {
	config := elasticsearch.ElasticsearchClusterConfig{}

	if v, ok := m["cold_storage_options"].([]interface{}); ok && len(v) > 0 {
		config.ColdStorageOptions = expandColdStorageOptions(v[0].(map[string]interface{}))
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

	if v, ok := m["instance_count"]; ok {
		config.InstanceCount = aws.Int64(int64(v.(int)))
	}
	if v, ok := m["instance_type"]; ok {
		config.InstanceType = aws.String(v.(string))
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

	return &config
}

func expandColdStorageOptions(tfMap map[string]interface{}) *elasticsearch.ColdStorageOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &elasticsearch.ColdStorageOptions{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func expandZoneAwarenessConfig(l []interface{}) *elasticsearch.ZoneAwarenessConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	zoneAwarenessConfig := &elasticsearch.ZoneAwarenessConfig{}

	if v, ok := m["availability_zone_count"]; ok && v.(int) > 0 {
		zoneAwarenessConfig.AvailabilityZoneCount = aws.Int64(int64(v.(int)))
	}

	return zoneAwarenessConfig
}

func flattenClusterConfig(c *elasticsearch.ElasticsearchClusterConfig) []map[string]interface{} {
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
		m["instance_count"] = aws.Int64Value(c.InstanceCount)
	}
	if c.InstanceType != nil {
		m["instance_type"] = aws.StringValue(c.InstanceType)
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

func flattenColdStorageOptions(coldStorageOptions *elasticsearch.ColdStorageOptions) []interface{} {
	if coldStorageOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled": aws.BoolValue(coldStorageOptions.Enabled),
	}

	return []interface{}{m}
}

func flattenZoneAwarenessConfig(zoneAwarenessConfig *elasticsearch.ZoneAwarenessConfig) []interface{} {
	if zoneAwarenessConfig == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"availability_zone_count": aws.Int64Value(zoneAwarenessConfig.AvailabilityZoneCount),
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
