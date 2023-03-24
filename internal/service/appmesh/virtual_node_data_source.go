package appmesh

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appmesh_virtual_node")
func DataSourceVirtualNode() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceVirtualNodeRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mesh_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"spec":         {},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceVirtualNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	/*
		conn := meta.(*conns.AWSClient).AppMeshConn()
		ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

		req := &appmesh.DescribeVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			VirtualNodeName: aws.String(d.Get("name").(string)),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		var resp *appmesh.DescribeVirtualNodeOutput

		err := resource.Retry(propagationTimeout, func() *resource.RetryError {
			var err error

			resp, err = conn.DescribeVirtualNode(req)

			if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
				return resource.RetryableError(err)
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			resp, err = conn.DescribeVirtualNode(req)
		}

		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		if err != nil {
			return fmt.Errorf("reading App Mesh Virtual Node: %w", err)
		}

		if resp == nil || resp.VirtualNode == nil {
			return fmt.Errorf("reading App Mesh Virtual Node: empty response")
		}

		if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
			if d.IsNewResource() {
				return fmt.Errorf("reading App Mesh Virtual Node: %s after creation", aws.StringValue(resp.VirtualNode.Status.Status))
			}

			log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		arn := aws.StringValue(resp.VirtualNode.Metadata.Arn)
		d.Set("name", resp.VirtualNode.VirtualNodeName)
		d.Set("mesh_name", resp.VirtualNode.MeshName)
		d.Set("mesh_owner", resp.VirtualNode.Metadata.MeshOwner)
		d.Set("arn", arn)
		d.Set("created_date", resp.VirtualNode.Metadata.CreatedAt.Format(time.RFC3339))
		d.Set("last_updated_date", resp.VirtualNode.Metadata.LastUpdatedAt.Format(time.RFC3339))
		d.Set("resource_owner", resp.VirtualNode.Metadata.ResourceOwner)
		err = d.Set("spec", flattenVirtualNodeSpec(resp.VirtualNode.Spec))
		if err != nil {
			return fmt.Errorf("setting spec: %w", err)
		}

		tags, err := ListTags(conn, arn)

		if err != nil {
			return fmt.Errorf("listing tags for App Mesh virtual node (%s): %w", arn, err)
		}

		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return fmt.Errorf("setting tags: %w", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return fmt.Errorf("setting tags_all: %w", err)
		}
	*/
	return diags
}
