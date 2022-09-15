package medialive

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/medialive"
	"github.com/aws/aws-sdk-go-v2/service/medialive/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceChannelCreate,
		ReadWithoutTimeout:   resourceChannelRead,
		UpdateWithoutTimeout: resourceChannelUpdate,
		DeleteWithoutTimeout: resourceChannelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cdi_input_specification": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resolution": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.CdiInputResolution](),
						},
					},
				},
			},
			"channel_class": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ChannelClass](),
			},
			"destinations": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"media_package_settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"channel_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"multiplex_settings": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"multiplex_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"program_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"settings": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password_param": {
										Type:     schema.TypeString,
										Required: true,
									},
									"stream_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"url": {
										Type:     schema.TypeString,
										Required: true,
									},
									"username": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"encoder_settings": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"audio_description": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audio_selector_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"audio_normalization_settings": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"algorithm": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[types.AudioNormalizationAlgorithm](),
												},
												"algorithm_control": {
													Type:             schema.TypeString,
													Optional:         true,
													Computed:         true,
													ValidateDiagFunc: enum.Validate[types.AudioNormalizationAlgorithmControl](),
												},
												"target_lkfs": {
													Type:     schema.TypeFloat,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
									"audio_type_control": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[types.AudioDescriptionAudioTypeControl](),
									},
									"audio_watermark_settings": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"nielsen_watermarks_settings": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"nielsen_cbet_settings": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"cbet_check_digit_string": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		"cbet_stepaside": {
																			Type:             schema.TypeString,
																			Required:         true,
																			ValidateDiagFunc: enum.Validate[types.NielsenWatermarksCbetStepaside](),
																		},
																		"csid": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
															},
															"nielsen_distribution_type": {
																Type:             schema.TypeString,
																Optional:         true,
																Computed:         true,
																ValidateDiagFunc: enum.Validate[types.NielsenWatermarksDistributionTypes](),
															},
															"nielsen_naes_ii_nw_settings": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"check_digit_string": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																		"sid": {
																			Type:     schema.TypeFloat,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									"codec_settings": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
								},
							},
						},
					},
				},
			},
			"input_specification": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"codec": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.InputCodec](),
						},
						"maximum_bitrate": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.InputMaximumBitrate](),
						},
						"input_resolution": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.InputResolution](),
						},
					},
				},
			},
			"log_level": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.LogLevel](),
			},
			"maintenance": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maintenance_day": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.MaintenanceDay](),
						},
						"maintenance_start_time": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
			},
			"vpc": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"public_address_allocation_ids": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"security_group_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 5,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

const (
	ResNameChannel = "Channel"
)

func resourceChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	in := &medialive.CreateChannelInput{
		Name:      aws.String(d.Get("name").(string)),
		RequestId: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("max_size"); ok {
		in.MaxSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("complex_argument"); ok && len(v.([]interface{})) > 0 {
		in.ComplexArguments = expandComplexArguments(v.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateChannel(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), err)
	}

	if out == nil || out.Channel == nil {
		return create.DiagError(names.MediaLive, create.ErrActionCreating, ResNameChannel, d.Get("name").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Channel.Id))

	if _, err := waitChannelCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForCreation, ResNameChannel, d.Id(), err)
	}

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	out, err := FindChannelByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaLive Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	p, err = structure.NormalizeJsonString(p)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	d.Set("policy", p)

	tags, err := ListTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionReading, ResNameChannel, d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionSetting, ResNameChannel, d.Id(), err)
	}

	return nil
}

func resourceChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	update := false

	in := &medialive.UpdateChannelInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = aws.String(d.Get("an_argument").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating MediaLive Channel (%s): %#v", d.Id(), in)
	out, err := conn.UpdateChannel(ctx, in)
	if err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionUpdating, ResNameChannel, d.Id(), err)
	}

	if _, err := waitChannelUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForUpdate, ResNameChannel, d.Id(), err)
	}

	return resourceChannelRead(ctx, d, meta)
}

func resourceChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).MediaLiveConn

	log.Printf("[INFO] Deleting MediaLive Channel %s", d.Id())

	_, err := conn.DeleteChannel(ctx, &medialive.DeleteChannelInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil
		}

		return create.DiagError(names.MediaLive, create.ErrActionDeleting, ResNameEndpoint, d.Id(), err)
	}

	if _, err := waitChannelDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.DiagError(names.MediaLive, create.ErrActionWaitingForDeletion, ResNameChannel, d.Id(), err)
	}

	return nil
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitChannelCreated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.ChannelStateCreating),
		Target:                    enum.Slice(types.ChannelStateIdle),
		Refresh:                   statusChannel(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelUpdated(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   enum.Slice(types.ChannelStateUpdating),
		Target:                    enum.Slice(types.ChannelStateIdle),
		Refresh:                   statusChannel(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func waitChannelDeleted(ctx context.Context, conn *medialive.Client, id string, timeout time.Duration) (*medialive.DescribeChannelOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: enum.Slice(types.ChannelStateDeleting),
		Target:  enum.Slice(types.ChannelStateDeleted),
		Refresh: statusChannel(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*medialive.DescribeChannelOutput); ok {
		return out, err
	}

	return nil, err
}

func statusChannel(ctx context.Context, conn *medialive.Client, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindChannelByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindChannelByID(ctx context.Context, conn *medialive.Client, id string) (*medialive.DescribeChannelOutput, error) {
	in := &medialive.DescribeChannelInput{
		ChannelId: aws.String(id),
	}
	out, err := conn.DescribeChannel(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &resource.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

func flattenComplexArgument(apiObject *medialive.ComplexArgument) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}

func flattenComplexArguments(apiObjects []*medialive.ComplexArgument) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

func expandComplexArgument(tfMap map[string]interface{}) *medialive.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &medialive.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

func expandComplexArguments(tfList []interface{}) []*medialive.ComplexArgument {

	var s []*medialive.ComplexArgument

	// s := make([]*medialive.ComplexArgument, 0)

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandComplexArgument(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}
