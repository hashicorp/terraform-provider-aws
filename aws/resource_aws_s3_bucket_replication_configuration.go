package aws

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/s3/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsS3BucketReplicationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsS3BucketReplicationConfigurationPut,
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
									"replication_time": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"status": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{s3.ReplicationTimeStatusEnabled}, false),
												},
												"time": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"minutes": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(0),
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
						"existing_object_replication": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"status": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{s3.ExistingObjectReplicationStatusEnabled}, false),
									},
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
		},
	}
}

func resourceAwsS3BucketReplicationConfigurationPut(d *schema.ResourceData, meta interface{}) error {
	// Get the bucket
	var bucket string
	if v, ok := d.GetOk("bucket"); ok {
		bucket = v.(string)
	} else {
		log.Printf("[ERROR] S3 Bucket name not set")
		return errors.New("[ERROR] S3 Bucket name not set")
	}
	d.SetId(bucket)

	return resourceAwsS3BucketReplicationConfigurationUpdate(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationRead(d *schema.ResourceData, meta interface{}) error {

	if _, ok := d.GetOk("bucket"); !ok {
		// during import operations, use the supplied ID for the bucket name
		d.Set("bucket", d.Id())
	}

	var bucket *string
	input := &s3.HeadBucketInput{}
	if rsp, ok := d.GetOk("bucket"); !ok {
		log.Printf("[ERROR] S3 Bucket name not set")
		return errors.New("[ERROR] S3 Bucket name not set")
	} else {
		bucket = aws.String(rsp.(string))
		input.Bucket = bucket
	}

	s3conn := meta.(*AWSClient).s3conn

	err := resource.Retry(waiter.BucketCreatedTimeout, func() *resource.RetryError {
		_, err := s3conn.HeadBucket(input)

		if d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
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

	if tfresource.TimedOut(err) {
		_, err = s3conn.HeadBucket(input)
	}

	if !d.IsNewResource() && tfawserr.ErrStatusCodeEquals(err, http.StatusNotFound) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket (%s) not found, removing from state", d.Id())
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket (%s): %w", d.Id(), err)
	}

	// Read the bucket replication configuration
	replicationResponse, err := retryOnAwsCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return s3conn.GetBucketReplication(&s3.GetBucketReplicationInput{
			Bucket: bucket,
		})
	})
	if err != nil && !isAWSErr(err, "ReplicationConfigurationNotFoundError", "") {
		return fmt.Errorf("error getting S3 Bucket replication: %s", err)
	}
	replication, ok := replicationResponse.(*s3.GetBucketReplicationOutput)
	if !ok || replication == nil {
		return fmt.Errorf("error reading replication_configuration")
	}
	r := replication.ReplicationConfiguration
	// set role
	if r.Role != nil && aws.StringValue(r.Role) != "" {
		d.Set("role", aws.StringValue(r.Role))
	}

	rules := make([]interface{}, 0, len(r.Rules))
	for _, v := range r.Rules {
		t := make(map[string]interface{})
		if v.Destination != nil {
			rd := make(map[string]interface{})
			if v.Destination.Bucket != nil {
				rd["bucket"] = aws.StringValue(v.Destination.Bucket)
			}
			if v.Destination.StorageClass != nil {
				rd["storage_class"] = aws.StringValue(v.Destination.StorageClass)
			}
			if v.Destination.EncryptionConfiguration != nil {
				if v.Destination.EncryptionConfiguration.ReplicaKmsKeyID != nil {
					rd["replica_kms_key_id"] = aws.StringValue(v.Destination.EncryptionConfiguration.ReplicaKmsKeyID)
				}
			}
			if v.Destination.Account != nil {
				rd["account_id"] = aws.StringValue(v.Destination.Account)
			}
			if v.Destination.AccessControlTranslation != nil {
				rdt := map[string]interface{}{
					"owner": aws.StringValue(v.Destination.AccessControlTranslation.Owner),
				}
				rd["access_control_translation"] = []interface{}{rdt}
			}
			if v.Destination.ReplicationTime != nil {
				if v.Destination.ReplicationTime.Status != nil {
					rd["replication_time"] = map[string]interface{}{
						"status": v.Destination.ReplicationTime.Status,
						"time": map[string]interface{}{
							"minutes": v.Destination.ReplicationTime.Time.Minutes,
						},
					}
				}
			}
			t["destination"] = []interface{}{rd}
		}

		if v.ExistingObjectReplication != nil {
			status := make(map[string]interface{})
			status["status"] = aws.StringValue(v.ExistingObjectReplication.Status)
			t["existing_object_replication"] = status
		}

		if v.ID != nil {
			t["id"] = aws.StringValue(v.ID)
		}
		if v.Prefix != nil {
			t["prefix"] = aws.StringValue(v.Prefix)
		}
		if v.Status != nil {
			t["status"] = aws.StringValue(v.Status)
		}
		if vssc := v.SourceSelectionCriteria; vssc != nil {
			tssc := make(map[string]interface{})
			if vssc.SseKmsEncryptedObjects != nil {
				tSseKms := make(map[string]interface{})
				if aws.StringValue(vssc.SseKmsEncryptedObjects.Status) == s3.SseKmsEncryptedObjectsStatusEnabled {
					tSseKms["enabled"] = true
				} else if aws.StringValue(vssc.SseKmsEncryptedObjects.Status) == s3.SseKmsEncryptedObjectsStatusDisabled {
					tSseKms["enabled"] = false
				}
				tssc["sse_kms_encrypted_objects"] = []interface{}{tSseKms}
			}
			t["source_selection_criteria"] = []interface{}{tssc}
		}

		if v.Priority != nil {
			t["priority"] = int(aws.Int64Value(v.Priority))
		}

		if f := v.Filter; f != nil {
			m := map[string]interface{}{}
			if f.Prefix != nil {
				m["prefix"] = aws.StringValue(f.Prefix)
			}
			if t := f.Tag; t != nil {
				m["tags"] = keyvaluetags.S3KeyValueTags([]*s3.Tag{t}).IgnoreAws().Map()
			}
			if a := f.And; a != nil {
				m["prefix"] = aws.StringValue(a.Prefix)
				m["tags"] = keyvaluetags.S3KeyValueTags(a.Tags).IgnoreAws().Map()
			}
			t["filter"] = []interface{}{m}

			if v.DeleteMarkerReplication != nil && v.DeleteMarkerReplication.Status != nil && aws.StringValue(v.DeleteMarkerReplication.Status) == s3.DeleteMarkerReplicationStatusEnabled {
				t["delete_marker_replication_status"] = aws.StringValue(v.DeleteMarkerReplication.Status)
			}
		}

		rules = append(rules, t)
	}
	d.Set("rules", schema.NewSet(rulesHash, rules))

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

		eor := rr["existing_object_replication"].([]interface{})
		if len(eor) > 0 {
			s := eor[0].(map[string]interface{})
			rcRule.ExistingObjectReplication = &s3.ExistingObjectReplication{
				Status: aws.String(s["status"].(string)),
			}
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

				rt, ok := bd["replication_time"].([]interface{})
				if ok && len(rt) > 0 {
					s := rt[0].(map[string]interface{})
					if t, ok := s["time"].([]interface{}); ok && len(t) > 0 {
						m := t[0].(map[string]interface{})
						ruleDestination.ReplicationTime = &s3.ReplicationTime{
							Status: aws.String(s["status"].(string)),
							Time: &s3.ReplicationTimeValue{
								Minutes: aws.Int64(int64(m["minutes"].(int))),
							},
						}
					}
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

	return resourceAwsS3BucketReplicationConfigurationRead(d, meta)
}

func resourceAwsS3BucketReplicationConfigurationDelete(d *schema.ResourceData, meta interface{}) error {

	return nil
}
