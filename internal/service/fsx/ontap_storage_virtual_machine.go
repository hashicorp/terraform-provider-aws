// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_fsx_ontap_storage_virtual_machine", name="ONTAP Storage Virtual Machine")
// @Tags(identifierAttribute="arn")
func resourceONTAPStorageVirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceONTAPStorageVirtualMachineCreate,
		ReadWithoutTimeout:   resourceONTAPStorageVirtualMachineRead,
		UpdateWithoutTimeout: resourceONTAPStorageVirtualMachineUpdate,
		DeleteWithoutTimeout: resourceONTAPStorageVirtualMachineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceONTAPStorageVirtualMachineV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceONTAPStorageVirtualMachineStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"active_directory_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"netbios_name": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: sdkv2.SuppressEquivalentStringCaseInsensitive,
							ValidateFunc:     validation.StringLenBetween(1, 15),
						},
						"self_managed_active_directory_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_ips": {
										Type:     schema.TypeSet,
										Required: true,
										MinItems: 1,
										MaxItems: 3,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsIPAddress,
										},
									},
									"domain_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"file_system_administrators_group": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"organizational_unit_distinguished_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 2000),
									},
									"password": {
										Type:         schema.TypeString,
										Sensitive:    true,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"username": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
								},
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoints": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iscsi": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"management": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"nfs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"smb": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ip_addresses": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
			"file_system_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(11, 21),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 47),
			},
			"root_volume_security_style": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(fsx.StorageVirtualMachineRootVolumeSecurityStyle_Values(), false),
			},
			"subtype": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"svm_admin_password": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringLenBetween(8, 50),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceONTAPStorageVirtualMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &fsx.CreateStorageVirtualMachineInput{
		FileSystemId: aws.String(d.Get("file_system_id").(string)),
		Name:         aws.String(name),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("active_directory_configuration"); ok {
		input.ActiveDirectoryConfiguration = expandCreateSvmActiveDirectoryConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_volume_security_style"); ok {
		input.RootVolumeSecurityStyle = aws.String(v.(string))
	}

	if v, ok := d.GetOk("svm_admin_password"); ok {
		input.SvmAdminPassword = aws.String(v.(string))
	}

	output, err := conn.CreateStorageVirtualMachineWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating FSx ONTAP Storage Virtual Machine (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.StorageVirtualMachine.StorageVirtualMachineId))

	if _, err := waitStorageVirtualMachineCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Storage Virtual Machine (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceONTAPStorageVirtualMachineRead(ctx, d, meta)...)
}

func resourceONTAPStorageVirtualMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	storageVirtualMachine, err := findStorageVirtualMachineByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] FSx ONTAP Storage Virtual Machine (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx ONTAP Storage Virtual Machine (%s): %s", d.Id(), err)
	}

	if err := d.Set("active_directory_configuration", flattenSvmActiveDirectoryConfiguration(d, storageVirtualMachine.ActiveDirectoryConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting active_directory_configuration: %s", err)
	}
	d.Set(names.AttrARN, storageVirtualMachine.ResourceARN)
	if err := d.Set("endpoints", flattenSvmEndpoints(storageVirtualMachine.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints: %s", err)
	}
	d.Set("file_system_id", storageVirtualMachine.FileSystemId)
	d.Set(names.AttrName, storageVirtualMachine.Name)
	// RootVolumeSecurityStyle and SVMAdminPassword are write only properties so they don't get returned from the describe API so we just store the original setting to state
	d.Set("root_volume_security_style", d.Get("root_volume_security_style").(string))
	d.Set("subtype", storageVirtualMachine.Subtype)
	d.Set("svm_admin_password", d.Get("svm_admin_password").(string))
	d.Set("uuid", storageVirtualMachine.UUID)

	setTagsOut(ctx, storageVirtualMachine.Tags)

	return diags
}

func resourceONTAPStorageVirtualMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &fsx.UpdateStorageVirtualMachineInput{
			ClientRequestToken:      aws.String(id.UniqueId()),
			StorageVirtualMachineId: aws.String(d.Id()),
		}

		if d.HasChange("active_directory_configuration") {
			input.ActiveDirectoryConfiguration = expandUpdateSvmActiveDirectoryConfiguration(d.Get("active_directory_configuration").([]interface{}))
		}

		if d.HasChange("svm_admin_password") {
			input.SvmAdminPassword = aws.String(d.Get("svm_admin_password").(string))
		}

		_, err := conn.UpdateStorageVirtualMachineWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating FSx ONTAP Storage Virtual Machine (%s): %s", d.Id(), err)
		}

		if _, err := waitStorageVirtualMachineUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Storage Virtual Machine (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceONTAPStorageVirtualMachineRead(ctx, d, meta)...)
}

func resourceONTAPStorageVirtualMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	log.Printf("[DEBUG] Deleting FSx ONTAP Storage Virtual Machine: %s", d.Id())
	_, err := conn.DeleteStorageVirtualMachineWithContext(ctx, &fsx.DeleteStorageVirtualMachineInput{
		StorageVirtualMachineId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeStorageVirtualMachineNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting FSx ONTAP Storage Virtual Machine (%s): %s", d.Id(), err)
	}

	if _, err := waitStorageVirtualMachineDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for FSx ONTAP Storage Virtual Machine (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandCreateSvmActiveDirectoryConfiguration(cfg []interface{}) *fsx.CreateSvmActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.CreateSvmActiveDirectoryConfiguration{}

	if v, ok := conf["netbios_name"].(string); ok && len(v) > 0 {
		out.NetBiosName = aws.String(v)
	}

	if v, ok := conf["self_managed_active_directory_configuration"].([]interface{}); ok {
		out.SelfManagedActiveDirectoryConfiguration = expandSelfManagedActiveDirectoryConfiguration(v)
	}

	return &out
}

func expandSelfManagedActiveDirectoryConfiguration(cfg []interface{}) *fsx.SelfManagedActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.SelfManagedActiveDirectoryConfiguration{}

	if v, ok := conf["dns_ips"].(*schema.Set); ok {
		out.DnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := conf["domain_name"].(string); ok && len(v) > 0 {
		out.DomainName = aws.String(v)
	}

	if v, ok := conf["file_system_administrators_group"].(string); ok && len(v) > 0 {
		out.FileSystemAdministratorsGroup = aws.String(v)
	}

	if v, ok := conf["organizational_unit_distinguished_name"].(string); ok && len(v) > 0 {
		out.OrganizationalUnitDistinguishedName = aws.String(v)
	}

	if v, ok := conf["password"].(string); ok && len(v) > 0 {
		out.Password = aws.String(v)
	}

	if v, ok := conf["username"].(string); ok && len(v) > 0 {
		out.UserName = aws.String(v)
	}

	return &out
}

func expandUpdateSvmActiveDirectoryConfiguration(cfg []interface{}) *fsx.UpdateSvmActiveDirectoryConfiguration {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.UpdateSvmActiveDirectoryConfiguration{}

	if v, ok := conf["netbios_name"].(string); ok && len(v) > 0 {
		out.NetBiosName = aws.String(v)
	}

	if v, ok := conf["self_managed_active_directory_configuration"].([]interface{}); ok {
		out.SelfManagedActiveDirectoryConfiguration = expandSelfManagedActiveDirectoryConfigurationUpdates(v)
	}

	return &out
}

func expandSelfManagedActiveDirectoryConfigurationUpdates(cfg []interface{}) *fsx.SelfManagedActiveDirectoryConfigurationUpdates {
	if len(cfg) < 1 {
		return nil
	}

	conf := cfg[0].(map[string]interface{})

	out := fsx.SelfManagedActiveDirectoryConfigurationUpdates{}

	if v, ok := conf["dns_ips"].(*schema.Set); ok {
		out.DnsIps = flex.ExpandStringSet(v)
	}

	if v, ok := conf["domain_name"].(string); ok && len(v) > 0 {
		out.DomainName = aws.String(v)
	}

	if v, ok := conf["file_system_administrators_group"].(string); ok && len(v) > 0 {
		out.FileSystemAdministratorsGroup = aws.String(v)
	}

	if v, ok := conf["organizational_unit_distinguished_name"].(string); ok && len(v) > 0 {
		out.OrganizationalUnitDistinguishedName = aws.String(v)
	}

	if v, ok := conf["password"].(string); ok && len(v) > 0 {
		out.Password = aws.String(v)
	}

	if v, ok := conf["username"].(string); ok && len(v) > 0 {
		out.UserName = aws.String(v)
	}

	return &out
}

func flattenSvmActiveDirectoryConfiguration(d *schema.ResourceData, rs *fsx.SvmActiveDirectoryConfiguration) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.NetBiosName != nil {
		m["netbios_name"] = rs.NetBiosName
	}

	if rs.SelfManagedActiveDirectoryConfiguration != nil {
		m["self_managed_active_directory_configuration"] = flattenSelfManagedActiveDirectoryAttributes(d, rs.SelfManagedActiveDirectoryConfiguration)
	}

	return []interface{}{m}
}

func flattenSelfManagedActiveDirectoryAttributes(d *schema.ResourceData, rs *fsx.SelfManagedActiveDirectoryAttributes) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DnsIps != nil {
		m["dns_ips"] = aws.StringValueSlice(rs.DnsIps)
	}

	if rs.DomainName != nil {
		m["domain_name"] = aws.StringValue(rs.DomainName)
	}

	if rs.OrganizationalUnitDistinguishedName != nil {
		if _, ok := d.GetOk("active_directory_configuration.0.self_managed_active_directory_configuration.0.organizational_unit_distinguished_name"); ok {
			m["organizational_unit_distinguished_name"] = aws.StringValue(rs.OrganizationalUnitDistinguishedName)
		}
	}

	if rs.UserName != nil {
		m["username"] = aws.StringValue(rs.UserName)
	}

	// Since we are in a configuration block and the FSx API does not return
	// the password or file_system_administrators_group, we need to set the values
	// if we can or Terraform will show a difference for the argument from empty string to the value.
	// This is not a pattern that should be used normally.
	// See also: flattenEmrKerberosAttributes
	if v, ok := d.GetOk("active_directory_configuration.0.self_managed_active_directory_configuration.0.file_system_administrators_group"); ok {
		m["file_system_administrators_group"] = v.(string)
	}
	if v, ok := d.GetOk("active_directory_configuration.0.self_managed_active_directory_configuration.0.password"); ok {
		m["password"] = v.(string)
	}

	return []interface{}{m}
}

func flattenSvmEndpoints(rs *fsx.SvmEndpoints) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.Iscsi != nil {
		m["iscsi"] = flattenSvmEndpoint(rs.Iscsi)
	}
	if rs.Management != nil {
		m["management"] = flattenSvmEndpoint(rs.Management)
	}
	if rs.Nfs != nil {
		m["nfs"] = flattenSvmEndpoint(rs.Nfs)
	}
	if rs.Smb != nil {
		m["smb"] = flattenSvmEndpoint(rs.Smb)
	}
	return []interface{}{m}
}

func flattenSvmEndpoint(rs *fsx.SvmEndpoint) []interface{} {
	if rs == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})
	if rs.DNSName != nil {
		m["dns_name"] = aws.StringValue(rs.DNSName)
	}
	if rs.IpAddresses != nil {
		m["ip_addresses"] = flex.FlattenStringSet(rs.IpAddresses)
	}

	return []interface{}{m}
}

func findStorageVirtualMachineByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.StorageVirtualMachine, error) {
	input := &fsx.DescribeStorageVirtualMachinesInput{
		StorageVirtualMachineIds: []*string{aws.String(id)},
	}

	return findStorageVirtualMachine(ctx, conn, input, tfslices.PredicateTrue[*fsx.StorageVirtualMachine]())
}

func findStorageVirtualMachine(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeStorageVirtualMachinesInput, filter tfslices.Predicate[*fsx.StorageVirtualMachine]) (*fsx.StorageVirtualMachine, error) {
	output, err := findStorageVirtualMachines(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findStorageVirtualMachines(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeStorageVirtualMachinesInput, filter tfslices.Predicate[*fsx.StorageVirtualMachine]) ([]*fsx.StorageVirtualMachine, error) {
	var output []*fsx.StorageVirtualMachine

	err := conn.DescribeStorageVirtualMachinesPagesWithContext(ctx, input, func(page *fsx.DescribeStorageVirtualMachinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.StorageVirtualMachines {
			if v != nil && filter(v) {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeStorageVirtualMachineNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusStorageVirtualMachine(ctx context.Context, conn *fsx.FSx, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStorageVirtualMachineByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Lifecycle), nil
	}
}

func waitStorageVirtualMachineCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecycleCreating, fsx.StorageVirtualMachineLifecyclePending},
		Target:  []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleMisconfigured},
		Refresh: statusStorageVirtualMachine(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if reason := output.LifecycleTransitionReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(reason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitStorageVirtualMachineUpdated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecyclePending},
		Target:  []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleMisconfigured},
		Refresh: statusStorageVirtualMachine(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if reason := output.LifecycleTransitionReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(reason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitStorageVirtualMachineDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleDeleting},
		Target:  []string{},
		Refresh: statusStorageVirtualMachine(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if reason := output.LifecycleTransitionReason; reason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(reason.Message)))
		}

		return output, err
	}

	return nil, err
}
