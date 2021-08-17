package aws

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsS3BucketReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketReplicationConfigurationCreate,
		Read:   resourceAwsS3BucketReplicationConfigurationRead,
		Update: resourceAwsS3BucketReplicationConfigurationUpdate,
		Delete: resourceAwsS3BucketReplicationConfigurationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 63),
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rules": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      rulesHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"destination": {
							Type:     schema.TypeList,
							MaxItems: 1,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateAwsAccountId,
									},
									"bucket": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validateArn,
									},
									"storage_class": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(s3.StorageClass_Values(), false),
									},
									"replica_kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"access_control_translation": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"owner": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(s3.OwnerOverride_Values(), false),
												},
											},
										},
									},
								},
							},
						},
						"source_selection_criteria": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"sse_kms_encrypted_objects": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
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
								},
							},
						},
						"prefix": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 1024),
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ReplicationRuleStatus_Values(), false),
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"prefix": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 1024),
									},
									"tags": tagsSchema(),
								},
							},
						},
						"delete_marker_replication_status": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{s3.DeleteMarkerReplicationStatusEnabled}, false),
						},
					},
				},
			},

			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},

		CustomizeDiff: SetTagsDiff,
	}
}

func resourceAwsS3BucketReplicationConfigurationCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceAwsS3BucketReplicationConfigurationUpdate(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn

	input := &s3.HeadBucketInput{
		Bucket: aws.String(d.Get("bucket").(string)),
	}

	err := resource.Retry(s3BucketCreationTimeout, func() *resource.RetryError {
		_, err := s3conn.HeadBucket(input)

		if d.IsNewResource() && isAWSErrRequestFailureStatusCode(err, 404) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})
	// Read the bucket replication configuration
	replicationResponse, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return s3conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: aws.String(d.Get("bucket").(string)),
		})
	})
	if err != nil && !isAWSErr(err, "ReplicationConfigurationNotFoundError", "") {
		return fmt.Errorf("error getting S3 Bucket replication: %s", err)
	}
	replication, ok := replicationResponse.(*s3.GetBucketReplicationOutput)
	if !ok || replication == nil {
		return fmt.Errorf("error reading replication_configuration")
	}

	return nil
}

func resourceAwsS3BucketReplicationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	s3conn := meta.(*AWSClient).s3conn
	bucket := d.Get("bucket").(string)

	rc := &s3.ReplicationConfiguration{}
	if val, ok := d.GetOk("role"); ok {
		rc.Role = aws.String(val.(string))
	}

	rcRules := d.Get("rules").(*schema.Set).List()
	rules := []*s3.ReplicationRule{}
	for _, v := range rcRules {
		rr := v.(map[string]interface{})
		rcRule := &s3.ReplicationRule{}
		if status, ok := rr["status"]; ok && status != "" {
			rcRule.Status = aws.String(status.(string))
		} else {
			continue
		}

		if rrid, ok := rr["id"]; ok && rrid != "" {
			rcRule.ID = aws.String(rrid.(string))
		}

		ruleDestination := &s3.Destination{}
		if dest, ok := rr["destination"].([]interface{}); ok && len(dest) > 0 {
			if dest[0] != nil {
				bd := dest[0].(map[string]interface{})
				ruleDestination.Bucket = aws.String(bd["bucket"].(string))

				if storageClass, ok := bd["storage_class"]; ok && storageClass != "" {
					ruleDestination.StorageClass = aws.String(storageClass.(string))
				}

				if replicaKmsKeyId, ok := bd["replica_kms_key_id"]; ok && replicaKmsKeyId != "" {
					ruleDestination.EncryptionConfiguration = &s3.EncryptionConfiguration{
						ReplicaKmsKeyID: aws.String(replicaKmsKeyId.(string)),
					}
				}

				if account, ok := bd["account_id"]; ok && account != "" {
					ruleDestination.Account = aws.String(account.(string))
				}

				if aclTranslation, ok := bd["access_control_translation"].([]interface{}); ok && len(aclTranslation) > 0 {
					aclTranslationValues := aclTranslation[0].(map[string]interface{})
					ruleAclTranslation := &s3.AccessControlTranslation{}
					ruleAclTranslation.Owner = aws.String(aclTranslationValues["owner"].(string))
					ruleDestination.AccessControlTranslation = ruleAclTranslation
				}
			}
		}
		rcRule.Destination = ruleDestination

		if ssc, ok := rr["source_selection_criteria"].([]interface{}); ok && len(ssc) > 0 {
			if ssc[0] != nil {
				sscValues := ssc[0].(map[string]interface{})
				ruleSsc := &s3.SourceSelectionCriteria{}
				if sseKms, ok := sscValues["sse_kms_encrypted_objects"].([]interface{}); ok && len(sseKms) > 0 {
					if sseKms[0] != nil {
						sseKmsValues := sseKms[0].(map[string]interface{})
						sseKmsEncryptedObjects := &s3.SseKmsEncryptedObjects{}
						if sseKmsValues["enabled"].(bool) {
							sseKmsEncryptedObjects.Status = aws.String(s3.SseKmsEncryptedObjectsStatusEnabled)
						} else {
							sseKmsEncryptedObjects.Status = aws.String(s3.SseKmsEncryptedObjectsStatusDisabled)
						}
						ruleSsc.SseKmsEncryptedObjects = sseKmsEncryptedObjects
					}
				}
				rcRule.SourceSelectionCriteria = ruleSsc
			}
		}

		if f, ok := rr["filter"].([]interface{}); ok && len(f) > 0 && f[0] != nil {
			// XML schema V2.
			rcRule.Priority = aws.Int64(int64(rr["priority"].(int)))
			rcRule.Filter = &s3.ReplicationRuleFilter{}
			filter := f[0].(map[string]interface{})
			tags := keyvaluetags.New(filter["tags"]).IgnoreAws().S3Tags()
			if len(tags) > 0 {
				rcRule.Filter.And = &s3.ReplicationRuleAndOperator{
					Prefix: aws.String(filter["prefix"].(string)),
					Tags:   tags,
				}
			} else {
				rcRule.Filter.Prefix = aws.String(filter["prefix"].(string))
			}

			if dmr, ok := rr["delete_marker_replication_status"].(string); ok && dmr != "" {
				rcRule.DeleteMarkerReplication = &s3.DeleteMarkerReplication{
					Status: aws.String(dmr),
				}
			} else {
				rcRule.DeleteMarkerReplication = &s3.DeleteMarkerReplication{
					Status: aws.String(s3.DeleteMarkerReplicationStatusDisabled),
				}
			}
		} else {
			// XML schema V1.
			rcRule.Prefix = aws.String(rr["prefix"].(string))
		}

		rules = append(rules, rcRule)
	}

	rc.Rules = rules
	i := &s3.PutBucketReplicationInput{
		Bucket:                   aws.String(bucket),
		ReplicationConfiguration: rc,
	}
	log.Printf("[DEBUG] S3 put bucket replication configuration: %#v", i)

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketReplication(i)
		if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") || isAWSErr(err, "InvalidRequest", "Versioning must be 'Enabled' on the bucket") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = s3conn.PutBucketReplication(i)
	}
	if err != nil {
		return fmt.Errorf("Error putting S3 replication configuration: %s", err)
	}

	return nil
	return resourceAwsS3BucketReplicationConfigurationRead(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}
