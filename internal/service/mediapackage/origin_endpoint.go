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
	awstypes "github.com/aws/aws-sdk-go-v2/service/mediapackage/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
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
				Optional: true,
				Computed: true,
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
								Schema: encryptionSchema,
							},
						},
						"hls_manifests": {
							Type:     schema.TypeList,
							Optional: true,
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
										Elem:     &schema.Schema{Type: schema.TypeString},
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
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
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
							Elem:     &schema.Schema{Type: schema.TypeString},
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
													Elem:     &schema.Schema{Type: schema.TypeString},
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
							Elem:     &schema.Schema{Type: schema.TypeString},
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
								Schema: streamSelectionSchema,
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
					Schema: hlsPackageSchema,
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
													Elem:     &schema.Schema{Type: schema.TypeString},
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
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: streamSelectionSchema,
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
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		Id:        aws.String("test"),
		Tags:      getTagsIn(ctx),
	}
	if v, ok := d.GetOk("authorization"); ok {
		in.Authorization = expandAuthorization(v.(*schema.Set).List()[0].(map[string]interface{}))
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
		in.HlsPackage = expandHlsPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("mss_package"); ok {
		in.MssPackage = expandMssPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
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
		in.Whitelist = flex.ExpandStringValueList(v.([]interface{}))
	}

	// TIP: -- 3. Call the AWS create function
	out, err := conn.CreateOriginEndpoint(ctx, in)
	if err != nil {
		// TIP: Since d.SetId() has not been called yet, you cannot use d.Id()
		// in error messages at this point.
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get("id").(string), err)
	}

	if out == nil || out.Origination == "" {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionCreating, ResNameOriginEndpoint, d.Get("id").(string), errors.New("empty output"))
	}

	d.SetId(aws.ToString(out.Id)) // TODO

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
	out, err := findOriginEndpoint(ctx, conn, d.Get("id").(string), d.Get("channel_id").(string))

	// TIP: -- 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MediaPackage OriginEndpoint (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionReading, ResNameOriginEndpoint, d.Id(), err)
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
	d.Set("arn", out.Arn)

	// TIP: Setting a complex type.
	// For more information, see:
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#data-handling-and-conversion
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#flatten-functions-for-blocks
	// https://hashicorp.github.io/terraform-provider-aws/data-handling-and-conversion/#root-typeset-of-resource-and-aws-list-of-structure

	if err := d.Set("authorization", flattenAuthorization(out.Authorization)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("cmaf_package", flattenCmafPackage(out.CmafPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("dash_package", flattenDashPackage(out.DashPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("description", out.Description); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}
	if err := d.Set("hls_package", flattenHlsPackage(out.HlsPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("mss_package", flattenMssPackage(out.MssPackage)); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("origination", out.Origination); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("start_over_window_seconds", out.StartoverWindowSeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("time_delay_seconds", out.TimeDelaySeconds); err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionSetting, ResNameOriginEndpoint, d.Id(), err)
	}

	if err := d.Set("whitelist", out.Whitelist); err != nil {
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

	if v, ok := d.GetOk("authorization"); ok {
		in.Authorization = expandAuthorization(v.(*schema.Set).List()[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cmaf_package"); ok && len(v.(*schema.Set).List()) > 0 {
		in.CmafPackage = expandCmafPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
		update = true
	}

	if v, ok := d.GetOk("dash_package"); ok && len(v.(*schema.Set).List()) > 0 {
		in.DashPackage = expandDashPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
		update = true
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
		update = true
	}

	if v, ok := d.GetOk("hls_package"); ok {
		in.HlsPackage = expandHlsPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
		update = true
	}

	if v, ok := d.GetOk("mss_package"); ok {
		in.MssPackage = expandMssPackage(v.(*schema.Set).List()[0].(map[string]interface{}))
		update = true
	}

	if v, ok := d.GetOk("manifest_name"); ok {
		in.Description = aws.String(v.(string))
		update = true
	}

	if v, ok := d.GetOk("origination"); ok {
		in.Origination = types.Origination(v.(string))
		update = true
	}

	if v, ok := d.GetOk("start_over_window_seconds"); ok {
		in.StartoverWindowSeconds = aws.Int32(int32(v.(int)))
		update = true
	}

	if v, ok := d.GetOk("time_delay_seconds"); ok {
		in.TimeDelaySeconds = aws.Int32(int32(v.(int)))
		update = true
	}

	if v, ok := d.GetOk("whitelist"); ok {
		in.Whitelist = flex.ExpandStringValueList(v.([]interface{}))
		update = true
	}

	if !update {
		return diags
	}

	log.Printf("[DEBUG] Updating MediaPackage OriginEndpoint (%s): %#v", d.Id(), in)
	_, err := conn.UpdateOriginEndpoint(ctx, in)
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionUpdating, ResNameOriginEndpoint, d.Id(), err)
	}

	return append(diags, resourceOriginEndpointRead(ctx, d, meta)...)
}

func resourceOriginEndpointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).MediaPackageClient(ctx)

	log.Printf("[INFO] Deleting MediaPackage OriginEndpoint %s", d.Id())

	_, err := conn.DeleteOriginEndpoint(ctx, &mediapackage.DeleteOriginEndpointInput{
		Id: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return create.AppendDiagError(diags, names.MediaPackage, create.ErrActionDeleting, ResNameOriginEndpoint, d.Id(), err)
	}

	return diags
}

// TIP: ==== FINDERS ====
// The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findOriginEndpoint(ctx context.Context, conn *mediapackage.Client, id, channelID string) (*types.OriginEndpoint, error) {
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

	var ep *types.OriginEndpoint

	for _, e := range out.OriginEndpoints {
		if aws.ToString(e.Id) == id {
			ep = &e
		}
	}

	return ep, nil
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

func flattenAuthorization(apiObject *types.Authorization) interface{} {
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

	if v, ok := tfMap["encryption"].(*schema.Set); ok && v.Len() > 0 {
		h.Encryption = expandHlsEncryption(v.List()[0].(map[string]interface{}))
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

	if v, ok := tfMap["stream_selection"].(*schema.Set); ok && v.Len() > 0 {
		h.StreamSelection = expandStreamSelection(v.List()[0].(map[string]interface{}))
	}

	if v, ok := tfMap["use_audio_rendition_group"].(bool); ok {
		h.UseAudioRenditionGroup = aws.Bool(v)
	}

	return h
}

func flattenHlsPackage(apiObject *types.HlsPackage) *schema.Set {
	m := map[string]interface{}{}

	if v := apiObject.AdMarkers; v != "" {
		m["ad_markers"] = string(v)
	}

	if v := apiObject.AdTriggers; v != nil {
		m["ad_triggers"] = flex.FlattenStringyValueList(v)
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
		m["playlist_window_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.ProgramDateTimeIntervalSeconds; v != nil {
		m["program_date_time_interval_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.SegmentDurationSeconds; v != nil {
		m["segment_duration_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.StreamSelection; v != nil {
		m["stream_selection"] = flattenStreamSelection(v)
	}

	if v := apiObject.UseAudioRenditionGroup; v != nil {
		m["use_audio_rendition_group"] = aws.ToBool(v)
	}

	resource := &schema.Resource{
		Schema: hlsPackageSchema,
	}

	s := schema.NewSet(schema.HashResource(resource), []interface{}{m})

	return s
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

func expandHlsEncryption(tfMap map[string]interface{}) *types.HlsEncryption {
	if tfMap == nil {
		return nil
	}

	h := &types.HlsEncryption{}

	if v, ok := tfMap["speke_key_provider"].(*schema.Set); ok && v.Len() > 0 {
		h.SpekeKeyProvider = expandSpekeKeyProvider(v.List()[0].(map[string]interface{}))
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

func flattenHlsEncryption(apiObject *types.HlsEncryption) *schema.Set {
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

	resource := &schema.Resource{
		Schema: encryptionSchema,
	}

	s := schema.NewSet(schema.HashResource(resource), []interface{}{m})

	return s
}

func expandSpekeKeyProvider(tfMap map[string]interface{}) *types.SpekeKeyProvider {
	s := &types.SpekeKeyProvider{
		ResourceId: aws.String(tfMap["resource_id"].(string)),
		RoleArn:    aws.String(tfMap["role_arn"].(string)),
		Url:        aws.String(tfMap["url"].(string)),
	}

	if v, ok := tfMap["system_ids"].([]interface{}); ok && len(v) > 0 {
		s.SystemIds = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["certificate_arn"].(string); ok && v != "" {
		s.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap["encryption_contract_configuration"].(*schema.Set); ok && v.Len() > 0 {
		c := &types.EncryptionContractConfiguration{
			PresetSpeke20Audio: types.PresetSpeke20Audio(v.List()[0].(map[string]interface{})["preset_speke20_audio"].(string)),
			PresetSpeke20Video: types.PresetSpeke20Video(v.List()[0].(map[string]interface{})["preset_speke20_video"].(string)),
		}
		s.EncryptionContractConfiguration = c
	}

	return s
}

func flattenSpekeKeyProvider(apiObject *types.SpekeKeyProvider) interface{} {
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

func flattenStreamSelection(apiObject *types.StreamSelection) interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaxVideoBitsPerSecond; v != nil {
		m["max_video_bits_per_second"] = int(aws.ToInt32(v))
	}

	if v := apiObject.MinVideoBitsPerSecond; v != nil {
		m["min_video_bits_per_second"] = int(aws.ToInt32(v))
	}

	if v := apiObject.StreamOrder; v != "" {
		m["stream_order"] = string(v)
	}

	resource := &schema.Resource{
		Schema: hlsPackageSchema,
	}

	s := schema.NewSet(schema.HashResource(resource), []interface{}{m})

	return s
}

func expandMssPackage(tfMap map[string]interface{}) *types.MssPackage {
	m := &types.MssPackage{}

	if v, ok := tfMap["encryption"].(*schema.Set); ok && v.Len() > 0 {
		m.Encryption = expandMssEncryption(v.List()[0].(map[string]interface{}))
	}

	if v, ok := tfMap["manifest_window_seconds"].(int); ok {
		m.ManifestWindowSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok {
		m.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].(*schema.Set); ok && v.Len() > 0 {
		m.StreamSelection = expandStreamSelection(v.List()[0].(map[string]interface{}))
	}

	return m
}

func flattenMssPackage(apiObject *types.MssPackage) interface{} {
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

	if v, ok := mssEncryptionSettings["speke_key_provider"].(*schema.Set); ok && v.Len() > 0 {
		m.SpekeKeyProvider = expandSpekeKeyProvider(v.List()[0].(map[string]interface{}))
	}

	return m
}

func flattenMssEncryption(apiObject *types.MssEncryption) interface{} {
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

	if v, ok := cmafPackageSettings["encryption"].(*schema.Set); ok && v.Len() > 0 {
		c.Encryption = expandCmafEncryption(v.List()[0].(map[string]interface{}))
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

	if v, ok := cmafPackageSettings["stream_selection"].(*schema.Set); ok && v.Len() > 0 {
		c.StreamSelection = expandStreamSelection(v.List()[0].(map[string]interface{}))
	}

	return c
}

func flattenCmafPackage(apiObject *types.CmafPackage) interface{} {
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

	if v, ok := tfMap["ad_triggers"].(*schema.Set); ok && v.Len() > 0 {
		d.AdTriggers = expandAdTriggers(v.List())
	}

	if v, ok := tfMap["ads_on_delivery_restrictions"].(string); ok && v != "" {
		d.AdsOnDeliveryRestrictions = types.AdsOnDeliveryRestrictions(v)
	}

	if v, ok := tfMap["encryption"].(*schema.Set); ok && v.Len() > 0 {
		d.Encryption = expandDashEncryption(v.List()[0].(map[string]interface{}))
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

	if v, ok := tfMap["period_triggers"].(*schema.Set); ok && v.Len() > 0 {
		d.PeriodTriggers = expandPeriodTriggers(v.List())
	}

	if v, ok := tfMap["profile"].(string); ok && v != "" {
		d.Profile = types.Profile(v)
	}

	if v, ok := tfMap["segment_duration_seconds"].(int); ok && v > 0 {
		d.SegmentDurationSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["stream_selection"].(*schema.Set); ok && v.Len() > 0 {
		d.StreamSelection = expandStreamSelection(v.List()[0].(map[string]interface{}))
	}

	return d
}

func flattenDashPackage(apiObject *types.DashPackage) interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AdTriggers; v != nil {
		m["ad_triggers"] = flex.FlattenStringyValueList(v)
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

	if v, ok := tfMap["speke_key_provider"].(*schema.Set); ok && v.Len() > 0 {
		d.SpekeKeyProvider = expandSpekeKeyProvider(v.List()[0].(map[string]interface{}))
	}

	if v, ok := tfMap["key_rotation_interval_seconds"].(int); ok && v > 0 {
		d.KeyRotationIntervalSeconds = aws.Int32(int32(v))
	}

	return d
}

func flattenDashEncryption(apiObject *types.DashEncryption) interface{} {
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

	if v, ok := tfMap["speke_key_provider"].(*schema.Set); ok && v.Len() > 0 {
		e.SpekeKeyProvider = expandSpekeKeyProvider(v.List()[0].(map[string]interface{}))
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

func flattenCmafEncryption(apiObject *types.CmafEncryption) interface{} {
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

		if v, ok := manifest["ad_triggers"].(*schema.Set); ok && v.Len() > 0 {
			m.AdTriggers = expandAdTriggers(v.List())
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
			m["ad_triggers"] = flex.FlattenStringyValueList(v)
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

var hlsPackageSchema = map[string]*schema.Schema{
	"ad_markers": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"ad_triggers": {
		Type:     schema.TypeList,
		Optional: true,
		Elem:     &schema.Schema{Type: schema.TypeString},
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
								Elem:     &schema.Schema{Type: schema.TypeString},
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
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: streamSelectionSchema,
		},
	},
	"use_audio_rendition_group": {
		Type:     schema.TypeBool,
		Optional: true,
	},
}

var encryptionSchema = map[string]*schema.Schema{
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
					Elem:     &schema.Schema{Type: schema.TypeString},
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
}

var streamSelectionSchema = map[string]*schema.Schema{
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
}
