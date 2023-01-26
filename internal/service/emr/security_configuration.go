package emr

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceSecurityConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSecurityConfigurationCreate,
		ReadWithoutTimeout:   resourceSecurityConfigurationRead,
		DeleteWithoutTimeout: resourceSecurityConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 10280),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(0, 10280-resource.UniqueIDSuffixLength),
			},

			"configuration": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsJSON,
			},

			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSecurityConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	var emrSCName string
	if v, ok := d.GetOk("name"); ok {
		emrSCName = v.(string)
	} else {
		if v, ok := d.GetOk("name_prefix"); ok {
			emrSCName = resource.PrefixedUniqueId(v.(string))
		} else {
			emrSCName = resource.PrefixedUniqueId("tf-emr-sc-")
		}
	}

	resp, err := conn.CreateSecurityConfigurationWithContext(ctx, &emr.CreateSecurityConfigurationInput{
		Name:                  aws.String(emrSCName),
		SecurityConfiguration: aws.String(d.Get("configuration").(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Security Configuration (%s): %s", emrSCName, err)
	}

	d.SetId(aws.StringValue(resp.Name))
	return append(diags, resourceSecurityConfigurationRead(ctx, d, meta)...)
}

func resourceSecurityConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	resp, err := conn.DescribeSecurityConfigurationWithContext(ctx, &emr.DescribeSecurityConfigurationInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
			log.Printf("[WARN] EMR Security Configuration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading EMR Security Configuration (%s): %s", d.Id(), err)
	}

	d.Set("creation_date", aws.TimeValue(resp.CreationDateTime).Format(time.RFC3339))
	d.Set("name", resp.Name)
	d.Set("configuration", resp.SecurityConfiguration)

	return diags
}

func resourceSecurityConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	_, err := conn.DeleteSecurityConfigurationWithContext(ctx, &emr.DeleteSecurityConfigurationInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "does not exist") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting EMR Security Configuration (%s): %s", d.Id(), err)
	}

	return diags
}
