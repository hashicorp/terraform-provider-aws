package ec2

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceIPAMScope() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceIPAMScopeCreate,
		ReadWithoutTimeout:   ResourceIPAMScopeRead,
		UpdateWithoutTimeout: ResourceIPAMScopeUpdate,
		DeleteWithoutTimeout: ResourceIPAMScopeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(3 * time.Minute),
			Update: schema.DefaultTimeout(3 * time.Minute),
			Delete: schema.DefaultTimeout(3 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipam_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipam_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ipam_scope_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pool_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func ResourceIPAMScopeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateIpamScopeInput{
		ClientToken:       aws.String(resource.UniqueId()),
		IpamId:            aws.String(d.Get("ipam_id").(string)),
		TagSpecifications: tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeIpamScope),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIpamScopeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IPAM Scope: %s", err)
	}

	d.SetId(aws.StringValue(output.IpamScope.IpamScopeId))

	if _, err := WaitIPAMScopeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) create: %s", d.Id(), err)
	}

	return append(diags, ResourceIPAMScopeRead(ctx, d, meta)...)
}

func ResourceIPAMScopeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	scope, err := FindIPAMScopeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IPAM Scope (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Scope (%s): %s", d.Id(), err)
	}

	ipamID := strings.Split(aws.StringValue(scope.IpamArn), "/")[1]
	d.Set("arn", scope.IpamScopeArn)
	d.Set("description", scope.Description)
	d.Set("ipam_arn", scope.IpamArn)
	d.Set("ipam_id", ipamID)
	d.Set("ipam_scope_type", scope.IpamScopeType)
	d.Set("is_default", scope.IsDefault)
	d.Set("pool_count", scope.PoolCount)

	tags := KeyValueTags(scope.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func ResourceIPAMScopeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	if d.HasChange("description") {
		input := &ec2.ModifyIpamScopeInput{
			IpamScopeId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.ModifyIpamScopeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Scope (%s): %s", d.Id(), err)
		}

		if _, err := WaitIPAMScopeUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating IPAM Scope (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, ResourceIPAMScopeRead(ctx, d, meta)...)
}

func ResourceIPAMScopeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	log.Printf("[DEBUG] Deleting IPAM Scope: %s", d.Id())
	_, err := conn.DeleteIpamScopeWithContext(ctx, &ec2.DeleteIpamScopeInput{
		IpamScopeId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidIPAMScopeIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IPAM Scope: (%s): %s", d.Id(), err)
	}

	if _, err := WaitIPAMScopeDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for IPAM Scope (%s) delete: %s", d.Id(), err)
	}

	return diags
}
