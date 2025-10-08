// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package templates

import (
	_ "embed"
)

//go:embed header_body.gtpl
var HeaderBody string

//go:embed get_tag_body.gtpl
var GetTagBody string

//go:embed list_tags_body.gtpl
var ListTagsBody string

//go:embed service_tags_map_body.gtpl
var ServiceTagsMapBody string

//go:embed service_tags_value_map_body.gtpl
var ServiceTagsValueMapBody string

//go:embed service_tags_slice_body.gtpl
var ServiceTagsSliceBody string

//go:embed update_tags_body.gtpl
var UpdateTagsBody string

//go:embed wait_tags_propagated_body.gtpl
var WaitTagsPropagatedBody string
