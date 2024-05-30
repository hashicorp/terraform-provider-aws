// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandProtocol(l []interface{}) *awstypes.FsxProtocol {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	protocol := &awstypes.FsxProtocol{}

	if v, ok := m["nfs"].([]interface{}); ok {
		protocol.NFS = expandNFS(v)
	}
	if v, ok := m["smb"].([]interface{}); ok {
		protocol.SMB = expandSMB(v)
	}

	return protocol
}

func flattenProtocol(protocol *awstypes.FsxProtocol) []interface{} {
	if protocol == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{}

	if protocol.NFS != nil {
		m["nfs"] = flattenNFS(protocol.NFS)
	}
	if protocol.SMB != nil {
		m["smb"] = flattenSMB(protocol.SMB)
	}

	return []interface{}{m}
}

func expandNFS(l []interface{}) *awstypes.FsxProtocolNfs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	protocol := &awstypes.FsxProtocolNfs{
		MountOptions: expandNFSMountOptions(m["mount_options"].([]interface{})),
	}

	return protocol
}

func expandSMB(l []interface{}) *awstypes.FsxProtocolSmb {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	protocol := &awstypes.FsxProtocolSmb{
		MountOptions: expandSMBMountOptions(m["mount_options"].([]interface{})),
	}
	if v, ok := m[names.AttrDomain].(string); ok && v != "" {
		protocol.Domain = aws.String(v)
	}
	if v, ok := m[names.AttrPassword].(string); ok && v != "" {
		protocol.Password = aws.String(v)
	}
	if v, ok := m["user"].(string); ok && v != "" {
		protocol.User = aws.String(v)
	}

	return protocol
}

// todo: go another level down?
func flattenNFS(nfs *awstypes.FsxProtocolNfs) []interface{} {
	if nfs == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mount_options": flattenNFSMountOptions(nfs.MountOptions),
	}

	return []interface{}{m}
}

func flattenSMB(smb *awstypes.FsxProtocolSmb) []interface{} {
	if smb == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mount_options": flattenSMBMountOptions(smb.MountOptions),
	}
	if v := smb.Domain; v != nil {
		m[names.AttrDomain] = aws.ToString(v)
	}
	if v := smb.Password; v != nil {
		m[names.AttrPassword] = aws.ToString(v)
	}
	if v := smb.User; v != nil {
		m["user"] = aws.ToString(v)
	}

	return []interface{}{m}
}
