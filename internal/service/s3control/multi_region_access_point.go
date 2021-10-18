package s3control

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))

	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountId = v.(string)
	}

	input := &s3control.CreateMultiRegionAccessPointInput{
		AccountId: aws.String(accountId),
		Details:   expandMultiRegionAccessPointDetails(d.Get("details").([]interface{})[0].(map[string]interface{})),
	}

	name := aws.StringValue(input.Details.Name)
	log.Printf("[DEBUG] Creating S3 Multi-Region Access Point: %s", input)
	output, err := conn.CreateMultiRegionAccessPoint(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Control Multi-Region Access Point (%s): %w", name, err)
	}

	if output == nil {
		return fmt.Errorf("error creating S3 Control Multi-Region Access Point (%s): empty response", name)
	}

	requestTokenARN := aws.StringValue(output.RequestTokenARN)
	_, err = waitS3MultiRegionAccessPointRequestSucceeded(conn, accountId, requestTokenARN, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point (%s) to create: %s", d.Id(), err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountId, name))

	return resourceMultiRegionAccessPointRead(d, meta)
}

func resourceMultiRegionAccessPointRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))

	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId, name, err := MultiRegionAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	output, err := conn.GetMultiRegionAccessPoint(&s3control.GetMultiRegionAccessPointInput{
		AccountId: aws.String(accountId),
		Name:      aws.String(name),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeNoSuchMultiRegionAccessPoint) {
		log.Printf("[WARN] S3 Multi-Region Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Multi-Region Access Point (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error reading S3 Multi-Region Access Point (%s): empty response", d.Id())
	}

	d.Set("account_id", accountId)
	d.Set("alias", output.AccessPoint.Alias)
	d.Set("domain_name", meta.(*conns.AWSClient).PartitionHostname(fmt.Sprintf("%s.accesspoint.s3-global", aws.StringValue(output.AccessPoint.Alias))))
	d.Set("status", output.AccessPoint.Status)

	multiRegionAccessPointARN := arn.ARN{
		AccountID: accountId,
		Partition: meta.(*conns.AWSClient).Partition,
		Resource:  fmt.Sprintf("accesspoint/%s", aws.StringValue(output.AccessPoint.Alias)),
		Service:   "s3",
	}

	d.Set("arn", multiRegionAccessPointARN.String())

	if err := d.Set("details", []interface{}{flattenMultiRegionAccessPointDetails(output.AccessPoint)}); err != nil {
		return fmt.Errorf("error setting details: %s", err)
	}

	return nil
}

func resourceMultiRegionAccessPointDelete(d *schema.ResourceData, meta interface{}) error {
	conn, err := getS3ControlConn(meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("Error getting S3Control Client: %s", err)
	}

	accountId, name, err := MultiRegionAccessPointParseId(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Multi-Region Access Point: %s", d.Id())
	output, err := conn.DeleteMultiRegionAccessPoint(&s3control.DeleteMultiRegionAccessPointInput{
		AccountId: aws.String(accountId),
		Details: &s3control.DeleteMultiRegionAccessPointInput_{
			Name: aws.String(name),
		},
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchMultiRegionAccessPoint) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Multi-Region Access Point (%s): %s", d.Id(), err)
	}

	requestTokenARN := aws.StringValue(output.RequestTokenARN)
	_, err = waitS3MultiRegionAccessPointRequestSucceeded(conn, accountId, requestTokenARN, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for S3 Multi-Region Access Point (%s) to delete: %w", d.Id(), err)
	}

	return nil
}

// MultiRegionAccessPointParseId returns the Account ID and Access Point Name (S3)
func MultiRegionAccessPointParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected ACCOUNT_ID:NAME", id)
	}

	return parts[0], parts[1], nil
}

func expandMultiRegionAccessPointDetails(tfMap map[string]interface{}) *s3control.CreateMultiRegionAccessPointInput_ {
	if tfMap == nil {
		return nil
	}

	apiObject := &s3control.CreateMultiRegionAccessPointInput_{}

	if v, ok := tfMap["name"].(string); ok {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["public_access_block"].([]interface{}); ok && len(v) > 0 {
		apiObject.PublicAccessBlock = expandS3AccessPointPublicAccessBlockConfiguration(v)
	}

	if v, ok := tfMap["region"]; ok {
		apiObject.Regions = expandMultiRegionAccessPointRegions(v.(*schema.Set).List())
	}

	return apiObject
}

func expandMultiRegionAccessPointRegions(tfList []interface{}) []*s3control.Region {
	regions := make([]*s3control.Region, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		value, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		region := &s3control.Region{
			Bucket: aws.String(value["bucket"].(string)),
		}

		regions = append(regions, region)
	}

	return regions
}

func flattenMultiRegionAccessPointDetails(apiObject *s3control.MultiRegionAccessPointReport) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.PublicAccessBlock; v != nil {
		tfMap["public_access_block"] = flattenS3AccessPointPublicAccessBlockConfiguration(v)
	}

	if v := apiObject.Regions; v != nil {
		tfMap["region"] = flattenMultiRegionAccessPointRegions(v)
	}

	return tfMap
}

func flattenMultiRegionAccessPointRegions(apiObjects []*s3control.RegionReport) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if apiObject.Bucket == nil {
			continue
		}

		m := map[string]interface{}{}
		if v := apiObject.Bucket; v != nil {
			m["bucket"] = aws.StringValue(v)
		}

		tfList = append(tfList, m)
	}

	return tfList
}

func getS3ControlConn(awsClient *conns.AWSClient) (*s3control.S3Control, error) {
	if awsClient.S3ControlConn.Config.Region != nil && *awsClient.S3ControlConn.Config.Region == endpoints.UsWest2RegionID {
		return awsClient.S3ControlConn, nil
	}

	sess, err := session.NewSession(&awsClient.S3ControlConn.Config)

	if err != nil {
		return nil, fmt.Errorf("error creating AWS S3Control session: %w", err)
	}

	// Multi-Region Access Point requires requests to be routed to the us-west-2 endpoint
	conn := s3control.New(sess.Copy(&aws.Config{Region: aws.String(endpoints.UsWest2RegionID)}))

	return conn, nil
}
