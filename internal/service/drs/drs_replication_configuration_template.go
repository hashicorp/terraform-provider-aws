package drs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/drs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/drs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Replication Configuration Template")
func newResourceReplicationConfigurationTemplate(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceReplicationConfigurationTemplate{}, nil
}

const (
	ResNameReplicationConfigurationTemplate = "ReplicationConfigurationTemplate"
)

type resourceReplicationConfigurationTemplate struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceReplicationConfigurationTemplate) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_drs_replication_configuration_template"
}

func (r *resourceReplicationConfigurationTemplate) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				Computed: true,
			},
			"associate_default_security_group": schema.BoolAttribute{
				Required: true,
			},
			"bandwidth_throttling": schema.Int64Attribute{
				Required: true,
			},
			"create_public_ip": schema.BoolAttribute{
				Required: true,
			},
			"data_plane_routing": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationDataPlaneRouting](),
			},
			"default_large_staging_disk_type": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationDefaultLargeStagingDiskType](),
			},
			"ebs_encryption": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.ReplicationConfigurationEbsEncryption](),
			},
			"ebs_encryption_key_arn": schema.StringAttribute{
				Required: true,
			},
			"replication_server_instance_type": schema.StringAttribute{
				Required: true,
			},
			"replication_servers_security_groups_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"staging_area_subnet_id": schema.StringAttribute{
				Required: true,
			},

			"staging_area_tags": tftags.TagsAttribute(),
			names.AttrTags:      tftags.TagsAttribute(),
			names.AttrTagsAll:   tftags.TagsAttributeComputedOnly(),

			"use_dedicated_replication_server": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"pit_policy": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[pitPolicy](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Optional: true,
						},
						"interval": schema.Int64Attribute{
							Optional: true,
						},
						"retention_duration": schema.Int64Attribute{
							Required: true,
						},
						"rule_id": schema.Int64Attribute{
							Optional: true,
						},
						"units": schema.StringAttribute{
							Required: true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
			},
		},
	}
}

func (r *resourceReplicationConfigurationTemplate) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceReplicationConfigurationTemplateData
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().DRSClient(ctx)

	input := &drs.CreateReplicationConfigurationTemplateInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTags(ctx)

	log.Printf("[DEBUG] Creating DRS Replication Configuration Template: %s", input)
	output, err := conn.CreateReplicationConfigurationTemplate(input)

	if err != nil {
		return fmt.Errorf("error creating DRS Replication Configuration Template: %w", err)
	}

	d.SetId(aws.ToString(output.ReplicationConfigurationTemplateID))

	return append(diags, resourceReplicationConfigurationTemplateRead(d, meta)...)
}

func resourceReplicationConfigurationTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSClient(ctx)
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	replicationConfigurationTemplate, err := FindReplicationConfigurationTemplateByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Replication Configuration Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Replication Configuration Template (%s): %w", d.Id(), err)
	}

	d.Set("arn", replicationConfigurationTemplate.Arn)
	d.Set("associate_default_security_group", replicationConfigurationTemplate.AssociateDefaultSecurityGroup)
	d.Set("bandwidth_throttling", replicationConfigurationTemplate.BandwidthThrottling)
	d.Set("create_public_ip", replicationConfigurationTemplate.CreatePublicIP)
	d.Set("data_plane_routing", replicationConfigurationTemplate.DataPlaneRouting)
	d.Set("default_large_staging_disk_type", replicationConfigurationTemplate.DefaultLargeStagingDiskType)
	d.Set("ebs_encryption", replicationConfigurationTemplate.EbsEncryption)
	d.Set("ebs_encryption_key_arn", replicationConfigurationTemplate.EbsEncryptionKeyArn)
	d.Set("pit_policy", replicationConfigurationTemplate.PitPolicy)
	d.Set("replication_server_instance_type", replicationConfigurationTemplate.ReplicationServerInstanceType)
	d.Set("replication_servers_security_groups_ids", replicationConfigurationTemplate.ReplicationServersSecurityGroupsIDs)
	d.Set("staging_area_subnet_id", replicationConfigurationTemplate.StagingAreaSubnetId)
	d.Set("use_dedicated_replication_server", replicationConfigurationTemplate.UseDedicatedReplicationServer)

	tags := KeyValueTags(replicationConfigurationTemplate.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("staging_area_tags", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceReplicationConfigurationTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSClient(ctx)

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	staging_tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("staging_area_tags").(map[string]interface{})))

	input := &drs.UpdateReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: aws.String(d.Id()),
	}

	if d.HasChange("associate_default_security_group") {
		input.AssociateDefaultSecurityGroup = aws.Bool(d.Get("associate_default_security_group").(bool))
	}

	if d.HasChange("bandwidth_throttling") {
		input.BandwidthThrottling = aws.Int64(d.Get("bandwidth_throttling").(int64))
	}

	if d.HasChange("create_public_ip") {
		input.CreatePublicIP = aws.Bool(d.Get("create_public_ip").(bool))
	}

	if d.HasChange("data_plane_routing") {
		input.DataPlaneRouting = aws.String(d.Get("data_plane_routing").(string))
	}

	if d.HasChange("default_large_staging_disk_type") {
		input.DefaultLargeStagingDiskType = aws.String(d.Get("default_large_staging_disk_type").(string))
	}

	if d.HasChange("ebs_encryption") {
		input.EbsEncryption = aws.String(d.Get("ebs_encryption").(string))
	}

	if d.HasChange("ebs_encryption_key_arn") {
		input.EbsEncryptionKeyArn = aws.String(d.Get("ebs_encryption_key_arn").(string))
	}

	if d.HasChange("pit_policy") {
		input.PitPolicy = expandPitPolicy(d.Get("pit_policy").(*schema.Set).List())
	}

	if d.HasChange("replication_server_instance_type") {
		input.ReplicationServerInstanceType = aws.String(d.Get("replication_server_instance_type").(string))
	}

	if d.HasChange("replication_servers_security_groups_ids") {
		input.ReplicationServersSecurityGroupsIDs = flex.ExpandStringList(d.Get("replication_servers_security_groups_ids").([]interface{}))
	}

	if d.HasChange("staging_area_subnet_id") {
		input.StagingAreaSubnetId = aws.String(d.Get("staging_area_subnet_id").(string))
	}

	if d.HasChange("staging_area_tags") {
		input.StagingAreaTags = Tags(staging_tags.IgnoreAWS())
	}

	if d.HasChange("use_dedicated_replication_server") {
		input.UseDedicatedReplicationServer = aws.Bool(d.Get("use_dedicated_replication_server").(bool))
	}

	_, err := conn.UpdateReplicationConfigurationTemplate(input)

	if err != nil {
		return fmt.Errorf("error updating replication configuration template (%s): : %w", d.Id(), err)
	}

	return resourceReplicationConfigurationTemplateRead(d, meta)

}

func resourceReplicationConfigurationTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DRSClient(ctx)

	log.Printf("[DEBUG] Deleting Replication Configuration Template: %s", d.Id())
	_, err := conn.DeleteReplicationConfigurationTemplate(&drs.DeleteReplicationConfigurationTemplateInput{
		ReplicationConfigurationTemplateID: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Replication Configuration Tempalte (%s): %w", d.Id(), err)
	}

	return nil
}

func expandPitPolicy(p []interface{}) []*awstypes.PITPolicyRule {
	pitPolicyRules := make([]*awstypes.PITPolicyRule, len(p))
	for i, pitPolicy := range p {
		pitPol := pitPolicy.(map[string]interface{})
		pitPolicyRules[i] = &awstypes.PITPolicyRule{
			Enabled:           aws.Bool(pitPol["enabled"].(bool)),
			Interval:          aws.Int32(pitPol["interval"].(int32)),
			RetentionDuration: aws.Int32(pitPol["retention_duration"].(int32)),
			RuleID:            pitPol["rule_id"].(int64),
			Units:             pitPol["units"].(string),
		}
	}
	return pitPolicyRules
}

func Tags(tags tftags.KeyValueTags) map[string]*string {
	return aws.StringMap(tags.Map())
}

func KeyValueTags(tags map[string]*string) tftags.KeyValueTags {
	return tftags.New(tags)
}

type resourceReplicationConfigurationTemplateData struct {
	AssociateDefaultSecurityGroup       types.Bool                                                                       `tfsdk:"associate_default_security_group"`
	BandwidthThrottling                 types.Int                                                                        `tfsdk:"bandwidth_throttling"`
	CreatePublicIP                      types.Bool                                                                       `tfsdk:"create_public_ip"`
	DataPlaneRouting                    fwtypes.StringEnum[awstypes.ReplicationConfigurationDataPlaneRouting]            `tfsdk:"data_plane_routing"`
	DefaultLargeStagingDiskType         fwtypes.StringEnum[awstypes.ReplicationConfigurationDefaultLargeStagingDiskType] `tfsdk:"default_large_staging_disk_type"`
	EbsEncryption                       fwtypes.StringEnum[awstypes.ReplicationConfigurationEbsEncryption]               `tfsdk:"ebs_encryption"`
	EbsEncryptionKeyARN                 types.String                                                                     `tfsdk:"ebs_encryption_key_arn"`
	PitPolicy                           fwtypes.ListNestedObjectValueOf[pitPolicy]                                       `tfsdk:"pit_policy"`
	ReplicationServerInstanceType       types.String                                                                     `tfsdk:"replication_server_instance_type"`
	ReplicationServersSecurityGroupsIDs types.List                                                                       `tfsdk:"replication_servers_security_groups_ids"`
	StagingAreaSubnetID                 types.String                                                                     `tfsdk:"staging_area_subnet_id"`
	UseDedicatedReplicationServer       types.Bool                                                                       `tfsdk:"use_dedicated_replication_server"`
	StagingAreaTags                     tftags.KeyValueTags                                                              `tfsdk:"staging_area_tags"`
	Tags                                tftags.KeyValueTags                                                              `tfsdk:"tags"`
}

type pitPolicy struct {
	Enabled           types.Bool   `tfsdk:"enabled"`
	Interval          types.Int    `tfsdk:"interval"`
	RetentionDuration types.Int    `tfsdk:"retention_duration"`
	RuleID            types.Int    `tfsdk:"rule_id"`
	Units             types.String `tfsdk:"units"`
}
