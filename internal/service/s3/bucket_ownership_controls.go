package s3

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceBucketOwnershipControls() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketOwnershipControlsCreate,
		Read:   resourceBucketOwnershipControlsRead,
		Update: resourceBucketOwnershipControlsUpdate,
		Delete: resourceBucketOwnershipControlsDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_ownership": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.ObjectOwnership_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceBucketOwnershipControlsCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(bucket),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Bucket (%s) Ownership Controls: %w", bucket, err)
	}

	d.SetId(bucket)

	return resourceBucketOwnershipControlsRead(d, meta)
}

func resourceBucketOwnershipControlsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.GetBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	output, err := conn.GetBucketOwnershipControls(input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		log.Printf("[WARN] S3 Bucket Ownership Controls (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading S3 Bucket (%s) Ownership Controls: empty response", d.Id())
	}

	d.Set("bucket", d.Id())

	if output.OwnershipControls == nil {
		d.Set("rule", nil)
	} else {
		if err := d.Set("rule", flattenOwnershipControlsRules(output.OwnershipControls.Rules)); err != nil {
			return fmt.Errorf("error setting rule: %w", err)
		}
	}

	return nil
}

func resourceBucketOwnershipControlsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.PutBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
		OwnershipControls: &s3.OwnershipControls{
			Rules: expandOwnershipControlsRules(d.Get("rule").([]interface{})),
		},
	}

	_, err := conn.PutBucketOwnershipControls(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	return resourceBucketOwnershipControlsRead(d, meta)
}

func resourceBucketOwnershipControlsDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3Conn

	input := &s3.DeleteBucketOwnershipControlsInput{
		Bucket: aws.String(d.Id()),
	}

	_, err := conn.DeleteBucketOwnershipControls(input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrCodeEquals(err, "OwnershipControlsNotFoundError") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Bucket (%s) Ownership Controls: %w", d.Id(), err)
	}

	return nil
}

func expandOwnershipControlsRules(tfList []interface{}) []*s3.OwnershipControlsRule {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []*s3.OwnershipControlsRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandOwnershipControlsRule(tfMap))
	}

	return apiObjects
}

func expandOwnershipControlsRule(tfMap map[string]interface{}) *s3.OwnershipControlsRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3.OwnershipControlsRule{}

	if v, ok := tfMap["object_ownership"].(string); ok && v != "" {
		apiObject.ObjectOwnership = aws.String(v)
	}

	return apiObject
}

func flattenOwnershipControlsRules(apiObjects []*s3.OwnershipControlsRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenOwnershipControlsRule(apiObject))
	}

	return tfList
}

func flattenOwnershipControlsRule(apiObject *s3.OwnershipControlsRule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ObjectOwnership; v != nil {
		tfMap["object_ownership"] = aws.StringValue(v)
	}

	return tfMap
}
