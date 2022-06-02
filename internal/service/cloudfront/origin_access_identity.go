package cloudfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceOriginAccessIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceOriginAccessIdentityCreate,
		Read:   resourceOriginAccessIdentityRead,
		Update: resourceOriginAccessIdentityUpdate,
		Delete: resourceOriginAccessIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_access_identity_path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"s3_canonical_user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOriginAccessIdentityCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	params := &cloudfront.CreateCloudFrontOriginAccessIdentityInput{
		CloudFrontOriginAccessIdentityConfig: expandOriginAccessIdentityConfig(d),
	}

	resp, err := conn.CreateCloudFrontOriginAccessIdentity(params)
	if err != nil {
		return err
	}
	d.SetId(aws.StringValue(resp.CloudFrontOriginAccessIdentity.Id))
	return resourceOriginAccessIdentityRead(d, meta)
}

func resourceOriginAccessIdentityRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	params := &cloudfront.GetCloudFrontOriginAccessIdentityInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetCloudFrontOriginAccessIdentity(params)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchCloudFrontOriginAccessIdentity) {
		names.LogNotFoundRemoveState(names.CloudFront, names.ErrActionReading, ResOriginAccessIdentity, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CloudFront, names.ErrActionReading, ResOriginAccessIdentity, d.Id(), err)
	}

	// Update attributes from DistributionConfig
	flattenOriginAccessIdentityConfig(d, resp.CloudFrontOriginAccessIdentity.CloudFrontOriginAccessIdentityConfig)
	// Update other attributes outside of DistributionConfig
	d.SetId(aws.StringValue(resp.CloudFrontOriginAccessIdentity.Id))
	d.Set("etag", resp.ETag)
	d.Set("s3_canonical_user_id", resp.CloudFrontOriginAccessIdentity.S3CanonicalUserId)
	d.Set("cloudfront_access_identity_path", fmt.Sprintf("origin-access-identity/cloudfront/%s", *resp.CloudFrontOriginAccessIdentity.Id))
	iamArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "iam",
		AccountID: "cloudfront",
		Resource:  fmt.Sprintf("user/CloudFront Origin Access Identity %s", *resp.CloudFrontOriginAccessIdentity.Id),
	}.String()
	d.Set("iam_arn", iamArn)
	return nil
}

func resourceOriginAccessIdentityUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	params := &cloudfront.UpdateCloudFrontOriginAccessIdentityInput{
		Id:                                   aws.String(d.Id()),
		CloudFrontOriginAccessIdentityConfig: expandOriginAccessIdentityConfig(d),
		IfMatch:                              aws.String(d.Get("etag").(string)),
	}
	_, err := conn.UpdateCloudFrontOriginAccessIdentity(params)
	if err != nil {
		return err
	}

	return resourceOriginAccessIdentityRead(d, meta)
}

func resourceOriginAccessIdentityDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn
	params := &cloudfront.DeleteCloudFrontOriginAccessIdentityInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	}

	_, err := conn.DeleteCloudFrontOriginAccessIdentity(params)

	return err
}

func expandOriginAccessIdentityConfig(d *schema.ResourceData) *cloudfront.OriginAccessIdentityConfig {
	originAccessIdentityConfig := &cloudfront.OriginAccessIdentityConfig{
		Comment: aws.String(d.Get("comment").(string)),
	}
	// This sets CallerReference if it's still pending computation (ie: new resource)
	if v, ok := d.GetOk("caller_reference"); !ok {
		originAccessIdentityConfig.CallerReference = aws.String(resource.UniqueId())
	} else {
		originAccessIdentityConfig.CallerReference = aws.String(v.(string))
	}
	return originAccessIdentityConfig
}

func flattenOriginAccessIdentityConfig(d *schema.ResourceData, originAccessIdentityConfig *cloudfront.OriginAccessIdentityConfig) {
	if originAccessIdentityConfig.Comment != nil {
		d.Set("comment", originAccessIdentityConfig.Comment)
	}
	d.Set("caller_reference", originAccessIdentityConfig.CallerReference)
}
