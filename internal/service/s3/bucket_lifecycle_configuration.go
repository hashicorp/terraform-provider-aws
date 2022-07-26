package s3

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketLifecycleConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketLifecycleConfigurationCreate,
		ReadContext:   resourceBucketLifecycleConfigurationRead,
		UpdateContext: resourceBucketLifecycleConfigurationUpdate,
		DeleteContext: resourceBucketLifecycleConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},

			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},

			"rule": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"abort_incomplete_multipart_upload": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days_after_initiation": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0, // API returns 0
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true, // API returns false; conflicts with date and days
									},
								},
							},
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							// If neither the filter block nor the prefix parameter in the rule are specified,
							// we apply the Default behavior from v3.x of the provider (Filter with empty string Prefix),
							// which will thus return a Filter in the GetBucketLifecycleConfiguration request and
							// require diff suppression.
							DiffSuppressFunc: suppressMissingFilterConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"and": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object_size_greater_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"object_size_less_than": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"prefix": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"tags": tftags.TagsSchema(),
											},
										},
									},
									"object_size_greater_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"object_size_less_than": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"tag": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},

						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},

						"noncurrent_version_expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"newer_noncurrent_versions": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(1),
									},
									"noncurrent_days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.TransitionStorageClass_Values(), false),
									},
								},
							},
						},

						"prefix": {
							Type:       schema.TypeString,
							Optional:   true,
							Deprecated: "Use filter instead",
						},

						"status": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								LifecycleRuleStatusDisabled,
								LifecycleRuleStatusEnabled,
							}, false),
						},

						"transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidUTCTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(s3.TransitionStorageClass_Values(), false),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceBucketLifecycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	rules, err := ExpandLifecycleRules(d.Get("rule").([]interface{}))
	if err != nil {
		return diag.Errorf("error creating S3 Lifecycle Configuration for bucket (%s): %s", bucket, err)
	}

	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfigurationWithContext(ctx, input)
	}, s3.ErrCodeNoSuchBucket)

	if err != nil {
		return diag.Errorf("error creating S3 Lifecycle Configuration for bucket (%s): %s", bucket, err)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	if err = waitForLifecycleConfigurationRulesStatus(ctx, conn, bucket, expectedBucketOwner, rules); err != nil {
		return diag.Errorf("error waiting for S3 Lifecycle Configuration for bucket (%s) to reach expected rules status after update: %s", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	var lastOutput, output *s3.GetBucketLifecycleConfigurationOutput

	err = resource.RetryContext(ctx, lifecycleConfigurationRulesSteadyTimeout, func() *resource.RetryError {
		var err error

		time.Sleep(lifecycleConfigurationExtraRetryDelay)

		output, err = conn.GetBucketLifecycleConfigurationWithContext(ctx, input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if lastOutput == nil || !reflect.DeepEqual(*lastOutput, *output) {
			lastOutput = output
			return resource.RetryableError(fmt.Errorf("bucket lifecycle configuration has not stablized; trying again"))
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.GetBucketLifecycleConfigurationWithContext(ctx, input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Lifecycle Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("rule", FlattenLifecycleRules(output.Rules)); err != nil {
		return diag.Errorf("error setting rule: %s", err)
	}

	return nil
}

func resourceBucketLifecycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	rules, err := ExpandLifecycleRules(d.Get("rule").([]interface{}))
	if err != nil {
		return diag.Errorf("error updating S3 Bucket Lifecycle Configuration rule: %s", err)
	}

	input := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(2*time.Minute, func() (interface{}, error) {
		return conn.PutBucketLifecycleConfigurationWithContext(ctx, input)
	}, ErrCodeNoSuchLifecycleConfiguration)

	if err != nil {
		return diag.Errorf("error updating S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	if err := waitForLifecycleConfigurationRulesStatus(ctx, conn, bucket, expectedBucketOwner, rules); err != nil {
		return diag.Errorf("error waiting for S3 Lifecycle Configuration for bucket (%s) to reach expected rules status after update: %s", d.Id(), err)
	}

	return resourceBucketLifecycleConfigurationRead(ctx, d, meta)
}

func resourceBucketLifecycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.DeleteBucketLifecycleWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting S3 Bucket Lifecycle Configuration (%s): %s", d.Id(), err)
	}

	return nil
}

// suppressMissingFilterConfigurationBlock suppresses the diff that results from an omitted
// filter configuration block and one returned from the S3 API.
// To work around the issue, https://github.com/hashicorp/terraform-plugin-sdk/issues/743,
// this method only looks for changes in the "filter.#" value and not its nested fields
// which are incorrectly suppressed when using the verify.SuppressMissingOptionalConfigurationBlock method.
func suppressMissingFilterConfigurationBlock(k, old, new string, d *schema.ResourceData) bool {
	if strings.HasSuffix(k, "filter.#") {
		o, n := d.GetChange(k)
		oVal, nVal := o.(int), n.(int)

		if oVal == 1 && nVal == 0 {
			return true
		}

		if oVal == 1 && nVal == 1 {
			return old == "1" && new == "0"
		}

		return false
	}
	return false
}
