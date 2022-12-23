package cloudfront

import (
	"context"
	"errors"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceOriginAccessControl() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceOriginAccessControlCreate,
		ReadWithoutTimeout:   resourceOriginAccessControlRead,
		UpdateWithoutTimeout: resourceOriginAccessControlUpdate,
		DeleteWithoutTimeout: resourceOriginAccessControlDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "Managed by Terraform",
				ValidateFunc: validation.StringLenBetween(0, 256),
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"origin_access_control_origin_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginAccessControlOriginTypes_Values(), false),
			},
			"signing_behavior": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginAccessControlSigningBehaviors_Values(), false),
			},
			"signing_protocol": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudfront.OriginAccessControlSigningProtocols_Values(), false),
			},
		},
	}
}

const (
	ResNameOriginAccessControl = "Origin Access Control"
)

func resourceOriginAccessControlCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	in := &cloudfront.CreateOriginAccessControlInput{
		OriginAccessControlConfig: &cloudfront.OriginAccessControlConfig{
			Description:                   aws.String(d.Get("description").(string)),
			Name:                          aws.String(d.Get("name").(string)),
			OriginAccessControlOriginType: aws.String(d.Get("origin_access_control_origin_type").(string)),
			SigningBehavior:               aws.String(d.Get("signing_behavior").(string)),
			SigningProtocol:               aws.String(d.Get("signing_protocol").(string)),
		},
	}

	out, err := conn.CreateOriginAccessControlWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.CloudFront, create.ErrActionCreating, ResNameOriginAccessControl, d.Get("name").(string), err)
	}

	if out == nil || out.OriginAccessControl == nil {
		return create.DiagError(names.CloudFront, create.ErrActionCreating, ResNameOriginAccessControl, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.OriginAccessControl.Id))

	return resourceOriginAccessControlRead(ctx, d, meta)
}

func resourceOriginAccessControlRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	out, err := findOriginAccessControlByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Origin Access Control (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.CloudFront, create.ErrActionReading, ResNameOriginAccessControl, d.Id(), err)
	}

	if out.OriginAccessControl == nil || out.OriginAccessControl.OriginAccessControlConfig == nil {
		return create.DiagError(names.CloudFront, create.ErrActionReading, ResNameOriginAccessControl, d.Id(), errors.New("empty output"))
	}

	config := out.OriginAccessControl.OriginAccessControlConfig

	d.Set("description", config.Description)
	d.Set("etag", out.ETag)
	d.Set("name", config.Name)
	d.Set("origin_access_control_origin_type", config.OriginAccessControlOriginType)
	d.Set("signing_behavior", config.SigningBehavior)
	d.Set("signing_protocol", config.SigningProtocol)

	return nil
}

func resourceOriginAccessControlUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	in := &cloudfront.UpdateOriginAccessControlInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
		OriginAccessControlConfig: &cloudfront.OriginAccessControlConfig{
			Description:                   aws.String(d.Get("description").(string)),
			Name:                          aws.String(d.Get("name").(string)),
			OriginAccessControlOriginType: aws.String(d.Get("origin_access_control_origin_type").(string)),
			SigningBehavior:               aws.String(d.Get("signing_behavior").(string)),
			SigningProtocol:               aws.String(d.Get("signing_protocol").(string)),
		},
	}

	log.Printf("[DEBUG] Updating CloudFront Origin Access Control (%s): %#v", d.Id(), in)
	_, err := conn.UpdateOriginAccessControlWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.CloudFront, create.ErrActionUpdating, ResNameOriginAccessControl, d.Id(), err)
	}

	return resourceOriginAccessControlRead(ctx, d, meta)
}

func resourceOriginAccessControlDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	log.Printf("[INFO] Deleting CloudFront Origin Access Control %s", d.Id())

	_, err := conn.DeleteOriginAccessControlWithContext(ctx, &cloudfront.DeleteOriginAccessControlInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchOriginAccessControl) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.CloudFront, create.ErrActionDeleting, ResNameOriginAccessControl, d.Id(), err)
	}

	return nil
}

func findOriginAccessControlByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetOriginAccessControlOutput, error) {
	in := &cloudfront.GetOriginAccessControlInput{
		Id: aws.String(id),
	}
	out, err := conn.GetOriginAccessControlWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchOriginAccessControl) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.OriginAccessControl == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
