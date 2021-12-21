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
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
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
			Update: schema.DefaultTimeout(60 * time.Minute),
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
									"start_at": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},
									"duration": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"value": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"unit": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(elasticsearch.TimeUnit_Values(), false),
												},
											},
										},
									},
									"cron_expression_for_recurrence": {
										Type:     schema.TypeString,
										Required: true,
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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-z][0-9a-z\-]{2,27}$`),
					"must start with a lowercase alphabet and be at least 3 and no more than 28 characters long."+
						" Valid characters are a-z (lowercase letters), 0-9, and - (hyphen)."),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
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
						"custom_endpoint_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validation.StringInSlice(elasticsearch.VolumeType_Values(), false),
							),
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
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
							ForceNew: true,
						},
						"kms_key_id": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							DiffSuppressFunc: suppressEquivalentKmsKeyIds,
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
							ForceNew: true,
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
						"warm_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"warm_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(2, 150),
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
			"log_publishing_options": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(elasticsearch.LogType_Values(), false),
						},
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
					},
				},
			},
			"elasticsearch_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.5",
			},
			"cognito_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         false,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"user_pool_id": {
							Type:     schema.TypeString,
							Required: true,
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
					},
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDomainImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	d.Set("domain_name", d.Id())
	return []*schema.ResourceData{d}, nil
}

func resourceDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// The API doesn't check for duplicate names
	// so w/out this check Create would act as upsert
	// and might cause duplicate domain to appear in state
	resp, err := FindDomainByName(conn, d.Get("domain_name").(string))
	if err == nil {
		return fmt.Errorf("Elasticsearch domain %s already exists", aws.StringValue(resp.DomainName))
	}

	inputCreateDomain := elasticsearch.CreateElasticsearchDomainInput{
		DomainName:           aws.String(d.Get("domain_name").(string)),
		ElasticsearchVersion: aws.String(d.Get("elasticsearch_version").(string)),
		TagList:              Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("access_policies"); ok {
		policy, err := structure.NormalizeJsonString(v.(string))

		if err != nil {
			return fmt.Errorf("policy (%s) is invalid JSON: %w", policy, err)
		}

		inputCreateDomain.AccessPolicies = aws.String(policy)
	}

	if v, ok := d.GetOk("advanced_options"); ok {
		inputCreateDomain.AdvancedOptions = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("advanced_security_options"); ok {
		inputCreateDomain.AdvancedSecurityOptions = expandAdvancedSecurityOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]interface{})) > 0 {
		inputCreateDomain.AutoTuneOptions = expandAutoTuneOptionsInput(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ebs_options"); ok {
		options := v.([]interface{})

		if len(options) == 1 {
			if options[0] == nil {
				return fmt.Errorf("At least one field is expected inside ebs_options")
			}

			s := options[0].(map[string]interface{})
			inputCreateDomain.EBSOptions = expandEBSOptions(s)
		}
	}

	if v, ok := d.GetOk("encrypt_at_rest"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside encrypt_at_rest")
		}

		s := options[0].(map[string]interface{})
		inputCreateDomain.EncryptionAtRestOptions = expandEncryptAtRestOptions(s)
	}

	if v, ok := d.GetOk("cluster_config"); ok {
		config := v.([]interface{})

		if len(config) == 1 {
			if config[0] == nil {
				return fmt.Errorf("At least one field is expected inside cluster_config")
			}
			m := config[0].(map[string]interface{})
			inputCreateDomain.ElasticsearchClusterConfig = expandClusterConfig(m)
		}
	}

	if v, ok := d.GetOk("node_to_node_encryption"); ok {
		options := v.([]interface{})

		s := options[0].(map[string]interface{})
		inputCreateDomain.NodeToNodeEncryptionOptions = expandNodeToNodeEncryptionOptions(s)
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

			inputCreateDomain.SnapshotOptions = &snapshotOptions
		}
	}

	if v, ok := d.GetOk("vpc_options"); ok {
		options := v.([]interface{})
		if options[0] == nil {
			return fmt.Errorf("At least one field is expected inside vpc_options")
		}

		s := options[0].(map[string]interface{})
		inputCreateDomain.VPCOptions = expandVPCOptions(s)
	}

	if v, ok := d.GetOk("log_publishing_options"); ok {
		inputCreateDomain.LogPublishingOptions = expandLogPublishingOptions(v.(*schema.Set))
	}

	if v, ok := d.GetOk("domain_endpoint_options"); ok {
		inputCreateDomain.DomainEndpointOptions = expandDomainEndpointOptions(v.([]interface{}))
	}

	if v, ok := d.GetOk("cognito_options"); ok {
		inputCreateDomain.CognitoOptions = expandCognitoOptions(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating Elasticsearch domain: %s", inputCreateDomain)

	// IAM Roles can take some time to propagate if set in AccessPolicies and created in the same terraform
	var out *elasticsearch.CreateElasticsearchDomainOutput
	err = resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error
		out, err = conn.CreateElasticsearchDomain(&inputCreateDomain)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidTypeException", "Error setting policy") {
				log.Printf("[DEBUG] Retrying creation of Elasticsearch domain %s", aws.StringValue(inputCreateDomain.DomainName))
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "enable a service-linked role to give Amazon ES permissions") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Domain is still being deleted") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Amazon Elasticsearch must be allowed to use the passed role") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "The passed role has not propagated yet") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Authentication error") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "Unauthorized Operation: Elasticsearch must be authorised to describe") {
				return resource.RetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "ValidationException", "The passed role must authorize Amazon Elasticsearch to describe") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		out, err = conn.CreateElasticsearchDomain(&inputCreateDomain)
	}
	if err != nil {
		return fmt.Errorf("Error creating Elasticsearch domain: %w", err)
	}

	d.SetId(aws.StringValue(out.DomainStatus.ARN))

	log.Printf("[DEBUG] Waiting for Elasticsearch domain %q to be created", d.Id())
	if err := WaitForDomainCreation(conn, d.Get("domain_name").(string)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain (%s) to be created: %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Elasticsearch domain %q created", d.Id())

	if v, ok := d.GetOk("auto_tune_options"); ok && len(v.([]interface{})) > 0 {

		log.Printf("[DEBUG] Modifying config for Elasticsearch domain %q", d.Id())

		inputUpdateDomainConfig := &elasticsearch.UpdateElasticsearchDomainConfigInput{
			DomainName: aws.String(d.Get("domain_name").(string)),
		}

		inputUpdateDomainConfig.AutoTuneOptions = expandAutoTuneOptions(v.([]interface{})[0].(map[string]interface{}))

		_, err = conn.UpdateElasticsearchDomainConfig(inputUpdateDomainConfig)

		if err != nil {
			return fmt.Errorf("Error modifying config for Elasticsearch domain: %s", err)
		}

		log.Printf("[DEBUG] Config for Elasticsearch domain %q modified", d.Id())
	}

	return resourceDomainRead(d, meta)
}

func resourceDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ElasticsearchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ds, err := FindDomainByName(conn, d.Get("domain_name").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Elasticsearch domain (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Elasticsearch domain (%s): %w", d.Id(), err)
	}

	log.Printf("[DEBUG] Received Elasticsearch domain: %s", ds)

	outDescribeDomainConfig, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received config for Elasticsearch domain: %s", outDescribeDomainConfig)

	dc := outDescribeDomainConfig.DomainConfig

	if ds.AccessPolicies != nil && aws.StringValue(ds.AccessPolicies) != "" {
		policies, err := verify.PolicyToSet(d.Get("access_policies").(string), aws.StringValue(ds.AccessPolicies))

		if err != nil {
			return err
		}

		d.Set("access_policies", policies)
	}

	options := advancedOptionsIgnoreDefault(d.Get("advanced_options").(map[string]interface{}), flex.PointersMapToStringList(ds.AdvancedOptions))
	if err = d.Set("advanced_options", options); err != nil {
		return fmt.Errorf("setting advanced_options %v: %w", options, err)
	}

	d.SetId(aws.StringValue(ds.ARN))
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
		err = d.Set("auto_tune_options", []interface{}{flattenAutoTuneOptions(v.Options)})
		if err != nil {
			return err
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
		err = d.Set("endpoint", endpoints["vpc"])
		if err != nil {
			return err
		}
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
		return fmt.Errorf("error listing tags for Elasticsearch Cluster (%s): %w", d.Id(), err)
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
		input := elasticsearch.UpdateElasticsearchDomainConfigInput{
			DomainName: aws.String(d.Get("domain_name").(string)),
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
				}
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

		_, err := conn.UpdateElasticsearchDomainConfig(&input)
		if err != nil {
			return err
		}

		if err := waitForDomainUpdate(conn, d.Get("domain_name").(string)); err != nil {
			return fmt.Errorf("error waiting for Elasticsearch Domain Update (%s) to succeed: %w", d.Id(), err)
		}

		if d.HasChange("elasticsearch_version") {
			upgradeInput := elasticsearch.UpgradeElasticsearchDomainInput{
				DomainName:    aws.String(d.Get("domain_name").(string)),
				TargetVersion: aws.String(d.Get("elasticsearch_version").(string)),
			}

			_, err := conn.UpgradeElasticsearchDomain(&upgradeInput)
			if err != nil {
				return fmt.Errorf("Failed to upgrade elasticsearch domain: %w", err)
			}

			if _, err := waitUpgradeSucceeded(conn, d.Get("domain_name").(string), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return fmt.Errorf("error waiting for Elasticsearch Domain Upgrade (%s) to succeed: %w", d.Id(), err)
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
	domainName := d.Get("domain_name").(string)

	log.Printf("[DEBUG] Deleting Elasticsearch domain: %q", domainName)
	_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
		DomainName: aws.String(domainName),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, elasticsearch.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Waiting for Elasticsearch domain %q to be deleted", domainName)
	if err := waitForDomainDelete(conn, d.Get("domain_name").(string)); err != nil {
		return fmt.Errorf("error waiting for Elasticsearch Domain (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}

func suppressEquivalentKmsKeyIds(k, old, new string, d *schema.ResourceData) bool {
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
