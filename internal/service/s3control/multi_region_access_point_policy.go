package s3control

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountId = v.(string)
	}

	input := &s3control.PutMultiRegionAccessPointPolicyInput{
		AccountId: aws.String(accountId),
		Details:   expandMultiRegionAccessPointPolicyDetails(d.Get("details").([]interface{})[0].(map[string]interface{})),
	}

	name := aws.StringValue(input.Details.Name)
	log.Printf("[DEBUG] Creating S3 Multi-Region Access Point policy: %s", d.Id())
	output, err := conn.PutMultiRegionAccessPointPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Multi-Region Access Point (%s) policy: %s", d.Id(), err)
	}

	requestTokenARN := aws.StringValue(output.RequestTokenARN)
	_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountId, requestTokenARN, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point Policy (%s) to be created: %s", d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, name))

	return resourceMultiRegionAccessPointPolicyRead(d, meta)
}

func resourceMultiRegionAccessPointPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId, name, err := MultiRegionAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	policyDocument, err := FindMultiRegionAccessPointPolicyDocumentByAccountIDAndName(conn, accountId, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Multi-Region Access Point Policy (%s): %w", d.Id(), err)
	}

	d.Set("account_id", accountId)
	d.Set("established", policyDocument.Established.Policy)
	d.Set("proposed", policyDocument.Proposed.Policy)
	d.Set("details", []interface{}{policyDocumentToDetailsMap(aws.String(name), policyDocument)})

	return nil
}

func resourceMultiRegionAccessPointPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId, _, err := MultiRegionAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("details") {
		log.Printf("[DEBUG] Updating S3 Multi-Region Access Point policy: %s", d.Id())
		output, err := conn.PutMultiRegionAccessPointPolicy(&s3control.PutMultiRegionAccessPointPolicyInput{
			AccountId: aws.String(accountId),
			Details:   expandMultiRegionAccessPointPolicyDetails(d.Get("details").([]interface{})[0].(map[string]interface{})),
		})

		if err != nil {
			return fmt.Errorf("error updating S3 Multi-Region Access Point (%s) policy: %s", d.Id(), err)
		}

		requestTokenARN := *output.RequestTokenARN
		_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountId, requestTokenARN, d.Timeout(schema.TimeoutUpdate))

		if err != nil {
			return fmt.Errorf("error waiting for S3 Multi-Region Access Point Policy (%s) to update: %s", d.Id(), err)
		}
	}

	return resourceMultiRegionAccessPointPolicyRead(d, meta)
}

func expandMultiRegionAccessPointPolicyDetails(tfMap map[string]interface{}) *s3control.PutMultiRegionAccessPointPolicyInput_ {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PutMultiRegionAccessPointPolicyInput_{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["policy"].(string); ok {
		apiObject.Policy = aws.String(v)
	}

	return apiObject
}

func policyDocumentToDetailsMap(multiRegionAccessPointName *string, policyDocument *s3control.MultiRegionAccessPointPolicyDocument) map[string]interface{} {
	details := map[string]interface{}{}

	details["name"] = aws.StringValue(multiRegionAccessPointName)
	details["policy"] = aws.StringValue(policyDocument.Proposed.Policy)

	return details
}
