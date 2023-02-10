package rds

import (
	"context"
	"log"
	"time"

	rds_sdkv2 "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "Managed by Terraform",
			},
			"family": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
			"parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apply_method": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "immediate",
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				Set: resourceParameterHash,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceClusterParameterGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	groupName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &rds.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(groupName),
		DBParameterGroupFamily:      aws.String(d.Get("family").(string)),
		Description:                 aws.String(d.Get("description").(string)),
		Tags:                        Tags(tags.IgnoreAWS()),
	}

	output, err := conn.CreateDBClusterParameterGroupWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DB Cluster Parameter Group (%s): %s", groupName, err)
	}

	d.SetId(groupName)

	// Set for update
	d.Set("arn", output.DBClusterParameterGroup.DBClusterParameterGroupArn)

	return append(diags, resourceClusterParameterGroupUpdate(ctx, d, meta)...)
}

func resourceClusterParameterGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dbClusterParameterGroup, err := FindDBClusterParameterGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Cluster Parameter Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupArn)
	d.Set("arn", arn)
	d.Set("description", dbClusterParameterGroup.Description)
	d.Set("family", dbClusterParameterGroup.DBParameterGroupFamily)
	d.Set("name", dbClusterParameterGroup.DBClusterParameterGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName)))

	// Only include user customized parameters as there's hundreds of system/default ones
	input := &rds.DescribeDBClusterParametersInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
		Source:                      aws.String("user"),
	}
	var parameters []*rds.Parameter

	err = conn.DescribeDBClusterParametersPagesWithContext(ctx, input, func(page *rds.DescribeDBClusterParametersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Parameters {
			if v != nil {
				parameters = append(parameters, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS Cluster Parameter Group (%s) parameters: %s", d.Id(), err)
	}

	if err := d.Set("parameter", flattenParameters(parameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parameter: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		log.Printf("[WARN] listing tags for RDS DB Cluster Parameter Group (%s): %s", d.Id(), err)
	} else {
		tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

		//lintignore:AWSR002
		if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
		}

		if err := d.Set("tags_all", tags.Map()); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
		}
	}

	return diags
}

func resourceClusterParameterGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn()

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

		// Expand the "parameter" set to aws-sdk-go compat []rds.Parameter
		parameters := expandParameters(ns.Difference(os).List())
		if len(parameters) > 0 {
			// We can only modify 20 parameters at a time, so walk them until
			// we've got them all.
			for parameters != nil {
				var paramsToModify []*rds.Parameter
				if len(parameters) <= clusterParameterGroupMaxParamsBulkEdit {
					paramsToModify, parameters = parameters[:], nil
				} else {
					paramsToModify, parameters = parameters[:clusterParameterGroupMaxParamsBulkEdit], parameters[clusterParameterGroupMaxParamsBulkEdit:]
				}

				input := &rds.ModifyDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(d.Id()),
					Parameters:                  paramsToModify,
				}

				_, err := conn.ModifyDBClusterParameterGroupWithContext(ctx, input)

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "modifying DB Cluster Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}

		toRemove := map[string]*rds.Parameter{}

		for _, p := range expandParameters(os.List()) {
			if p.ParameterName != nil {
				toRemove[*p.ParameterName] = p
			}
		}

		for _, p := range expandParameters(ns.List()) {
			if p.ParameterName != nil {
				delete(toRemove, *p.ParameterName)
			}
		}

		// Reset parameters that have been removed
		var resetParameters []*rds.Parameter
		for _, v := range toRemove {
			resetParameters = append(resetParameters, v)
		}
		if len(resetParameters) > 0 {
			for resetParameters != nil {
				var paramsToReset []*rds.Parameter
				if len(resetParameters) <= clusterParameterGroupMaxParamsBulkEdit {
					paramsToReset, resetParameters = resetParameters[:], nil
				} else {
					paramsToReset, resetParameters = resetParameters[:clusterParameterGroupMaxParamsBulkEdit], resetParameters[clusterParameterGroupMaxParamsBulkEdit:]
				}

				input := &rds.ResetDBClusterParameterGroupInput{
					DBClusterParameterGroupName: aws.String(d.Id()),
					Parameters:                  paramsToReset,
					ResetAllParameters:          aws.Bool(false),
				}

				err := resource.Retry(3*time.Minute, func() *resource.RetryError {
					_, err := conn.ResetDBClusterParameterGroupWithContext(ctx, input)
					if err != nil {
						if tfawserr.ErrMessageContains(err, "InvalidDBParameterGroupState", "has pending changes") {
							return resource.RetryableError(err)
						}
						return resource.NonRetryableError(err)
					}
					return nil
				})

				if tfresource.TimedOut(err) {
					_, err = conn.ResetDBClusterParameterGroupWithContext(ctx, input)
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "resetting DB Cluster Parameter Group (%s): %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating RDS Cluster Parameter Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceClusterParameterGroupRead(ctx, d, meta)...)
}

func resourceClusterParameterGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSClient()

	input := &rds_sdkv2.DeleteDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting RDS DB Cluster Parameter Group: %s", d.Id())
	err := resource.RetryContext(ctx, 3*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBClusterParameterGroup(ctx, input)
		if errs.IsA[*types.DBParameterGroupNotFoundFault](err) {
			return nil
		} else if errs.IsA[*types.InvalidDBParameterGroupStateFault](err) {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteDBClusterParameterGroup(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS Cluster Parameter Group (%s): %s", d.Id(), err)
	}

	return diags
}

func FindDBClusterParameterGroupByName(ctx context.Context, conn *rds.RDS, name string) (*rds.DBClusterParameterGroup, error) {
	input := &rds.DescribeDBClusterParameterGroupsInput{
		DBClusterParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeDBClusterParameterGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBParameterGroupNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DBClusterParameterGroups) == 0 || output.DBClusterParameterGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	dbClusterParameterGroup := output.DBClusterParameterGroups[0]

	// Eventual consistency check.
	if aws.StringValue(dbClusterParameterGroup.DBClusterParameterGroupName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return dbClusterParameterGroup, nil
}
