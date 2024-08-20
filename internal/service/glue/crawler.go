// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func targets() []string {
	return []string{"s3_target", "dynamodb_target", "mongodb_target", "jdbc_target", "catalog_target", "delta_target", "iceberg_target", "hudi_target"}
}

// @SDKResource("aws_glue_crawler", name="Crawler")
// @Tags(identifierAttribute="arn")
func ResourceCrawler() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCrawlerCreate,
		ReadWithoutTimeout:   resourceCrawlerRead,
		UpdateWithoutTimeout: resourceCrawlerUpdate,
		DeleteWithoutTimeout: resourceCrawlerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"catalog_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrDatabaseName: {
							Type:     schema.TypeString,
							Required: true,
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
						"tables": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"classifiers": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrConfiguration: {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				ValidateFunc: validation.StringIsJSON,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"delta_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"create_native_delta_table": {
							Type:     schema.TypeBool,
							Optional: true,
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
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 2048),
			},
			"dynamodb_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrPath: {
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
			"hudi_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"maximum_traversal_depth": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"paths": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"iceberg_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"maximum_traversal_depth": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"paths": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"jdbc_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"enable_additional_metadata": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[awstypes.JdbcMetadataEntry](),
							},
						},
						"exclusions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"lake_formation_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAccountID: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidAccountID,
						},
						"use_lake_formation_credentials": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"lineage_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"crawler_lineage_settings": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.CrawlerLineageSettingsDisable,
							ValidateDiagFunc: enum.Validate[awstypes.CrawlerLineageSettings](),
						},
					},
				},
			},
			"mongodb_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrPath: {
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
			names.AttrName: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z_$#\/-]+$`), ""),
				),
			},
			"recrawl_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recrawl_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.RecrawlBehaviorCrawlEverything,
							ValidateDiagFunc: enum.Validate[awstypes.RecrawlBehavior](),
						},
					},
				},
			},
			names.AttrRole: {
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
			"s3_target": {
				Type:         schema.TypeList,
				Optional:     true,
				MinItems:     1,
				AtLeastOneOf: targets(),
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
						names.AttrPath: {
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
			names.AttrSchedule: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"schema_change_policy": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.DeleteBehaviorDeprecateInDatabase,
							ValidateDiagFunc: enum.Validate[awstypes.DeleteBehavior](),
						},
						"update_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.UpdateBehaviorUpdateInDatabase,
							ValidateDiagFunc: enum.Validate[awstypes.UpdateBehavior](),
						},
					},
				},
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"table_prefix": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCrawlerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	glueConn := meta.(*conns.AWSClient).GlueClient(ctx)
	name := d.Get(names.AttrName).(string)

	crawlerInput, err := createCrawlerInput(ctx, d, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Crawler (%s): %s", name, err)
	}

	// Retry for IAM eventual consistency
	err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		_, err = glueConn.CreateCrawler(ctx, crawlerInput)
		if err != nil {
			// InvalidInputException: Insufficient Lake Formation permission(s) on xxx
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Insufficient Lake Formation permission") {
				return retry.RetryableError(err)
			}

			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Service is unable to assume provided role") {
				return retry.RetryableError(err)
			}

			// InvalidInputException: com.amazonaws.services.glue.model.AccessDeniedException: You need to enable AWS Security Token Service for this region. . Please verify the role's TrustPolicy.
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Please verify the role's TrustPolicy") {
				return retry.RetryableError(err)
			}

			// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "is not authorized") {
				return retry.RetryableError(err)
			}

			// InvalidInputException: SQS queue arn:aws:sqs:us-west-2:*******:tf-acc-test-4317277351691904203 does not exist or the role provided does not have access to it.
			if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "SQS queue") && errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "does not exist or the role provided does not have access to it") {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = glueConn.CreateCrawler(ctx, crawlerInput)
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Crawler (%s): %s", name, err)
	}
	d.SetId(name)

	return append(diags, resourceCrawlerRead(ctx, d, meta)...)
}

func resourceCrawlerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	crawler, err := FindCrawlerByName(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Crawler (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Crawler (%s): %s", d.Id(), err)
	}

	crawlerARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("crawler/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, crawlerARN)
	d.Set(names.AttrName, crawler.Name)
	d.Set(names.AttrDatabaseName, crawler.DatabaseName)
	d.Set(names.AttrRole, crawler.Role)
	d.Set(names.AttrConfiguration, crawler.Configuration)
	d.Set(names.AttrDescription, crawler.Description)
	d.Set("security_configuration", crawler.CrawlerSecurityConfiguration)
	d.Set(names.AttrSchedule, "")
	if crawler.Schedule != nil {
		d.Set(names.AttrSchedule, crawler.Schedule.ScheduleExpression)
	}
	if err := d.Set("classifiers", flex.FlattenStringValueList(crawler.Classifiers)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting classifiers: %s", err)
	}
	d.Set("table_prefix", crawler.TablePrefix)

	if crawler.SchemaChangePolicy != nil {
		if err := d.Set("schema_change_policy", flattenCrawlerSchemaChangePolicy(crawler.SchemaChangePolicy)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting schema_change_policy: %s", err)
		}
	}

	if crawler.Targets != nil {
		if err := d.Set("dynamodb_target", flattenDynamoDBTargets(crawler.Targets.DynamoDBTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting dynamodb_target: %s", err)
		}

		if err := d.Set("jdbc_target", flattenJDBCTargets(crawler.Targets.JdbcTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting jdbc_target: %s", err)
		}

		if err := d.Set("s3_target", flattenS3Targets(crawler.Targets.S3Targets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting s3_target: %s", err)
		}

		if err := d.Set("catalog_target", flattenCatalogTargets(crawler.Targets.CatalogTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting catalog_target: %s", err)
		}

		if err := d.Set("mongodb_target", flattenMongoDBTargets(crawler.Targets.MongoDBTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting mongodb_target: %s", err)
		}

		if err := d.Set("delta_target", flattenDeltaTargets(crawler.Targets.DeltaTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting delta_target: %s", err)
		}

		if err := d.Set("hudi_target", flattenHudiTargets(crawler.Targets.HudiTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting hudi_target: %s", err)
		}

		if err := d.Set("iceberg_target", flattenIcebergTargets(crawler.Targets.IcebergTargets)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting iceberg_target: %s", err)
		}
	}

	if err := d.Set("lineage_configuration", flattenCrawlerLineageConfiguration(crawler.LineageConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lineage_configuration: %s", err)
	}

	if err := d.Set("lake_formation_configuration", flattenLakeFormationConfiguration(crawler.LakeFormationConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lake_formation_configuration: %s", err)
	}

	if err := d.Set("recrawl_policy", flattenCrawlerRecrawlPolicy(crawler.RecrawlPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting recrawl_policy: %s", err)
	}

	return diags
}

func resourceCrawlerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	glueConn := meta.(*conns.AWSClient).GlueClient(ctx)
	name := d.Get(names.AttrName).(string)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		updateCrawlerInput, err := updateCrawlerInput(d, name)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Crawler (%s): %s", d.Id(), err)
		}

		// Retry for IAM eventual consistency
		err = retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
			_, err := glueConn.UpdateCrawler(ctx, updateCrawlerInput)
			if err != nil {
				// InvalidInputException: Insufficient Lake Formation permission(s) on xxx
				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Insufficient Lake Formation permission") {
					return retry.RetryableError(err)
				}

				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Service is unable to assume provided role") {
					return retry.RetryableError(err)
				}

				// InvalidInputException: com.amazonaws.services.glue.model.AccessDeniedException: You need to enable AWS Security Token Service for this region. . Please verify the role's TrustPolicy.
				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "Please verify the role's TrustPolicy") {
					return retry.RetryableError(err)
				}

				// InvalidInputException: Unable to retrieve connection tf-acc-test-8656357591012534997: User: arn:aws:sts::*******:assumed-role/tf-acc-test-8656357591012534997/AWS-Crawler is not authorized to perform: glue:GetConnection on resource: * (Service: AmazonDataCatalog; Status Code: 400; Error Code: AccessDeniedException; Request ID: 4d72b66f-9c75-11e8-9faf-5b526c7be968)
				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "is not authorized") {
					return retry.RetryableError(err)
				}

				// InvalidInputException: SQS queue arn:aws:sqs:us-west-2:*******:tf-acc-test-4317277351691904203 does not exist or the role provided does not have access to it.
				if errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "SQS queue") && errs.IsAErrorMessageContains[*awstypes.InvalidInputException](err, "does not exist or the role provided does not have access to it") {
					return retry.RetryableError(err)
				}

				return retry.NonRetryableError(err)
			}
			return nil
		})

		if tfresource.TimedOut(err) {
			_, err = glueConn.UpdateCrawler(ctx, updateCrawlerInput)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Crawler (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCrawlerRead(ctx, d, meta)...)
}

func resourceCrawlerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	glueConn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Crawler: %s", d.Id())
	_, err := glueConn.DeleteCrawler(ctx, &glue.DeleteCrawlerInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Crawler (%s): %s", d.Id(), err)
	}

	return diags
}

func createCrawlerInput(ctx context.Context, d *schema.ResourceData, crawlerName string) (*glue.CreateCrawlerInput, error) {
	crawlerInput := &glue.CreateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
		Role:         aws.String(d.Get(names.AttrRole).(string)),
		Tags:         getTagsIn(ctx),
		Targets:      expandCrawlerTargets(d),
	}
	if description, ok := d.GetOk(names.AttrDescription); ok {
		crawlerInput.Description = aws.String(description.(string))
	}
	if schedule, ok := d.GetOk(names.AttrSchedule); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	}
	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = flex.ExpandStringValueList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	if tablePrefix, ok := d.GetOk("table_prefix"); ok {
		crawlerInput.TablePrefix = aws.String(tablePrefix.(string))
	}
	if configuration, ok := d.GetOk(names.AttrConfiguration); ok {
		crawlerInput.Configuration = aws.String(configuration.(string))
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok {
		configuration, err := structure.NormalizeJsonString(v)
		if err != nil {
			return nil, fmt.Errorf("configuration contains an invalid JSON: %v", err)
		}
		crawlerInput.Configuration = aws.String(configuration)
	}

	if securityConfiguration, ok := d.GetOk("security_configuration"); ok {
		crawlerInput.CrawlerSecurityConfiguration = aws.String(securityConfiguration.(string))
	}

	if v, ok := d.GetOk("lineage_configuration"); ok {
		crawlerInput.LineageConfiguration = expandCrawlerLineageConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("lake_formation_configuration"); ok {
		crawlerInput.LakeFormationConfiguration = expandLakeFormationConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("recrawl_policy"); ok {
		crawlerInput.RecrawlPolicy = expandCrawlerRecrawlPolicy(v.([]interface{}))
	}

	return crawlerInput, nil
}

func updateCrawlerInput(d *schema.ResourceData, crawlerName string) (*glue.UpdateCrawlerInput, error) {
	crawlerInput := &glue.UpdateCrawlerInput{
		Name:         aws.String(crawlerName),
		DatabaseName: aws.String(d.Get(names.AttrDatabaseName).(string)),
		Role:         aws.String(d.Get(names.AttrRole).(string)),
		Targets:      expandCrawlerTargets(d),
	}
	if description, ok := d.GetOk(names.AttrDescription); ok {
		crawlerInput.Description = aws.String(description.(string))
	}

	if schedule, ok := d.GetOk(names.AttrSchedule); ok {
		crawlerInput.Schedule = aws.String(schedule.(string))
	} else {
		crawlerInput.Schedule = aws.String("")
	}

	if classifiers, ok := d.GetOk("classifiers"); ok {
		crawlerInput.Classifiers = flex.ExpandStringValueList(classifiers.([]interface{}))
	}

	crawlerInput.SchemaChangePolicy = expandSchemaChangePolicy(d.Get("schema_change_policy").([]interface{}))

	crawlerInput.TablePrefix = aws.String(d.Get("table_prefix").(string))

	if v, ok := d.GetOk(names.AttrConfiguration); ok {
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

	if v, ok := d.GetOk("lake_formation_configuration"); ok {
		crawlerInput.LakeFormationConfiguration = expandLakeFormationConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("recrawl_policy"); ok {
		crawlerInput.RecrawlPolicy = expandCrawlerRecrawlPolicy(v.([]interface{}))
	}

	return crawlerInput, nil
}

func expandSchemaChangePolicy(v []interface{}) *awstypes.SchemaChangePolicy {
	if len(v) == 0 {
		return nil
	}

	schemaPolicy := &awstypes.SchemaChangePolicy{}

	member := v[0].(map[string]interface{})

	if updateBehavior, ok := member["update_behavior"]; ok && updateBehavior.(string) != "" {
		schemaPolicy.UpdateBehavior = awstypes.UpdateBehavior(updateBehavior.(string))
	}

	if deleteBehavior, ok := member["delete_behavior"]; ok && deleteBehavior.(string) != "" {
		schemaPolicy.DeleteBehavior = awstypes.DeleteBehavior(deleteBehavior.(string))
	}
	return schemaPolicy
}

func expandCrawlerTargets(d *schema.ResourceData) *awstypes.CrawlerTargets {
	crawlerTargets := &awstypes.CrawlerTargets{}

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

	if v, ok := d.GetOk("hudi_target"); ok {
		crawlerTargets.HudiTargets = expandHudiTargets(v.([]interface{}))
	}

	if v, ok := d.GetOk("iceberg_target"); ok {
		crawlerTargets.IcebergTargets = expandIcebergTargets(v.([]interface{}))
	}

	return crawlerTargets
}

func expandDynamoDBTargets(targets []interface{}) []awstypes.DynamoDBTarget {
	if len(targets) < 1 {
		return []awstypes.DynamoDBTarget{}
	}

	perms := make([]awstypes.DynamoDBTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandDynamoDBTarget(cfg)
	}
	return perms
}

func expandDynamoDBTarget(cfg map[string]interface{}) awstypes.DynamoDBTarget {
	target := awstypes.DynamoDBTarget{
		Path:    aws.String(cfg[names.AttrPath].(string)),
		ScanAll: aws.Bool(cfg["scan_all"].(bool)),
	}

	if v, ok := cfg["scan_rate"].(float64); ok && v != 0 {
		target.ScanRate = aws.Float64(v)
	}

	return target
}

func expandS3Targets(targets []interface{}) []awstypes.S3Target {
	if len(targets) < 1 {
		return []awstypes.S3Target{}
	}

	perms := make([]awstypes.S3Target, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandS3Target(cfg)
	}
	return perms
}

func expandS3Target(cfg map[string]interface{}) awstypes.S3Target {
	target := awstypes.S3Target{
		Path: aws.String(cfg[names.AttrPath].(string)),
	}

	if v, ok := cfg["connection_name"]; ok {
		target.ConnectionName = aws.String(v.(string))
	}

	if v, ok := cfg["exclusions"]; ok {
		target.Exclusions = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := cfg["sample_size"]; ok && v.(int) > 0 {
		target.SampleSize = aws.Int32(int32(v.(int)))
	}

	if v, ok := cfg["event_queue_arn"]; ok {
		target.EventQueueArn = aws.String(v.(string))
	}

	if v, ok := cfg["dlq_event_queue_arn"]; ok {
		target.DlqEventQueueArn = aws.String(v.(string))
	}

	return target
}

func expandJDBCTargets(targets []interface{}) []awstypes.JdbcTarget {
	if len(targets) < 1 {
		return []awstypes.JdbcTarget{}
	}

	perms := make([]awstypes.JdbcTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandJDBCTarget(cfg)
	}
	return perms
}

func expandJDBCTarget(cfg map[string]interface{}) awstypes.JdbcTarget {
	target := awstypes.JdbcTarget{
		Path:           aws.String(cfg[names.AttrPath].(string)),
		ConnectionName: aws.String(cfg["connection_name"].(string)),
	}

	if v, ok := cfg["enable_additional_metadata"].([]interface{}); ok {
		target.EnableAdditionalMetadata = flex.ExpandStringyValueList[awstypes.JdbcMetadataEntry](v)
	}

	if v, ok := cfg["exclusions"].([]interface{}); ok {
		target.Exclusions = flex.ExpandStringValueList(v)
	}

	return target
}

func expandCatalogTargets(targets []interface{}) []awstypes.CatalogTarget {
	if len(targets) < 1 {
		return []awstypes.CatalogTarget{}
	}

	perms := make([]awstypes.CatalogTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandCatalogTarget(cfg)
	}
	return perms
}

func expandCatalogTarget(cfg map[string]interface{}) awstypes.CatalogTarget {
	target := awstypes.CatalogTarget{
		DatabaseName: aws.String(cfg[names.AttrDatabaseName].(string)),
		Tables:       flex.ExpandStringValueList(cfg["tables"].([]interface{})),
	}

	if v, ok := cfg["connection_name"].(string); ok {
		target.ConnectionName = aws.String(v)
	}

	if v, ok := cfg["dlq_event_queue_arn"].(string); ok {
		target.DlqEventQueueArn = aws.String(v)
	}

	if v, ok := cfg["event_queue_arn"].(string); ok {
		target.EventQueueArn = aws.String(v)
	}

	return target
}

func expandMongoDBTargets(targets []interface{}) []awstypes.MongoDBTarget {
	if len(targets) < 1 {
		return []awstypes.MongoDBTarget{}
	}

	perms := make([]awstypes.MongoDBTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandMongoDBTarget(cfg)
	}
	return perms
}

func expandMongoDBTarget(cfg map[string]interface{}) awstypes.MongoDBTarget {
	target := awstypes.MongoDBTarget{
		ConnectionName: aws.String(cfg["connection_name"].(string)),
		Path:           aws.String(cfg[names.AttrPath].(string)),
		ScanAll:        aws.Bool(cfg["scan_all"].(bool)),
	}

	return target
}

func expandDeltaTargets(targets []interface{}) []awstypes.DeltaTarget {
	if len(targets) < 1 {
		return []awstypes.DeltaTarget{}
	}

	perms := make([]awstypes.DeltaTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandDeltaTarget(cfg)
	}
	return perms
}

func expandDeltaTarget(cfg map[string]interface{}) awstypes.DeltaTarget {
	target := awstypes.DeltaTarget{
		CreateNativeDeltaTable: aws.Bool(cfg["create_native_delta_table"].(bool)),
		DeltaTables:            flex.ExpandStringValueSet(cfg["delta_tables"].(*schema.Set)),
		WriteManifest:          aws.Bool(cfg["write_manifest"].(bool)),
	}

	if v, ok := cfg["connection_name"].(string); ok {
		target.ConnectionName = aws.String(v)
	}

	return target
}

func expandHudiTargets(targets []interface{}) []awstypes.HudiTarget {
	if len(targets) < 1 {
		return []awstypes.HudiTarget{}
	}

	perms := make([]awstypes.HudiTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandHudiTarget(cfg)
	}
	return perms
}

func expandHudiTarget(cfg map[string]interface{}) awstypes.HudiTarget {
	target := awstypes.HudiTarget{
		Paths:                 flex.ExpandStringValueSet(cfg["paths"].(*schema.Set)),
		MaximumTraversalDepth: aws.Int32(int32(cfg["maximum_traversal_depth"].(int))),
	}

	if v, ok := cfg["exclusions"]; ok {
		target.Exclusions = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := cfg["connection_name"].(string); ok {
		target.ConnectionName = aws.String(v)
	}

	return target
}

func expandIcebergTargets(targets []interface{}) []awstypes.IcebergTarget {
	if len(targets) < 1 {
		return []awstypes.IcebergTarget{}
	}

	perms := make([]awstypes.IcebergTarget, len(targets))
	for i, rawCfg := range targets {
		cfg := rawCfg.(map[string]interface{})
		perms[i] = expandIcebergTarget(cfg)
	}
	return perms
}

func expandIcebergTarget(cfg map[string]interface{}) awstypes.IcebergTarget {
	target := awstypes.IcebergTarget{
		Paths:                 flex.ExpandStringValueSet(cfg["paths"].(*schema.Set)),
		MaximumTraversalDepth: aws.Int32(int32(cfg["maximum_traversal_depth"].(int))),
	}

	if v, ok := cfg["exclusions"]; ok {
		target.Exclusions = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := cfg["connection_name"].(string); ok {
		target.ConnectionName = aws.String(v)
	}

	return target
}

func flattenS3Targets(s3Targets []awstypes.S3Target) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, s3Target := range s3Targets {
		attrs := make(map[string]interface{})
		attrs["exclusions"] = flex.FlattenStringValueList(s3Target.Exclusions)
		attrs[names.AttrPath] = aws.ToString(s3Target.Path)
		attrs["connection_name"] = aws.ToString(s3Target.ConnectionName)

		if s3Target.SampleSize != nil {
			attrs["sample_size"] = aws.ToInt32(s3Target.SampleSize)
		}

		attrs["event_queue_arn"] = aws.ToString(s3Target.EventQueueArn)
		attrs["dlq_event_queue_arn"] = aws.ToString(s3Target.DlqEventQueueArn)

		result = append(result, attrs)
	}
	return result
}

func flattenCatalogTargets(CatalogTargets []awstypes.CatalogTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, catalogTarget := range CatalogTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(catalogTarget.ConnectionName)
		attrs["tables"] = flex.FlattenStringValueList(catalogTarget.Tables)
		attrs[names.AttrDatabaseName] = aws.ToString(catalogTarget.DatabaseName)
		attrs["event_queue_arn"] = aws.ToString(catalogTarget.EventQueueArn)
		attrs["dlq_event_queue_arn"] = aws.ToString(catalogTarget.DlqEventQueueArn)

		result = append(result, attrs)
	}
	return result
}

func flattenDynamoDBTargets(dynamodbTargets []awstypes.DynamoDBTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, dynamodbTarget := range dynamodbTargets {
		attrs := make(map[string]interface{})
		attrs[names.AttrPath] = aws.ToString(dynamodbTarget.Path)
		attrs["scan_all"] = aws.ToBool(dynamodbTarget.ScanAll)
		attrs["scan_rate"] = aws.ToFloat64(dynamodbTarget.ScanRate)

		result = append(result, attrs)
	}
	return result
}

func flattenJDBCTargets(jdbcTargets []awstypes.JdbcTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, jdbcTarget := range jdbcTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(jdbcTarget.ConnectionName)
		attrs["exclusions"] = flex.FlattenStringValueList(jdbcTarget.Exclusions)
		attrs["enable_additional_metadata"] = flex.FlattenStringyValueList(jdbcTarget.EnableAdditionalMetadata)
		attrs[names.AttrPath] = aws.ToString(jdbcTarget.Path)

		result = append(result, attrs)
	}
	return result
}

func flattenMongoDBTargets(mongoDBTargets []awstypes.MongoDBTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, mongoDBTarget := range mongoDBTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(mongoDBTarget.ConnectionName)
		attrs[names.AttrPath] = aws.ToString(mongoDBTarget.Path)
		attrs["scan_all"] = aws.ToBool(mongoDBTarget.ScanAll)

		result = append(result, attrs)
	}
	return result
}

func flattenDeltaTargets(deltaTargets []awstypes.DeltaTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, deltaTarget := range deltaTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(deltaTarget.ConnectionName)
		attrs["create_native_delta_table"] = aws.ToBool(deltaTarget.CreateNativeDeltaTable)
		attrs["delta_tables"] = flex.FlattenStringValueSet(deltaTarget.DeltaTables)
		attrs["write_manifest"] = aws.ToBool(deltaTarget.WriteManifest)

		result = append(result, attrs)
	}
	return result
}

func flattenHudiTargets(hudiTargets []awstypes.HudiTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, hudiTarget := range hudiTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(hudiTarget.ConnectionName)
		attrs["maximum_traversal_depth"] = aws.ToInt32(hudiTarget.MaximumTraversalDepth)
		attrs["paths"] = flex.FlattenStringValueSet(hudiTarget.Paths)
		attrs["exclusions"] = flex.FlattenStringValueList(hudiTarget.Exclusions)

		result = append(result, attrs)
	}
	return result
}

func flattenIcebergTargets(icebergTargets []awstypes.IcebergTarget) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, icebergTarget := range icebergTargets {
		attrs := make(map[string]interface{})
		attrs["connection_name"] = aws.ToString(icebergTarget.ConnectionName)
		attrs["maximum_traversal_depth"] = aws.ToInt32(icebergTarget.MaximumTraversalDepth)
		attrs["paths"] = flex.FlattenStringValueSet(icebergTarget.Paths)
		attrs["exclusions"] = flex.FlattenStringValueList(icebergTarget.Exclusions)

		result = append(result, attrs)
	}
	return result
}

func flattenCrawlerSchemaChangePolicy(cfg *awstypes.SchemaChangePolicy) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"delete_behavior": string(cfg.DeleteBehavior),
		"update_behavior": string(cfg.UpdateBehavior),
	}

	return []map[string]interface{}{m}
}

func expandCrawlerLineageConfiguration(cfg []interface{}) *awstypes.LineageConfiguration {
	m := cfg[0].(map[string]interface{})

	target := &awstypes.LineageConfiguration{
		CrawlerLineageSettings: awstypes.CrawlerLineageSettings(m["crawler_lineage_settings"].(string)),
	}
	return target
}

func flattenCrawlerLineageConfiguration(cfg *awstypes.LineageConfiguration) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"crawler_lineage_settings": string(cfg.CrawlerLineageSettings),
	}

	return []map[string]interface{}{m}
}

func expandLakeFormationConfiguration(cfg []interface{}) *awstypes.LakeFormationConfiguration {
	m := cfg[0].(map[string]interface{})

	target := &awstypes.LakeFormationConfiguration{}

	if v, ok := m[names.AttrAccountID].(string); ok {
		target.AccountId = aws.String(v)
	}

	if v, ok := m["use_lake_formation_credentials"].(bool); ok {
		target.UseLakeFormationCredentials = aws.Bool(v)
	}

	return target
}

func flattenLakeFormationConfiguration(cfg *awstypes.LakeFormationConfiguration) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrAccountID:              aws.ToString(cfg.AccountId),
		"use_lake_formation_credentials": aws.ToBool(cfg.UseLakeFormationCredentials),
	}

	return []map[string]interface{}{m}
}

func expandCrawlerRecrawlPolicy(cfg []interface{}) *awstypes.RecrawlPolicy {
	m := cfg[0].(map[string]interface{})

	target := &awstypes.RecrawlPolicy{
		RecrawlBehavior: awstypes.RecrawlBehavior(m["recrawl_behavior"].(string)),
	}
	return target
}

func flattenCrawlerRecrawlPolicy(cfg *awstypes.RecrawlPolicy) []map[string]interface{} {
	if cfg == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"recrawl_behavior": string(cfg.RecrawlBehavior),
	}

	return []map[string]interface{}{m}
}
