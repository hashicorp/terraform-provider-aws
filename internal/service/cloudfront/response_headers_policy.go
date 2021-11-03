package cloudfront

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceResponseHeadersPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceResponseHeadersPolicyCreate,
		Read:   resourceResponseHeadersPolicyRead,
		Update: resourceResponseHeadersPolicyUpdate,
		Delete: resourceResponseHeadersPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"comment": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cors_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access_control_allow_credentials": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"access_control_allow_headers": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"access_control_allow_methods": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"access_control_allow_origins": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"access_control_expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"items": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"access_control_max_age_sec": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"origin_override": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"custom_headers_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"items": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"etag": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"security_headers_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_security_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"content_security_policy": {
										Type:     schema.TypeString,
										Required: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"content_type_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"frame_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"frame_option": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"DENY", "SAMEORIGIN"}, false),
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"referrer_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"referrer_policy": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"no-referrer", "no-referrer-when-downgrade", "origin", "origin-when-cross-origin", "same-origin", "strict-origin", "strict-origin-when-cross-origin", "unsafe-url"}, false),
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"strict_transport_security": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_max_age_sec": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"include_subdomains": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"preload": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"xss_protection": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode_block": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"override": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"protection": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"report_uri": {
										Type:     schema.TypeString,
										Optional: true,
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

func resourceResponseHeadersPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	request := &cloudfront.CreateResponseHeadersPolicyInput{
		ResponseHeadersPolicyConfig: expandCloudFrontResponseHeadersPolicyConfig(d),
	}

	resp, err := conn.CreateResponseHeadersPolicy(request)

	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(resp.ResponseHeadersPolicy.Id))

	return resourceResponseHeadersPolicyRead(d, meta)
}

func resourceResponseHeadersPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	output, err := FindResponseHeadersPolicyByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Response Headers Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFront Response Headers Policy (%s): %w", d.Id(), err)
	}

	d.Set("etag", output.ETag)

	if err := setCloudFrontResponseHeadersPolicy(d, output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig); err != nil {
		return fmt.Errorf("unable to store response headers policy in config: %s", err.Error())
	}

	return nil
}

func resourceResponseHeadersPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	request := &cloudfront.UpdateResponseHeadersPolicyInput{
		ResponseHeadersPolicyConfig: expandCloudFrontResponseHeadersPolicyConfig(d),
		Id:                          aws.String(d.Id()),
		IfMatch:                     aws.String(d.Get("etag").(string)),
	}

	_, err := conn.UpdateResponseHeadersPolicy(request)
	if err != nil {
		return err
	}

	return resourceResponseHeadersPolicyRead(d, meta)
}

func resourceResponseHeadersPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	log.Printf("[DEBUG] Deleting CloudFront Response Headers Policy: (%s)", d.Id())
	_, err := conn.DeleteResponseHeadersPolicy(&cloudfront.DeleteResponseHeadersPolicyInput{
		Id:      aws.String(d.Id()),
		IfMatch: aws.String(d.Get("etag").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResponseHeadersPolicy) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFront Response Headers Policy (%s): %w", d.Id(), err)
	}

	return nil
}
