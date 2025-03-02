// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mediapackage

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/mediapackage/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage"
	"github.com/aws/aws-sdk-go-v2/service/mediapackage/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_mediapackage_origin_endpoint", name="Origin Endpoint")
func ResourceOriginEndpoint() *schema.Resource {
	return &schema.Resource{
		// TIP: ==== ASSIGN CRUD FUNCTIONS ====
		// These 4 functions handle CRUD responsibilities below.
		CreateWithoutTimeout: resourceOriginEndpointCreate,
		ReadWithoutTimeout:   resourceOriginEndpointRead,
		UpdateWithoutTimeout: resourceOriginEndpointUpdate,
		DeleteWithoutTimeout: resourceOriginEndpointDelete,

		// TIP: ==== TERRAFORM IMPORTING ====
		// If Read can get all the information it needs from the Identifier
		// (i.e., d.Id()), you can use the Passthrough importer. Otherwise,
		// you'll need a custom import function.
		//
		// See more:
		// https://hashicorp.github.io/terraform-provider-aws/add-import-support/
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#implicit-state-passthrough
		// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#virtual-attributes
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		// TIP: ==== CONFIGURABLE TIMEOUTS ====
		// Users can configure timeout lengths but you need to use the times they
		// provide. Access the timeout they configure (or the defaults) using,
		// e.g., d.Timeout(schema.TimeoutCreate) (see below). The times here are
		// the defaults if they don't configure timeouts.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		// TIP: ==== SCHEMA ====
		// In the schema, add each of the attributes in snake case (e.g.,
		// delete_automated_backups).
		//
		// Formatting rules:
		// * Alphabetize attributes to make them easier to find.
		// * Do not add a blank line between attributes.
		//
		// Attribute basics:
		// * If a user can provide a value ("configure a value") for an
		//   attribute (e.g., instances = 5), we call the attribute an
		//   "argument."
		// * You change the way users interact with attributes using:
		//     - Required
		//     - Optional
		//     - Computed
		// * There are only four valid combinations:
		//
		// 1. Required only - the user must provide a value
		// Required: true,
		//
		// 2. Optional only - the user can configure or omit a value; do not
		//    use Default or DefaultFunc
		// Optional: true,
		//
		// 3. Computed only - the provider can provide a value but the user
		//    cannot, i.e., read-only
		// Computed: true,
		//
		// 4. Optional AND Computed - the provider or user can provide a value;
		//    use this combination if you are using Default or DefaultFunc
		// Optional: true,
		// Computed: true,
		//
		// You will typically find arguments in the input struct
		// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
		// they are only in the input struct (e.g., ModifyDBInstanceInput) for
		// the modify operation.
		//
		// For more about schema options, visit
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Schema

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"channel_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchema(),
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authorization": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cdn_identifier_secret": {
							Type:     schema.TypeString,
							Required: true,
						},
						"secrets_role_arn": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cmaf_package": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"system_ids": {
													Type:     schema.TypeList,
													Required: true,
												},
												"url": {
													Type:     schema.TypeString,
													Required: true,
												},
												"certificate_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_contract_configuration": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"preset_speke20_audio": {
																Type:     schema.TypeString,
																Required: true,
															},
															"preset_speke20_video": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"constant_initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"encryption_method": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"hls_manifests": {
							Type: schema.TypeList,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrID: {
										Type:     schema.TypeString,
										Required: true,
									},
									"ad_markers": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"ad_triggers": {
										Type:     schema.TypeList,
										Optional: true,
									},
									"ads_on_delivery_restrictions": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"include_iframe_only_stream": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"manifest_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"playlist_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"playlist_window_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"program_date_time_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"stream_selection": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"min_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"stream_order": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"dash_package": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_triggers": {
							Type:     schema.TypeList,
							Optional: true,
						},
						"ads_on_delivery_restrictions": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"system_ids": {
													Type:     schema.TypeList,
													Required: true,
												},
												"url": {
													Type:     schema.TypeString,
													Required: true,
												},
												"certificate_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_contract_configuration": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"preset_speke20_audio": {
																Type:     schema.TypeString,
																Required: true,
															},
															"preset_speke20_video": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"included_iframe_only_stream": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"manifest_layout": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"manifest_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_buffer_time_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"min_update_period_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"period_triggers": {
							Type:     schema.TypeList,
							Optional: true,
						},
						"profile": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_template_format": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"stream_selection": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"min_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"stream_order": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"suggested_presentation_delay_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"utc_timing": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"utc_timing_uri": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hls_package": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ad_markers": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ad_triggers": {
							Type:     schema.TypeList,
							Optional: true,
						},
						"ads_on_delivery_restrictions": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"encryption": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"system_ids": {
													Type:     schema.TypeList,
													Required: true,
												},
												"url": {
													Type:     schema.TypeString,
													Required: true,
												},
												"certificate_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_contract_configuration": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"preset_speke20_audio": {
																Type:     schema.TypeString,
																Required: true,
															},
															"preset_speke20_video": {
																Type:     schema.TypeString,
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"constant_initialization_vector": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"encryption_method": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"key_rotation_interval_seconds": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"repeat_ext_x_key": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"include_dvb_subtitles": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"include_iframe_only_stream": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"playlist_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"playlist_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"program_date_time_interval_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stream_selection": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"min_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"stream_order": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"use_audio_rendition_group": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"include_iframe_only_stream": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"manifest_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mss_package": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"speke_key_provider": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"role_arn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"system_ids": {
													Type:     schema.TypeList,
													Required: true,
												},
												"url": {
													Type:     schema.TypeString,
													Required: true,
												},
												"certificate_arn": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"encryption_contract_configuration": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"preset_speke20_audio": {
																Type:     schema.TypeString,
																Required: true,
															},
															"preset_speke20_video": {
																Type:     schema.TypeString,
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
						"manifest_window_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"segment_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"stream_selection": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"min_video_bits_per_second": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"stream_order": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"origination": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"start_over_window_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"time_delay_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"whitelist": {
				Type:     schema.TypeList,
				Optional: true,
			},
		},
	}
}

const (
	ResNameOriginEndpoint = "Origin Endpoint"
)

func resourceOriginEndpointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE CREATE ====
	// Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a create input structure
	// 3. Call the AWS create/put function
	// 4. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work. At a minimum, set the
	//    resource ID. E.g., d.SetId(<Identifier, such as AWS ID or ARN>)
	// 5. Use a waiter to wait for create to complete
	// 6. Call the Read function in the Create return

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	// TIP: -- 2. Populate a create input structure
	in := &mediapackage.CreateOriginEndpointInput{
		ChannelId: aws.String(d.Get("channel_id").(string)),
		Id:        aws.String(d.Get("id").(string)),
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		// OriginEndpointName: aws.String(d.Get("name").(string)),
		// OriginEndpointType: aws.String(d.Get("type").(string)),

		// TIP: Not all resources support tags and tags don't always make sense. If
		// your resource doesn't need tags, you can remove the tags lines here and
		// below. Many resources do include tags so this a reminder to include them
		// where possible.
		Tags: getTagsIn(ctx),
	}
	if v, ok := d.GetOk("authorization"); ok {
		in.Authorization = expandAuthorization(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("cmaf_package"); ok && len(v.(*schema.Set).List()) > 0 {
		in.CmafPackage = expandCmafPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("dash_package"); ok && len(v.(*schema.Set).List()) > 0 {
		in.DashPackage = expandDashPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hls_package"); ok {
		in.HlsPackage = expandHlsPackage(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("mss_package"); ok {
		in.MssPackage = expandMssPackage(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("manifest_name"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("origination"); ok {
		in.Origination = types.Origination(v.(string))
	}

	if v, ok := d.GetOk("start_over_window_seconds"); ok {
		in.StartoverWindowSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("time_delay_seconds"); ok {
		in.TimeDelaySeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("whitelist"); ok {
		in.Whitelist = v.([]string)
	}

	// TIP: -- 3. Call the AWS create function
	out, err := conn.CreateOriginEndpoint(ctx, in)
	if err != nil {
		// TIP: Since d.SetId() has not been called yet, you cannot use d.Id()
		// in error messages at this point.
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get("name").(string), err)
	}

	if out == nil || out.Origination == "" {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get("name").(string), errors.New("empty output"))
	}

	// TIP: -- 4. Set the minimum arguments and/or attributes for the Read function to
	// work.
	d.SetId(aws.ToString(out.Id))

	// TIP: -- 5. Use a waiter to wait for create to complete
	if _, err := waitOriginEndpointCreated(ctx, conn, *out.ChannelId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionWaitingForCreation, ResNameOriginEndpoint, d.Id(), err)
	}

	// TIP: -- 6. Call the Read function in the Create return
	return append(diags, resourceOriginEndpointRead(ctx, d, meta)...)
}

func resourceOriginEndpointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE READ ====
	// Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get the resource from AWS
	// 3. Set ID to empty where resource is not new and not found
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return diags

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	// TIP: -- 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	channelID := d.Get("channel_id").(string)
	out, err := findOriginEndpointsByChannelID(ctx, conn, channelID)

	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaPackage OriginEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionReading, ResNameOriginEndpoint, d.Id(), err)
	}

	var ep *types.OriginEndpoint

	for _, o := range out {
		if aws.ToString(o.Id) == d.Id() {
			ep = &o
		}
	}

	if ep == nil {
		log.Printf("[WARN] MediaPackage OriginEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	// TIP: -- 4. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.TypeString, schema.TypeBool,
	// schema.TypeInt, and schema.TypeFloat), a simple Set call (e.g.,
	// d.Set("arn", out.Arn) is sufficient. No error or nil checking is
	// necessary.
	//
	// However, there are some situations where more handling is needed.
	// a. Complex data types (e.g., schema.TypeList, schema.TypeSet)
	// b. Where errorneous diffs occur. For example, a schema.TypeString may be
	//    a JSON. AWS may return the JSON in a slightly different order but it
	//    is equivalent to what is already set. In that case, you may check if
	//    it is equivalent before setting the different JSON.
	d.Set("arn", ep.Arn)

	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#data-handling-and-conversion
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#flatten-functions-for-blocks
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#root-typeset-of-resource-and-aws-list-of-structure

	if err := d.Set("authorization", flattenAuthorization(ep.Authorization)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("cmaf_package", flattenCmafPackage(ep.CmafPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("dash_package", flattenDashPackage(ep.DashPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("description", ep.Description); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("hls_package", flattenHlsPackage(ep.HlsPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("mss_package", flattenMssPackage(ep.MssPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("origination", ep.Origination); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("start_over_window_seconds", ep.StartoverWindowSeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("time_delay_seconds", ep.TimeDelaySeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("whitelist", ep.Whitelist); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	return diags
}

func resourceOriginEndpointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE UPDATE ====
	// Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have ForceNew: true, set
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the main resource function will not have a
	// UpdateWithoutTimeout defined. In the case of c., Update and Create are
	// the same.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a modify input structure and check for changes
	// 3. Call the AWS modify/update function
	// 4. Use a waiter to wait for update to complete
	// 5. Call the Read function in the Update return

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	// TIP: -- 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	update := false

	in := &mediapackage.UpdateOriginEndpointInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.Authorization = expandAuthorization(d.Get("an_argument").(map[string]interface{}))
		update = true
	}

	if d.HasChanges("cmaf_package") {
		in.CmafPackage = expandCmafPackage(d.Get("cmaf_package").(map[string]interface{}))
		update = true
	}

	if d.HasChanges("description") {
		in.Description = aws.String(d.Get("description").(string))
		update = true
	}

	if d.HasChanges("hls_package") {
		in.HlsPackage = expandHlsPackage(d.Get("hls_package").(map[string]interface{}))
		update = true
	}

	if d.HasChanges("mss_package") {
		in.MssPackage = expandMssPackage(d.Get("mss_package").(map[string]interface{}))
		update = true
	}

	if !update {
		// TIP: If update doesn't do anything at all, which is rare, you can
		// return diags. Otherwise, return a read call, as below.
		return diags
	}

	// TIP: -- 3. Call the AWS modify/update function
	log.Printf("[DEBUG] Updating MediaPackage OriginEndpoint (%s): %#v", d.Id(), in)
	out, err := conn.UpdateOriginEndpoint(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionUpdating, ResNameOriginEndpoint, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for update to complete
	if _, err := waitOriginEndpointUpdated(ctx, conn, aws.ToString(out.ChannelId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionWaitingForUpdate, ResNameOriginEndpoint, d.Id(), err)
	}

	// TIP: -- 5. Call the Read function in the Update return
	return append(diags, resourceOriginEndpointRead(ctx, d, meta)...)
}

func resourceOriginEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// TIP: ==== RESOURCE DELETE ====
	// Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a delete input structure
	// 3. Call the AWS delete function
	// 4. Use a waiter to wait for delete to complete
	// 5. Return diags

	// TIP: -- 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	// TIP: -- 2. Populate a delete input structure
	log.Printf("[INFO] Deleting MediaPackage OriginEndpoint %s", d.Id())

	// TIP: -- 3. Call the AWS delete function
	_, err := conn.DeleteOriginEndpoint(ctx, &mediapackage.DeleteOriginEndpointInput{
		Id: aws.String(d.Id()),
	})

	// TIP: On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionDeleting, ResNameOriginEndpoint, d.Id(), err)
	}

	// TIP: -- 4. Use a waiter to wait for delete to complete
	if _, err := waitOriginEndpointDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionWaitingForDeletion, ResNameOriginEndpoint, d.Id(), err)
	}

	// TIP: -- 5. Return diags
	return diags
}

// TIP: ==== STATUS CONSTANTS ====
// Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: ==== WAITERS ====
// Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitOriginEndpointCreated(ctx context.Context, conn *mediapackage.Client, channelID string, timeout time.Duration) (*mediapackage.OriginEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusOriginEndpoint(ctx, conn, channelID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediapackage.OriginEndpoint); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitOriginEndpointUpdated(ctx context.Context, conn *mediapackage.Client, channelID string, timeout time.Duration) (*mediapackage.OriginEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusOriginEndpoint(ctx, conn, channelID),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediapackage.OriginEndpoint); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitOriginEndpointDeleted(ctx context.Context, conn *mediapackage.Client, id string, timeout time.Duration) (*mediapackage.OriginEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusOriginEndpoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*mediapackage.OriginEndpoint); ok {
		return out, err
	}

	return nil, err
}

// TIP: ==== STATUS ====
// The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Waiters consume the values returned by status functions. Design status so
// that it can be reused by a create, update, and delete waiter, if possible.
func statusOriginEndpoint(ctx context.Context, conn *mediapackage.Client, channelID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findOriginEndpointsByChannelID(ctx, conn, channelID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString(out.Status), nil
	}
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findOriginEndpointsByChannelID(ctx context.Context, conn *mediapackage.Client, channelID string) ([]types.OriginEndpoint, error) {
	in := &mediapackage.ListOriginEndpointsInput{
		ChannelId: aws.String(channelID),
	}
	out, err := conn.ListOriginEndpoints(ctx, in)
	if err != nil {
		return nil, err
	}
	if len(out.OriginEndpoints) == 0 {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	return out.OriginEndpoints, nil
}

func expandAuthorization(tfMap map[string]interface{}) *types.Authorization {
	if tfMap == nil {
		return nil
	}

	a := &types.Authorization{}

	if v, ok := tfMap["cdn_identifier_secret"].(string); ok && v != "" {
		a.CdnIdentifierSecret = aws.String(v)
	}

	if v, ok := tfMap["secrets_role_arn"].(string); ok && v != "" {
		a.SecretsRoleArn = aws.String(v)
	}

	return a
}

func flattenAuthorization(apiObject *types.Authorization) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CdnIdentifierSecret; v != nil {
		m["cdn_identifier_secret"] = aws.ToString(v)
	}

	if v := apiObject.SecretsRoleArn; v != nil {
		m["secrets_role_arn"] = aws.ToString(v)
	}

	return m
}

func expandHlsPackage(tfMap map[string]interface{}) *types.HlsPackage {
	if tfMap == nil {
		return nil
	}

	h := &types.HlsPackage{}

	if v, ok := tfMap["ad_markers"].(string); ok && v != "" {
		h.AdMarkers = types.AdMarkers(v)
	}

	if v, ok := tfMap["ad_triggers"].([]interface{}); ok && len(v) > 0 {
		h.AdTriggers = expandAdTriggers(v)
	}

	if v, ok := tfMap["ads_on_delivery_restrictions"].(string); ok && v != "" {
		h.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
	}

	if v, ok := tfMap["encryption"].(map[string]interface{}); ok && len(v) > 0 {
		h.Encryption = expandHlsEncryption(v)
	}

	if v, ok := tfMap["include_dvb_subtitles"].(bool); ok {
		h.IncludeDvbSubtitles = aws.Bool(v)
	}

	if v, ok := tfMap["include_iframe_only_stream"].(bool); ok {
		h.IncludeIframeOnlyStream = aws.Bool(v)
	}

	if v, ok := tfMap["playlist_type"].(string); ok && v != "" {
		h.PlaylistType = types.PlaylistType(v)
	}

	if v, ok := tfMap["playlist_window_seconds"].(int); ok {
		h.PlaylistWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["program_date_time_interval_seconds"].(int); ok {
		h.ProgramDateTimeIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok {
		h.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].(map[string]interface{}); ok && len(v) > 0 {
		h.StreamSelection = expandStreamSelection(v)
	}

	if v, ok := tfMap["use_audio_rendition_group"].(bool); ok {
		h.UseAudioRenditionGroup = aws.Bool(v)
	}

	return h
}

func flattenHlsPackage(apiObject *types.HlsPackage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AdMarkers; v != "" {
		m["ad_markers"] = string(v)
	}

	if v := apiObject.AdTriggers; v != nil {
		m["ad_triggers"] = flattenAdTriggers(v)
	}

	if v := apiObject.AdsOnDeliveryRestrictions; v != "" {
		m["ads_on_delivery_restrictions"] = string(v)
	}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenHlsEncryption(v)
	}

	if v := apiObject.IncludeDvbSubtitles; v != nil {
		m["include_dvb_subtitles"] = aws.ToBool(v)
	}

	if v := apiObject.IncludeIframeOnlyStream; v != nil {
		m["include_iframe_only_stream"] = aws.ToBool(v)
	}

	if v := apiObject.PlaylistType; v != "" {
		m["playlist_type"] = string(v)
	}

	if v := apiObject.PlaylistWindowSeconds; v != nil {
		m["playlist_window_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.ProgramDateTimeIntervalSeconds; v != nil {
		m["program_date_time_interval_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	if v := apiObject.UseAudioRenditionGroup; v != nil {
		m["use_audio_rendition_group"] = aws.ToBool(v)
	}

	return m
}

func expandAdTriggers(tfList []interface{}) []types.AdTriggersElement {
	if len(tfList) == 0 {
		return nil
	}

	var as []types.AdTriggersElement

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		var a types.AdTriggersElement

		if v, ok := m["ad_trigger"].(string); ok && v != "" {
			a = types.AdTriggersElement(v)
		}

		as = append(as, a)
	}

	return as
}

func flattenAdTriggers(apiList []types.AdTriggersElement) []interface{} {
	if len(apiList) == 0 {
		return nil
	}

	var l []interface{}

	for _, v := range apiList {
		l = append(l, map[string]interface{}{
			"ad_trigger": v,
		})
	}

	return l
}

func expandHlsEncryption(tfMap map[string]interface{}) *types.HlsEncryption {
	if tfMap == nil {
		return nil
	}

	h := &types.HlsEncryption{}

	if v, ok := tfMap["speke_key_provider"].(map[string]interface{}); ok && v != nil {
		h.SpekeKeyProvider = expandSpekeKeyProvider(v)
	}

	if v, ok := tfMap["constant_initialization_vector"].(string); ok && v != "" {
		h.ConstantInitializationVector = aws.String(v)
	}

	if v, ok := tfMap["encryption_method"].(string); ok && v != "" {
		h.EncryptionMethod = types.EncryptionMethod(v)
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok {
		h.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["repeat_ext_x_key"].(bool); ok {
		h.RepeatExtXKey = aws.Bool(v)
	}

	return h
}

func flattenHlsEncryption(apiObject *types.HlsEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.ConstantInitializationVector; v != nil {
		m["constant_initialization_vector"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionMethod; v != "" {
		m["encryption_method"] = string(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.RepeatExtXKey; v != nil {
		m["repeat_ext_x_key"] = aws.ToBool(v)
	}

	return m
}

func expandSpekeKeyProvider(tfMap map[string]interface{}) *types.SpekeKeyProvider {
	s := &types.SpekeKeyProvider{
		ResourceId: aws.String(tfMap["resource_id"].(string)),
		RoleArn:    aws.String(tfMap["role_arn"].(string)),
		Url:        aws.String(tfMap["url"].(string)),
		SystemIds:  tfMap["system_ids"].([]string),
	}

	if v, ok := tfMap["certificate_arn"].(string); ok && v != "" {
		s.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap["encryption_contract_configuration"]; ok && len(v.(map[string]interface{})) > 0 {
		c := &types.EncryptionContractConfiguration{
			PresetSpeke20Audio: types.PresetSpeke20Audio(v.(map[string]interface{})["preset_speke20_audio"].(string)),
			PresetSpeke20Video: types.PresetSpeke20Video(v.(map[string]interface{})["preset_speke20_video"].(string)),
		}
		s.EncryptionContractConfiguration = c
	}

	return s
}

func flattenSpekeKeyProvider(apiObject *types.SpekeKeyProvider) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"resource_id": aws.ToString(apiObject.ResourceId),
		"role_arn":    aws.ToString(apiObject.RoleArn),
		"url":         aws.ToString(apiObject.Url),
		"system_ids":  apiObject.SystemIds,
	}

	if v := apiObject.CertificateArn; v != nil {
		m["certificate_arn"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionContractConfiguration; v != nil {
		m["encryption_contract_configuration"] = map[string]interface{}{
			"preset_speke20_audio": string(v.PresetSpeke20Audio),
			"preset_speke20_video": string(v.PresetSpeke20Video),
		}
	}

	return m
}

func expandStreamSelection(tfMap map[string]interface{}) *types.StreamSelection {
	s := &types.StreamSelection{}

	if v, ok := tfMap["max_video_bits_per_second"].(int); ok {
		s.MaxVideoBitsPerSecond = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_video_bits_per_second"].(int); ok {
		s.MinVideoBitsPerSecond = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_order"].(string); ok {
		s.StreamOrder = types.StreamOrder(v)
	}

	return s
}

func flattenStreamSelection(apiObject *types.StreamSelection) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaxVideoBitsPerSecond; v != nil {
		m["max_video_bits_per_second"] = aws.ToInt32(v)
	}

	if v := apiObject.MinVideoBitsPerSecond; v != nil {
		m["min_video_bits_per_second"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamOrder; v != "" {
		m["stream_order"] = string(v)
	}

	return m
}

func expandMssPackage(tfMap map[string]interface{}) *types.MssPackage {
	m := &types.MssPackage{}

	if v, ok := tfMap["encryption"]; ok && len(v.(map[string]interface{})) > 0 {
		m.Encryption = expandMssEncryption(v.(map[string]interface{}))
	}

	if v, ok := tfMap["manifest_window_seconds"].(int); ok {
		m.ManifestWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok {
		m.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"]; ok && len(v.(map[string]interface{})) > 0 {
		m.StreamSelection = expandStreamSelection(v.(map[string]interface{}))
	}

	return m
}

func flattenMssPackage(apiObject *types.MssPackage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenMssEncryption(v)
	}

	if v := apiObject.ManifestWindowSeconds; v != nil {
		m["manifest_window_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	return m
}

func expandMssEncryption(mssEncryptionSettings map[string]interface{}) *types.MssEncryption {
	m := &types.MssEncryption{}

	if v, ok := mssEncryptionSettings["speke_key_provider"]; ok && len(v.(map[string]interface{})) > 0 {
		m.SpekeKeyProvider = expandSpekeKeyProvider(v.(map[string]interface{}))
	}

	return m
}

func flattenMssEncryption(apiObject *types.MssEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	return m
}

func expandCmafPackage(cmafPackageSettings map[string]interface{}) *types.CmafPackageCreateOrUpdateParameters {
	c := &types.CmafPackageCreateOrUpdateParameters{}

	if v, ok := cmafPackageSettings["encryption"]; ok && len(v.(map[string]interface{})) > 0 {
		c.Encryption = expandCmafEncryption(v.(map[string]interface{}))
	}

	if v, ok := cmafPackageSettings["hls_manifests"]; ok {
		c.HlsManifests = expandHlsManifests(v.([]interface{}))
	}

	if v, ok := cmafPackageSettings["segment_duration_seconds"].(int); ok && v > 0 {
		c.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := cmafPackageSettings["segment_prefix"].(string); ok && v != "" {
		c.SegmentPrefix = aws.String(v)
	}

	if v, ok := cmafPackageSettings["stream_selection"]; ok && len(v.(map[string]interface{})) > 0 {
		c.StreamSelection = expandStreamSelection(v.(map[string]interface{}))
	}

	return c
}

func flattenCmafPackage(apiObject *types.CmafPackage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenCmafEncryption(v)
	}

	if v := apiObject.HlsManifests; v != nil {
		m["hls_manifests"] = flattenHlsManifests(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.SegmentPrefix; v != nil {
		m["segment_prefix"] = aws.ToString(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	return m
}

func expandDashPackage(tfMap map[string]interface{}) *types.DashPackage {
	if tfMap == nil {
		return nil
	}

	d := &types.DashPackage{}

	if v, ok := tfMap["ad_triggers"].([]interface{}); ok && len(v) > 0 {
		d.AdTriggers = expandAdTriggers(v)
	}

	if v, ok := tfMap["ads_on_delivery_restrictions"].(string); ok && v != "" {
		d.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
	}

	if v, ok := tfMap["encryption"].(map[string]interface{}); ok && len(v) > 0 {
		d.Encryption = expandDashEncryption(v)
	}

	if v, ok := tfMap["manifest_window_seconds"].(int); ok && v > 0 {
		d.ManifestWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_buffer_time_seconds"].(int); ok && v > 0 {
		d.MinBufferTimeSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_update_period_seconds"].(int); ok && v > 0 {
		d.MinUpdatePeriodSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["period_triggers"].([]interface{}); ok && len(v) > 0 {
		d.PeriodTriggers = expandPeriodTriggers(v)
	}

	if v, ok := tfMap["profile"].(string); ok && v != "" {
		d.Profile = types.Profile(v)
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok && v > 0 {
		d.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].(map[string]interface{}); ok && len(v) > 0 {
		d.StreamSelection = expandStreamSelection(v)
	}

	return d
}

func flattenDashPackage(apiObject *types.DashPackage) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AdTriggers; v != nil {
		m["ad_triggers"] = flattenAdTriggers(v)
	}

	if v := apiObject.AdsOnDeliveryRestrictions; v != "" {
		m["ads_on_delivery_restrictions"] = string(v)
	}

	if v := apiObject.Encryption; v != nil {
		m["encryption"] = flattenDashEncryption(v)
	}

	if v := apiObject.ManifestWindowSeconds; v != nil {
		m["manifest_window_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MinBufferTimeSeconds; v != nil {
		m["min_buffer_time_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.MinUpdatePeriodSeconds; v != nil {
		m["min_update_period_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.PeriodTriggers; v != nil {
		m["period_triggers"] = flattenPeriodTriggers(v)
	}

	if v := apiObject.Profile; v != "" {
		m["profile"] = string(v)
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = aws.ToInt32(v)
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	return m
}

func expandDashEncryption(tfMap map[string]interface{}) *types.DashEncryption {
	if tfMap == nil {
		return nil
	}

	d := &types.DashEncryption{}

	if v, ok := tfMap["speke_key_provider"].(map[string]interface{}); ok && len(v) > 0 {
		d.SpekeKeyProvider = expandSpekeKeyProvider(v)
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok && v > 0 {
		d.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	return d
}

func flattenDashEncryption(apiObject *types.DashEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = aws.ToInt32(v)
	}

	return m
}

func expandPeriodTriggers(tfList []interface{}) []types.PeriodTriggersElement {
	if len(tfList) == 0 {
		return nil
	}

	var l []types.PeriodTriggersElement

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		var p types.PeriodTriggersElement

		if v, ok := m["period_trigger"].(string); ok && v != "" {
			p = types.PeriodTriggersElement(v)
		}

		l = append(l, p)
	}

	return l
}

func flattenPeriodTriggers(apiList []types.PeriodTriggersElement) []interface{} {
	if len(apiList) == 0 {
		return nil
	}

	var l []interface{}

	for _, v := range apiList {
		l = append(l, map[string]interface{}{
			"period_trigger": v,
		})
	}

	return l
}

func expandCmafEncryption(tfMap map[string]interface{}) *types.CmafEncryption {
	if tfMap == nil {
		return nil
	}

	e := &types.CmafEncryption{}

	if v, ok := tfMap["speke_key_provider"].(map[string]interface{}); ok && len(v) > 0 {
		e.SpekeKeyProvider = expandSpekeKeyProvider(v)
	}

	if v, ok := tfMap["constant_initialization_vector"].(string); ok && v != "" {
		e.ConstantInitializationVector = aws.String(v)
	}

	if v, ok := tfMap["encryption_method"].(string); ok && v != "" {
		e.EncryptionMethod = types.CmafEncryptionMethod(v)
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok && v > 0 {
		e.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	return e
}

func flattenCmafEncryption(apiObject *types.CmafEncryption) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SpekeKeyProvider; v != nil {
		m["speke_key_provider"] = flattenSpekeKeyProvider(v)
	}

	if v := apiObject.ConstantInitializationVector; v != nil {
		m["constant_initialization_vector"] = aws.ToString(v)
	}

	if v := apiObject.EncryptionMethod; v != "" {
		m["encryption_method"] = string(v)
	}

	if v := apiObject.KeyRotationIntervalSeconds; v != nil {
		m["key_rotation_interval_seconds"] = aws.ToInt32(v)
	}

	return m
}

func expandHlsManifests(tfList []interface{}) []types.HlsManifestCreateOrUpdateParameters {
	if len(tfList) == 0 {
		return nil
	}

	var hs []types.HlsManifestCreateOrUpdateParameters

	for _, tfManifest := range tfList {
		manifest := tfManifest.(map[string]interface{})
		m := types.HlsManifestCreateOrUpdateParameters{
			Id: aws.String(manifest["id"].(string)),
		}

		if v, ok := manifest["ad_markers"].(string); ok && v != "" {
			m.AdMarkers = types.AdMarkers(v)
		}

		if v, ok := manifest["ad_triggers"].([]interface{}); ok && len(v) > 0 {
			m.AdTriggers = expandAdTriggers(v)
		}

		if v, ok := manifest["ads_on_delivery_restrictions"].(string); ok && v != "" {
			m.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
		}

		if v, ok := manifest["include_iframe_only_stream"].(bool); ok {
			m.IncludeIframeOnlyStream = aws.Bool(v)
		}

		if v, ok := manifest["manifest_name"].(string); ok && v != "" {
			m.ManifestName = aws.String(v)
		}

		if v, ok := manifest["playlist_type"].(string); ok && v != "" {
			m.PlaylistType = types.PlaylistType(v)
		}

		if v, ok := manifest["playlist_window_seconds"].(int); ok && v > 0 {
			m.PlaylistWindowSeconds = aws.Int32(int32(v))
		}

		if v, ok := manifest["program_date_time_interval_seconds"].(int); ok && v > 0 {
			m.ProgramDateTimeIntervalSeconds = aws.Int32(int32(v))
		}

		hs = append(hs, m)
	}

	return hs
}

func flattenHlsManifests(apiList []types.HlsManifest) []interface{} {
	if len(apiList) == 0 {
		return nil
	}

	var hs []interface{}

	for _, manifest := range apiList {
		m := map[string]interface{}{
			"id": manifest.Id,
		}

		if v := manifest.AdMarkers; v != "" {
			m["ad_markers"] = string(v)
		}

		if v := manifest.AdTriggers; v != nil {
			m["ad_triggers"] = flattenAdTriggers(v)
		}

		if v := manifest.AdsOnDeliveryRestrictions; v != "" {
			m["ads_on_delivery_restrictions"] = string(v)
		}

		if v := manifest.IncludeIframeOnlyStream; v != nil {
			m["include_iframe_only_stream"] = aws.ToBool(v)
		}

		if v := manifest.ManifestName; v != nil {
			m["manifest_name"] = aws.ToString(v)
		}

		if v := manifest.PlaylistType; v != "" {
			m["playlist_type"] = string(v)
		}

		if v := manifest.PlaylistWindowSeconds; v != nil {
			m["playlist_window_seconds"] = aws.ToInt32(v)
		}

		if v := manifest.ProgramDateTimeIntervalSeconds; v != nil {
			m["program_date_time_interval_seconds"] = aws.ToInt32(v)
		}

		hs = append(hs, m)
	}

	return hs
}
