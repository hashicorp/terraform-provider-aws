package s3control

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMultiRegionAccessPoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceMultiRegionAccessPointCreate,
		Read:   resourceMultiRegionAccessPointRead,
		Delete: resourceMultiRegionAccessPointDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"details": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validateS3MultiRegionAccessPointName,
						},
						"public_access_block": {
							Type:             schema.TypeList,
							Optional:         true,
							ForceNew:         true,
							MinItems:         0,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_public_acls": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"block_public_policy": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"ignore_public_acls": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"restrict_public_buckets": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
								},
							},
						},
						"region": {
							Type:     schema.TypeSet,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							MaxItems: 20,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(3, 255),
									},
								},
							},
						},
					},
				},
			},
			"domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMultiRegionAccessPointCreate(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.CreateMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
	}

	if v, ok := d.GetOk("details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Details = expandCreateMultiRegionAccessPointInput_(v.([]interface{})[0].(map[string]interface{}))
	}

	resourceID := MultiRegionAccessPointCreateResourceID(accountID, aws.StringValue(input.Details.Name))

	log.Printf("[DEBUG] Creating S3 Multi-Region Access Point: %s", input)
	output, err := conn.CreateMultiRegionAccessPoint(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Multi-Region Access Point (%s): %w", resourceID, err)
	}

	d.SetId(resourceID)

	_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for Multi-Region Access Point (%s) create: %s", d.Id(), err)
	}

	return resourceMultiRegionAccessPointRead(d, meta)
}

func resourceMultiRegionAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	accessPoint, err := FindMultiRegionAccessPointByAccountIDAndName(conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Multi-Region Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Multi-Region Access Point (%s): %w", d.Id(), err)
	}

	alias := aws.StringValue(accessPoint.Alias)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "s3",
		AccountID: accountID,
		Resource:  fmt.Sprintf("accesspoint/%s", alias),
	}.String()
	d.Set("account_id", accountID)
	d.Set("alias", alias)
	d.Set("arn", arn)
	if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointReport(accessPoint)}); err != nil {
		return fmt.Errorf("error setting details: %w", err)
	}
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide//MultiRegionAccessPointRequests.html#MultiRegionAccessPointHostnames.
	d.Set("domain_name", meta.(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.accesspoint.s3-global", alias)))
	d.Set("status", accessPoint.Status)

	return nil
}

func resourceMultiRegionAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn, err := ConnForMRAP(meta.(*conns.AWSClient))

	if err != nil {
		return err
	}

	accountID, name, err := MultiRegionAccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Multi-Region Access Point: %s", d.Id())
	output, err := conn.DeleteMultiRegionAccessPoint(&s3control.DeleteMultiRegionAccessPointInput{
		AccountId: aws.String(accountID),
		Details: &s3control.DeleteMultiRegionAccessPointInput_{
			Name: aws.String(name),
		},
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchMultiRegionAccessPoint) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Multi-Region Access Point (%s): %w", d.Id(), err)
	}

	_, err = waitMultiRegionAccessPointRequestSucceeded(conn, accountID, aws.StringValue(output.RequestTokenARN), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func ConnForMRAP(client *conns.AWSClient) (*s3control.S3Control, error) {
	originalConn := client.S3ControlConn
	// All Multi-Region Access Point actions are routed to the US West (Oregon) Region.
	region := endpoints.UsWest2RegionID

	if originalConn.Config.Region != nil && aws.StringValue(originalConn.Config.Region) == region {
		return originalConn, nil
	}

	sess, err := conns.NewSessionForRegion(&originalConn.Config, region, client.TerraformVersion)

	if err != nil {
		return nil, fmt.Errorf("error creating AWS session: %w", err)
	}

	return s3control.New(sess), nil
}

const multiRegionAccessPointResourceIDSeparator = ":"

func MultiRegionAccessPointCreateResourceID(accountID, accessPointName string) string {
	parts := []string{accountID, accessPointName}
	id := strings.Join(parts, multiRegionAccessPointResourceIDSeparator)

	return id
}

func MultiRegionAccessPointParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, multiRegionAccessPointResourceIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected account-id%[2]saccess-point-name", id, multiRegionAccessPointResourceIDSeparator)
}

func expandCreateMultiRegionAccessPointInput_(tfMap map[string]interface{}) *s3control.CreateMultiRegionAccessPointInput_ {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.CreateMultiRegionAccessPointInput_{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["public_access_block"].([]interface{}); ok && len(v) > 0 {
		apiObject.PublicAccessBlock = expandPublicAccessBlockConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["region"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Regions = expandRegions(v.List())
	}

	return apiObject
}

func expandPublicAccessBlockConfiguration(tfMap map[string]interface{}) *s3control.PublicAccessBlockConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.PublicAccessBlockConfiguration{}

	if v, ok := tfMap["block_public_acls"].(bool); ok {
		apiObject.BlockPublicAcls = aws.Bool(v)
	}

	if v, ok := tfMap["block_public_policy"].(bool); ok {
		apiObject.BlockPublicPolicy = aws.Bool(v)
	}

	if v, ok := tfMap["ignore_public_acls"].(bool); ok {
		apiObject.IgnorePublicAcls = aws.Bool(v)
	}

	if v, ok := tfMap["restrict_public_buckets"].(bool); ok {
		apiObject.RestrictPublicBuckets = aws.Bool(v)
	}

	return apiObject
}

func expandRegion(tfMap map[string]interface{}) *s3control.Region {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.Region{}

	if v, ok := tfMap["bucket"].(string); ok {
		apiObject.Bucket = aws.String(v)
	}

	return apiObject
}

func expandRegions(tfList []interface{}) []*s3control.Region {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*s3control.Region

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandRegion(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenMultiRegionAccessPointReport(apiObject *s3control.MultiRegionAccessPointReport) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.PublicAccessBlock; v != nil {
		tfMap["public_access_block"] = []interface{}{flattenPublicAccessBlockConfiguration(v)}
	}

	if v := apiObject.Regions; v != nil {
		tfMap["region"] = flattenRegionReports(v)
	}

	return tfMap
}

func flattenPublicAccessBlockConfiguration(apiObject *s3control.PublicAccessBlockConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockPublicAcls; v != nil {
		tfMap["block_public_acls"] = aws.BoolValue(v)
	}

	if v := apiObject.BlockPublicPolicy; v != nil {
		tfMap["block_public_policy"] = aws.BoolValue(v)
	}

	if v := apiObject.IgnorePublicAcls; v != nil {
		tfMap["ignore_public_acls"] = aws.BoolValue(v)
	}

	if v := apiObject.RestrictPublicBuckets; v != nil {
		tfMap["restrict_public_buckets"] = aws.BoolValue(v)
	}

	return tfMap
}

func flattenRegionReport(apiObject *s3control.RegionReport) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Bucket; v != nil {
		tfMap["bucket"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenRegionReports(apiObjects []*s3control.RegionReport) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenRegionReport(apiObject))
	}

	return tfList
}
