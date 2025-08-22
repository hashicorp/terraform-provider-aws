// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandProtocol(l []any) *awstypes.FsxProtocol {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	protocol := &awstypes.FsxProtocol{}

	if v, ok := m["nfs"].([]any); ok {
		protocol.NFS = expandNFS(v)
	}
	if v, ok := m["smb"].([]any); ok {
		protocol.SMB = expandSMB(v)
	}

	return protocol
}

func flattenProtocol(protocol *awstypes.FsxProtocol) []any {
	if protocol == nil {
		return []any{}
	}

	m := map[string]any{}

	if protocol.NFS != nil {
		m["nfs"] = flattenNFS(protocol.NFS)
	}
	if protocol.SMB != nil {
		m["smb"] = flattenSMB(protocol.SMB)
	}

	return []any{m}
}

func expandNFS(l []any) *awstypes.FsxProtocolNfs {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	protocol := &awstypes.FsxProtocolNfs{
		MountOptions: expandNFSMountOptions(m["mount_options"].([]any)),
	}

	return protocol
}

func expandSMB(l []any) *awstypes.FsxProtocolSmb {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	protocol := &awstypes.FsxProtocolSmb{
		MountOptions: expandSMBMountOptions(m["mount_options"].([]any)),
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
func flattenNFS(nfs *awstypes.FsxProtocolNfs) []any {
	if nfs == nil {
		return []any{}
	}

	m := map[string]any{
		"mount_options": flattenNFSMountOptions(nfs.MountOptions),
	}

	return []any{m}
}

func flattenSMB(smb *awstypes.FsxProtocolSmb) []any {
	if smb == nil {
		return []any{}
	}

	m := map[string]any{
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

	return []any{m}
}
