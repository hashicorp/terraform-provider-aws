package s3control

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMultiRegionAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceMultiRegionAccessPointPolicyCreate,
		Read:   resourceMultiRegionAccessPointPolicyRead,
		Update: resourceMultiRegionAccessPointPolicyUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
			Update: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"details": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateS3MultiRegionAccessPointName,
						},
						"policy": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
							StateFunc: func(v interface{}) string {
								json, _ := structure.NormalizeJsonString(v)
								return json
							},
						},
					},
				},
			},
			"established": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"proposed": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMultiRegionAccessPointPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	resourceID := MultiRegionAccessPointCreateResourceID(accountID, aws.StringValue(input.Details.Name))

	log.Printf("[DEBUG] Creating S3 Multi-Region Access Point Policy: %s", input)
	output, err := conn.PutMultiRegionAccessPointPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Multi-Region Access Point (%s) Policy: %w", resourceID, err)
	}

	d.SetId(resourceID)

	_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point Policy (%s) create: %w", d.Id(), err)
	}

	return resourceMultiRegionAccessPointPolicyRead(d, meta)
}

func resourceMultiRegionAccessPointPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	policyDocument, err := FindMultiRegionAccessPointPolicyDocumentByAccountIDAndName(conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Multi-Region Access Point Policy (%s): %w", d.Id(), err)
	}

	d.Set("account_id", accountID)
	if policyDocument != nil {
		var oldDetails map[string]interface{}
		if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			oldDetails = v.([]interface{})[0].(map[string]interface{})
		}

		if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointPolicyDocument(name, policyDocument, oldDetails)}); err != nil {
			return fmt.Errorf("error setting details: %w", err)
		}
	} else {
		d.Set("details", nil)
	}
	if v := policyDocument.Established; v != nil {
		d.Set("established", v.Policy)
	} else {
		d.Set("established", nil)
	}
	if v := policyDocument.Proposed; v != nil {
		d.Set("proposed", v.Policy)
	} else {
		d.Set("proposed", nil)
	}

	return nil
}

func resourceMultiRegionAccessPointPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID, _, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandPutMultiRegionAccessPointPolicyInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	log.Printf("[DEBUG] Updating S3 Multi-Region Access Point Policy: %s", input)
	output, err := conn.PutMultiRegionAccessPointPolicy(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Multi-Region Access Point Policy (%s): %w", d.Id(), err)
	}

	_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point Policy (%s) update: %w", d.Id(), err)
	}

	return resourceMultiRegionAccessPointPolicyRead(d, meta)
}

func expandPutMultiRegionAccessPointPolicyInput_(tfMap map[string]interface{}) *s3control.PutMultiRegionAccessPointPolicyInput_ {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PutMultiRegionAccessPointPolicyInput_{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["policy"].(string); ok {
		policy, err := structure.NormalizeJsonString(v)

		if err != nil {
			policy = v
		}

		apiObject.Policy = aws.String(policy)
	}

	return apiObject
}

func flattenMultiRegionAccessPointPolicyDocument(name string, apiObject *s3control.MultiRegionAccessPointPolicyDocument, old map[string]interface{}) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["name"] = name

	if v := apiObject.Proposed; v != nil {
		if v := v.Policy; v != nil {
			policyToSet := aws.StringValue(v)
			if old != nil {
				if w, ok := old["policy"].(string); ok {
					var err error
					policyToSet, err = verify.PolicyToSet(w, aws.StringValue(v))

					if err != nil {
						policyToSet = aws.StringValue(v)
					}
				}
			}
			tfMap["policy"] = policyToSet
		}
	}

	return tfMap
}
