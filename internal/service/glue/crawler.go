package glue

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceCrawler() *schema.Resource {
	return &schema.Resource{
		Create: resourceCrawlerCreate,
		Read:   resourceCrawlerRead,
		Update: resourceCrawlerUpdate,
		Delete: resourceCrawlerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9-_$#\/]+$`), ""),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				// Glue API always returns name
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					newARN, err := arn.Parse(new)

					if err != nil {
						return false
					}

					return old == strings.TrimPrefix(newARN.Resource, "role/")
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"schedule": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"classifiers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schema_change_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      glue.DeleteBehaviorDeprecateInDatabase,
							ValidateFunc: validation.StringInSlice(glue.DeleteBehavior_Values(), false),
						},
						"update_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      glue.UpdateBehaviorUpdateInDatabase,
							ValidateFunc: validation.StringInSlice(glue.UpdateBehavior_Values(), false),
						},
					},
				},
			},
			"table_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"s3_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"dlq_event_queue_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"event_queue_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"sample_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 249),
						},
					},
				},
			},
			"dynamodb_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"scan_all": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"scan_rate": {
							Type:         schema.TypeFloat,
							Optional:     true,
							ValidateFunc: validation.FloatBetween(0.1, 1.5),
						},
					},
				},
			},
			"mongodb_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"scan_all": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"jdbc_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"delta_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"delta_tables": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"write_manifest": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"catalog_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tables": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"configuration": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ValidateFunc: validation.StringIsJSON,
			},
			"lineage_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"crawler_lineage_settings": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      glue.CrawlerLineageSettingsDisable,
							ValidateFunc: validation.StringInSlice(glue.CrawlerLineageSettings_Values(), false),
						},
					},
				},
			},
			"recrawl_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recrawl_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      glue.RecrawlBehaviorCrawlEverything,
							ValidateFunc: validation.StringInSlice(glue.RecrawlBehavior_Values(), false),
						},
					},
				},
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCrawlerCreate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*conns.AWSClient).GlueConn
	name := d.Get("name").(string)

	crawlerInput, err := createCrawlerInput(d, name, meta.(*conns.AWSClient).DefaultTagsConfig)
	if err != nil {
		return err
	}

	// Retry for IAM eventual consistency
	err = resource.Retry(propagationTimeout, func() *resource.RetryError {
		_, err = glueConn.CreateCrawler(crawlerInput)
		if err != nil {
			// InvalidInputException: Insufficient Lake Formation permission(s) on xxx
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "Insufficient Lake Formation permission") {
				return resource.RetryableError(err)
			}

			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "Service is unable to assume role") {
				return resource.RetryableError(err)
			}

			// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "is not authorized") {
				return resource.RetryableError(err)
			}

			// InvalidInputException: SQS queue arn:aws:sqs:us-west-2:*******:tf-acc-test-4317277351691904203 does not exist or the role provided does not have access to it.
			if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "SQS queue") && tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "does not exist or the role provided does not have access to it") {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = glueConn.CreateCrawler(crawlerInput)
	}
	if err != nil {
		return fmt.Errorf("error creating Glue crawler: %w", err)
	}
	d.SetId(name)

	return resourceCrawlerRead(d, meta)
}

func createCrawlerInput(d *schema.ResourceData, crawlerName string, defaultTagsConfig *tftags.DefaultConfig) (*glue.CreateCrawlerInput, error) {
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	crawlerInput := &glue.CreateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get("database_name").(string)),
		Role:         aws.String(d.Get("role").(string)),
		Tags:         Tags(tags.IgnoreAWS()),
		Targets:      expandCrawlerTargets(d),
	}
	if description, ok := d.GetOk("description"); ok {
		crawlerInput.Description = aws.String(description.(string))
	}
	if schedule, ok := d.GetOk("schedule"); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	}
	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = flex.ExpandStringList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	if tablePrefix, ok := d.GetOk("table_prefix"); ok {
		crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	}
	if configuration, ok := d.GetOk("configuration"); ok {
		crawlerInput.Configuration = aws.String(configuration.(string))
	}

	if v, ok := d.GetOk("configuration"); ok {
		configuration, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("Configuration contains an invalid JSON: %v", err)
		}
		crawlerInput.Configuration = aws.String(configuration)
	}

	if securityConfiguration, ok := d.GetOk("security_configuration"); ok {
		crawlerInput.CrawlerSecurityConfiguration = aws.String(securityConfiguration.(string))
	}

	if v, ok := d.GetOk("lineage_configuration"); ok {
		crawlerInput.LineageConfiguration = expandCrawlerLineageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("recrawl_policy"); ok {
		crawlerInput.RecrawlPolicy = expandCrawlerRecrawlPolicy(v.([]interface{}))
	}

	return crawlerInput, nil
}

func updateCrawlerInput(d *schema.ResourceData, crawlerName string) (*glue.UpdateCrawlerInput, error) {
	crawlerInput := &glue.UpdateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get("database_name").(string)),
		Role:         aws.String(d.Get("role").(string)),
		Targets:      expandCrawlerTargets(d),
	}
	if description, ok := d.GetOk("description"); ok {
		crawlerInput.Description = aws.String(description.(string))
	}

	if schedule, ok := d.GetOk("schedule"); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	} else {
		crawlerInput.Schedule = aws.String("")
	}

	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = flex.ExpandStringList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	crawlerInput.TablePrefix = aws.String(d.Get("table_prefix").(string))

	if v, ok := d.GetOk("configuration"); ok {
		configuration, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("Configuration contains an invalid JSON: %v", err)
		}
		crawlerInput.Configuration = aws.String(configuration)
	} else {
		crawlerInput.Configuration = aws.String("")
	}

	if securityConfiguration, ok := d.GetOk("security_configuration"); ok {
		crawlerInput.CrawlerSecurityConfiguration = aws.String(securityConfiguration.(string))
	}

	if v, ok := d.GetOk("lineage_configuration"); ok {
		crawlerInput.LineageConfiguration = expandCrawlerLineageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("recrawl_policy"); ok {
		crawlerInput.RecrawlPolicy = expandCrawlerRecrawlPolicy(v.([]interface{}))
	}

	return crawlerInput, nil
}

func expandSchemaChangePolicy(v []interface{}) *glue.SchemaChangePolicy {
	if len(v) == 0 {
		return nil
	}

	schemaPolicy := &glue.SchemaChangePolicy{}

	member := v[0].(map[string]interface{})

	if updateBehavior, ok := member["update_behavior"]; ok && updateBehavior.(string) != "" {
		schemaPolicy.UpdateBehavior = aws.String(updateBehavior.(string))
	}

	if deleteBehavior, ok := member["delete_behavior"]; ok && deleteBehavior.(string) != "" {
		schemaPolicy.DeleteBehavior = aws.String(deleteBehavior.(string))
	}
	return schemaPolicy
}

func expandCrawlerTargets(d *schema.ResourceData) *glue.CrawlerTargets {
	crawlerTargets := &glue.CrawlerTargets{}

	log.Print("[DEBUG] Creating crawler target")

	if v, ok := d.GetOk("dynamodb_target"); ok {
		crawlerTargets.DynamoDBTargets = expandDynamoDBTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("jdbc_target"); ok {
		crawlerTargets.JdbcTargets = expandJDBCTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("s3_target"); ok {
		crawlerTargets.S3Targets = expandS3Targets(v.([]interface{}))
	}

	if v, ok := d.GetOk("catalog_target"); ok {
		crawlerTargets.CatalogTargets = expandCatalogTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("mongodb_target"); ok {
		crawlerTargets.MongoDBTargets = expandMongoDBTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("delta_target"); ok {
		crawlerTargets.DeltaTargets = expandDeltaTargets(v.([]interface{}))
	}

	return crawlerTargets
}

func expandDynamoDBTargets(targets []interface{}) []*glue.DynamoDBTarget {
	if len(targets) < 1 {
		return []*glue.DynamoDBTarget{}
	}

	perms := make([]*glue.DynamoDBTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandDynamoDBTarget(cfg)
	}
	return perms
}

func expandDynamoDBTarget(cfg map[string]interface{}) *glue.DynamoDBTarget {
	target := &glue.DynamoDBTarget{
		Path:    aws.String(cfg["path"].(string)),
		ScanAll: aws.Bool(cfg["scan_all"].(bool)),
	}

	if v, ok := cfg["scan_rate"].(float64); ok && v != 0 {
		target.ScanRate = aws.Float64(v)
	}

	return target
}

func expandS3Targets(targets []interface{}) []*glue.S3Target {
	if len(targets) < 1 {
		return []*glue.S3Target{}
	}

	perms := make([]*glue.S3Target, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandS3Target(cfg)
	}
	return perms
}

func expandS3Target(cfg map[string]interface{}) *glue.S3Target {
	target := &glue.S3Target{
		Path: aws.String(cfg["path"].(string)),
	}

	if v, ok := cfg["connection_name"]; ok {
		target.ConnectionName = aws.String(v.(string))
	}

	if v, ok := cfg["exclusions"]; ok {
		target.Exclusions = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := cfg["sample_size"]; ok && v.(int) > 0 {
		target.SampleSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := cfg["event_queue_arn"]; ok {
		target.EventQueueArn = aws.String(v.(string))
	}

	if v, ok := cfg["dlq_event_queue_arn"]; ok {
		target.DlqEventQueueArn = aws.String(v.(string))
	}

	return target
}

func expandJDBCTargets(targets []interface{}) []*glue.JdbcTarget {
	if len(targets) < 1 {
		return []*glue.JdbcTarget{}
	}

	perms := make([]*glue.JdbcTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandJDBCTarget(cfg)
	}
	return perms
}

func expandJDBCTarget(cfg map[string]interface{}) *glue.JdbcTarget {
	target := &glue.JdbcTarget{
		Path:           aws.String(cfg["path"].(string)),
		ConnectionName: aws.String(cfg["connection_name"].(string)),
	}

	if exclusions, ok := cfg["exclusions"]; ok {
		target.Exclusions = flex.ExpandStringList(exclusions.([]interface{}))
	}
	return target
}

func expandCatalogTargets(targets []interface{}) []*glue.CatalogTarget {
	if len(targets) < 1 {
		return []*glue.CatalogTarget{}
	}

	perms := make([]*glue.CatalogTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandCatalogTarget(cfg)
	}
	return perms
}

func expandCatalogTarget(cfg map[string]interface{}) *glue.CatalogTarget {
	target := &glue.CatalogTarget{
		DatabaseName: aws.String(cfg["database_name"].(string)),
		Tables:       flex.ExpandStringList(cfg["tables"].([]interface{})),
	}

	return target
}

func expandMongoDBTargets(targets []interface{}) []*glue.MongoDBTarget {
	if len(targets) < 1 {
		return []*glue.MongoDBTarget{}
	}

	perms := make([]*glue.MongoDBTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandMongoDBTarget(cfg)
	}
	return perms
}

func expandMongoDBTarget(cfg map[string]interface{}) *glue.MongoDBTarget {
	target := &glue.MongoDBTarget{
		ConnectionName: aws.String(cfg["connection_name"].(string)),
		Path:           aws.String(cfg["path"].(string)),
		ScanAll:        aws.Bool(cfg["scan_all"].(bool)),
	}

	return target
}

func expandDeltaTargets(targets []interface{}) []*glue.DeltaTarget {
	if len(targets) < 1 {
		return []*glue.DeltaTarget{}
	}

	perms := make([]*glue.DeltaTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandDeltaTarget(cfg)
	}
	return perms
}

func expandDeltaTarget(cfg map[string]interface{}) *glue.DeltaTarget {
	target := &glue.DeltaTarget{
		ConnectionName: aws.String(cfg["connection_name"].(string)),
		DeltaTables:    flex.ExpandStringSet(cfg["delta_tables"].(*schema.Set)),
		WriteManifest:  aws.Bool(cfg["write_manifest"].(bool)),
	}

	return target
}

func resourceCrawlerUpdate(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*conns.AWSClient).GlueConn
	name := d.Get("name").(string)

	if d.HasChangesExcept("tags", "tags_all") {
		updateCrawlerInput, err := updateCrawlerInput(d, name)
		if err != nil {
			return err
		}

		// Retry for IAM eventual consistency
		err = resource.Retry(propagationTimeout, func() *resource.RetryError {
			_, err := glueConn.UpdateCrawler(updateCrawlerInput)
			if err != nil {
				// InvalidInputException: Insufficient Lake Formation permission(s) on xxx
				if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "Insufficient Lake Formation permission") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "Service is unable to assume role") {
					return resource.RetryableError(err)
				}

				// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
				if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "is not authorized") {
					return resource.RetryableError(err)
				}

				// InvalidInputException: SQS queue arn:aws:sqs:us-west-2:*******:tf-acc-test-4317277351691904203 does not exist or the role provided does not have access to it.
				if tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "SQS queue") && tfawserr.ErrMessageContains(err, glue.ErrCodeInvalidInputException, "does not exist or the role provided does not have access to it") {
					return resource.RetryableError(err)
				}

				return resource.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = glueConn.UpdateCrawler(updateCrawlerInput)
		}

		if err != nil {
			return fmt.Errorf("error updating Glue crawler: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(glueConn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return resourceCrawlerRead(d, meta)
}

func resourceCrawlerRead(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glue.GetCrawlerInput{
		Name: aws.String(d.Id()),
	}

	crawlerOutput, err := glueConn.GetCrawler(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error reading Glue crawler: %w", err)
	}

	crawler := crawlerOutput.Crawler
	if crawler == nil {
		log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	crawlerARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("crawler/%s", d.Id()),
	}.String()
	d.Set("arn", crawlerARN)
	d.Set("name", crawler.Name)
	d.Set("database_name", crawler.DatabaseName)
	d.Set("role", crawler.Role)
	d.Set("configuration", crawler.Configuration)
	d.Set("description", crawler.Description)
	d.Set("security_configuration", crawler.CrawlerSecurityConfiguration)
	d.Set("schedule", "")
	if crawler.Schedule != nil {
		d.Set("schedule", crawler.Schedule.ScheduleExpression)
	}
	if err := d.Set("classifiers", flex.FlattenStringList(crawler.Classifiers)); err != nil {
		return fmt.Errorf("error setting classifiers: %w", err)
	}
	d.Set("table_prefix", crawler.TablePrefix)

	if crawler.SchemaChangePolicy != nil {
		if err := d.Set("schema_change_policy", flattenCrawlerSchemaChangePolicy(crawler.SchemaChangePolicy)); err != nil {
			return fmt.Errorf("error setting schema_change_policy: %w", err)
		}
	}

	if crawler.Targets != nil {
		if err := d.Set("dynamodb_target", flattenDynamoDBTargets(crawler.Targets.DynamoDBTargets)); err != nil {
			return fmt.Errorf("error setting dynamodb_target: %w", err)
		}

		if err := d.Set("jdbc_target", flattenJDBCTargets(crawler.Targets.JdbcTargets)); err != nil {
			return fmt.Errorf("error setting jdbc_target: %w", err)
		}

		if err := d.Set("s3_target", flattenS3Targets(crawler.Targets.S3Targets)); err != nil {
			return fmt.Errorf("error setting s3_target: %w", err)
		}

		if err := d.Set("catalog_target", flattenCatalogTargets(crawler.Targets.CatalogTargets)); err != nil {
			return fmt.Errorf("error setting catalog_target: %w", err)
		}

		if err := d.Set("mongodb_target", flattenMongoDBTargets(crawler.Targets.MongoDBTargets)); err != nil {
			return fmt.Errorf("error setting mongodb_target: %w", err)
		}

		if err := d.Set("delta_target", flattenDeltaTargets(crawler.Targets.DeltaTargets)); err != nil {
			return fmt.Errorf("error setting delta_target: %w", err)
		}
	}

	tags, err := ListTags(glueConn, crawlerARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Crawler (%s): %w", crawlerARN, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	if err := d.Set("lineage_configuration", flattenCrawlerLineageConfiguration(crawler.LineageConfiguration)); err != nil {
		return fmt.Errorf("error setting lineage_configuration: %w", err)
	}

	if err := d.Set("recrawl_policy", flattenCrawlerRecrawlPolicy(crawler.RecrawlPolicy)); err != nil {
		return fmt.Errorf("error setting recrawl_policy: %w", err)
	}

	return nil
}

func flattenS3Targets(s3Targets []*glue.S3Target) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, s3Target := range s3Targets {
		attrs := make(map[string]interface{})
		attrs["exclusions"] = flex.FlattenStringList(s3Target.Exclusions)
		attrs["path"] = aws.StringValue(s3Target.Path)
		attrs["connection_name"] = aws.StringValue(s3Target.ConnectionName)

		if s3Target.SampleSize != nil {
			attrs["sample_size"] = aws.Int64Value(s3Target.SampleSize)
		}

		attrs["event_queue_arn"] = aws.StringValue(s3Target.EventQueueArn)
		attrs["dlq_event_queue_arn"] = aws.StringValue(s3Target.DlqEventQueueArn)

		result = append(result, attrs)
	}
	return result
}

func flattenCatalogTargets(CatalogTargets []*glue.CatalogTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, catalogTarget := range CatalogTargets {
		attrs := make(map[string]interface{})
		attrs["tables"] = flex.FlattenStringList(catalogTarget.Tables)
		attrs["database_name"] = aws.StringValue(catalogTarget.DatabaseName)

		result = append(result, attrs)
	}
	return result
}

func flattenDynamoDBTargets(dynamodbTargets []*glue.DynamoDBTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, dynamodbTarget := range dynamodbTargets {
		attrs := make(map[string]interface{})
		attrs["path"] = aws.StringValue(dynamodbTarget.Path)
		attrs["scan_all"] = aws.BoolValue(dynamodbTarget.ScanAll)
		attrs["scan_rate"] = aws.Float64Value(dynamodbTarget.ScanRate)

		result = append(result, attrs)
	}
	return result
}

func flattenJDBCTargets(jdbcTargets []*glue.JdbcTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, jdbcTarget := range jdbcTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.StringValue(jdbcTarget.ConnectionName)
		attrs["exclusions"] = flex.FlattenStringList(jdbcTarget.Exclusions)
		attrs["path"] = aws.StringValue(jdbcTarget.Path)

		result = append(result, attrs)
	}
	return result
}

func flattenMongoDBTargets(mongoDBTargets []*glue.MongoDBTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, mongoDBTarget := range mongoDBTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.StringValue(mongoDBTarget.ConnectionName)
		attrs["path"] = aws.StringValue(mongoDBTarget.Path)
		attrs["scan_all"] = aws.BoolValue(mongoDBTarget.ScanAll)

		result = append(result, attrs)
	}
	return result
}

func flattenDeltaTargets(deltaTargets []*glue.DeltaTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, deltaTarget := range deltaTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.StringValue(deltaTarget.ConnectionName)
		attrs["delta_tables"] = flex.FlattenStringSet(deltaTarget.DeltaTables)
		attrs["write_manifest"] = aws.BoolValue(deltaTarget.WriteManifest)

		result = append(result, attrs)
	}
	return result
}

func resourceCrawlerDelete(d *schema.ResourceData, meta interface{}) error {
	glueConn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] deleting Glue crawler: %s", d.Id())
	_, err := glueConn.DeleteCrawler(&glue.DeleteCrawlerInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Glue crawler: %w", err)
	}
	return nil
}

func flattenCrawlerSchemaChangePolicy(cfg *glue.SchemaChangePolicy) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"delete_behavior": aws.StringValue(cfg.DeleteBehavior),
		"update_behavior": aws.StringValue(cfg.UpdateBehavior),
	}

	return []map[string]interface{}{m}
}

func expandCrawlerLineageConfiguration(cfg []interface{}) *glue.LineageConfiguration {
	m := cfg[0].(map[string]interface{})

	target := &glue.LineageConfiguration{
		CrawlerLineageSettings: aws.String(m["crawler_lineage_settings"].(string)),
	}
	return target
}

func flattenCrawlerLineageConfiguration(cfg *glue.LineageConfiguration) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"crawler_lineage_settings": aws.StringValue(cfg.CrawlerLineageSettings),
	}

	return []map[string]interface{}{m}
}

func expandCrawlerRecrawlPolicy(cfg []interface{}) *glue.RecrawlPolicy {
	m := cfg[0].(map[string]interface{})

	target := &glue.RecrawlPolicy{
		RecrawlBehavior: aws.String(m["recrawl_behavior"].(string)),
	}
	return target
}

func flattenCrawlerRecrawlPolicy(cfg *glue.RecrawlPolicy) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"recrawl_behavior": aws.StringValue(cfg.RecrawlBehavior),
	}

	return []map[string]interface{}{m}
}
