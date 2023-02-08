package neptune

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceClusterEndpoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterEndpointCreate,
		ReadWithoutTimeout:   resourceClusterEndpointRead,
		UpdateWithoutTimeout: resourceClusterEndpointUpdate,
		DeleteWithoutTimeout: resourceClusterEndpointDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"cluster_endpoint_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validIdentifier,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"READER", "WRITER", "ANY"}, false),
			},
			"static_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"excluded_members": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &neptune.CreateDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Get("cluster_endpoint_identifier").(string)),
		DBClusterIdentifier:         aws.String(d.Get("cluster_identifier").(string)),
		EndpointType:                aws.String(d.Get("endpoint_type").(string)),
	}

	if attr := d.Get("static_members").(*schema.Set); attr.Len() > 0 {
		input.StaticMembers = flex.ExpandStringSet(attr)
	}

	if attr := d.Get("excluded_members").(*schema.Set); attr.Len() > 0 {
		input.ExcludedMembers = flex.ExpandStringSet(attr)
	}

	// Tags are currently only supported in AWS Commercial.
	if len(tags) > 0 && meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateDBClusterEndpointWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster Endpoint: %s", err)
	}

	clusterId := aws.StringValue(out.DBClusterIdentifier)
	endpointId := aws.StringValue(out.DBClusterEndpointIdentifier)
	d.SetId(fmt.Sprintf("%s:%s", clusterId, endpointId))

	_, err = WaitDBClusterEndpointAvailable(ctx, conn, d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Endpoint (%q) to be Available: %s", d.Id(), err)
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := FindEndpointByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[DEBUG] Neptune Cluster Endpoint (%s) not found", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Neptune Cluster Endpoint (%s): %s", d.Id(), err)
	}

	d.Set("cluster_endpoint_identifier", resp.DBClusterEndpointIdentifier)
	d.Set("cluster_identifier", resp.DBClusterIdentifier)
	d.Set("endpoint_type", resp.CustomEndpointType)
	d.Set("endpoint", resp.Endpoint)
	d.Set("excluded_members", flex.FlattenStringSet(resp.ExcludedMembers))
	d.Set("static_members", flex.FlattenStringSet(resp.StaticMembers))

	arn := aws.StringValue(resp.DBClusterEndpointArn)
	d.Set("arn", arn)

	// Tags are currently only supported in AWS Commercial.
	if meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		tags, err := ListTags(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing tags for Neptune Cluster Endpoint (%s): %s", arn, err)
		}

		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
		}
	} else {
		d.Set("tags", nil)
		d.Set("tags_all", nil)
	}

	return diags
}

func resourceClusterEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	if d.HasChangesExcept("tags", "tags_all") {
		req := &neptune.ModifyDBClusterEndpointInput{
			DBClusterEndpointIdentifier: aws.String(d.Get("cluster_endpoint_identifier").(string)),
		}

		if d.HasChange("endpoint_type") {
			req.EndpointType = aws.String(d.Get("endpoint_type").(string))
		}

		if d.HasChange("static_members") {
			req.StaticMembers = flex.ExpandStringSet(d.Get("static_members").(*schema.Set))
		}

		if d.HasChange("excluded_members") {
			req.ExcludedMembers = flex.ExpandStringSet(d.Get("excluded_members").(*schema.Set))
		}

		_, err := conn.ModifyDBClusterEndpointWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Cluster Endpoint (%q): %s", d.Id(), err)
		}

		_, err = WaitDBClusterEndpointAvailable(ctx, conn, d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Endpoint (%q) to be Available: %s", d.Id(), err)
		}
	}

	// Tags are currently only supported in AWS Commercial.
	if d.HasChange("tags_all") && meta.(*conns.AWSClient).Partition == endpoints.AwsPartitionID {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Cluster Endpoint (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return append(diags, resourceClusterEndpointRead(ctx, d, meta)...)
}

func resourceClusterEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	endpointId := d.Get("cluster_endpoint_identifier").(string)
	input := &neptune.DeleteDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(endpointId),
	}

	_, err := conn.DeleteDBClusterEndpointWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) ||
			tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "Neptune Cluster Endpoint cannot be deleted: %s", err)
	}
	_, err = WaitDBClusterEndpointDeleted(ctx, conn, d.Id())
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for Neptune Cluster Endpoint (%q) to be Deleted: %s", d.Id(), err)
	}

	return diags
}
