package neptune

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const clusterParameterGroupMaxParamsBulkEdit = 20

func ResourceClusterParameterGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClusterParameterGroupCreate,
		ReadWithoutTimeout:   resourceClusterParameterGroupRead,
		UpdateWithoutTimeout: resourceClusterParameterGroupUpdate,
		DeleteWithoutTimeout: resourceClusterParameterGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validParamGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validParamGroupNamePrefix,
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  neptune.ApplyMethodPendingReboot,
							ValidateFunc: validation.StringInSlice([]string{
								neptune.ApplyMethodImmediate,
								neptune.ApplyMethodPendingReboot,
							}, false),
						},
					},
				},
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := neptune.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	_, err := conn.CreateDBClusterParameterGroupWithContext(ctx, &createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Neptune Cluster Parameter Group (%s): %s", groupName, err)
	}

	d.SetId(aws.StringValue(createOpts.DBClusterParameterGroupName))

	if v, ok := d.GetOk("parameter"); ok && v.(*schema.Set).Len() > 0 {
		err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), expandParameters(v.(*schema.Set).List()))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	describeOpts := neptune.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, &describeOpts)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	if describeResp == nil || len(describeResp.DBClusterParameterGroups) == 0 {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s): not found", d.Id())
		}

		log.Printf("[WARN] Neptune Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("name", describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupName)
	d.Set("family", describeResp.DBClusterParameterGroups[0].DBParameterGroupFamily)
	d.Set("description", describeResp.DBClusterParameterGroups[0].Description)
	arn := aws.StringValue(describeResp.DBClusterParameterGroups[0].DBClusterParameterGroupArn)
	d.Set("arn", arn)

	// Only include user customized parameters as there's hundreds of system/default ones
	describeParametersOpts := neptune.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}

	describeParametersResp, err := conn.DescribeDBClusterParametersWithContext(ctx, &describeParametersOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(describeParametersResp.Parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting neptune parameter: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Get("arn").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	if d.HasChange("parameter") {
		o, n := d.GetChange("parameter")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		parameters := expandParameters(ns.Difference(os).List())

		if len(parameters) > 0 {
			err := modifyClusterParameterGroupParameters(ctx, conn, d.Id(), parameters)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Neptune Cluster Parameter Group (%s) parameter: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Neptune Cluster Parameter Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NeptuneConn()

	input := neptune.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBClusterParameterGroupWithContext(ctx, &input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBParameterGroupNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Neptune Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func modifyClusterParameterGroupParameters(ctx context.Context, conn *neptune.Neptune, name string, parameters []*neptune.Parameter) error {
	// We can only modify 20 parameters at a time, so walk them until
	// we've got them all.
	for parameters != nil {
		var paramsToModify []*neptune.Parameter
		if len(parameters) <= clusterParameterGroupMaxParamsBulkEdit {
			paramsToModify, parameters = parameters[:], nil
		} else {
			paramsToModify, parameters = parameters[:clusterParameterGroupMaxParamsBulkEdit], parameters[clusterParameterGroupMaxParamsBulkEdit:]
		}

		modifyOpts := neptune.ModifyDBClusterParameterGroupInput{
			DBClusterParameterGroupName: aws.String(name),
			Parameters:                  paramsToModify,
		}

		_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, &modifyOpts)
		if err != nil {
			return err
		}
	}

	return nil
}
