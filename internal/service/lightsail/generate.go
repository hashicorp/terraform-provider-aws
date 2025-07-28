// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run ../../generate/listpages/main.go -ListOps=GetRelationalDatabases,GetLoadBalancers,GetDisks,GetDistributions,GetDomains -InputPaginator=PageToken -OutputPaginator=NextPageToken
//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -TagInIDElem=ResourceName -CreateTags -UpdateTags
//go:generate go run ../../generate/servicepackage/main.go
// ONLY generate directives and package declaration! Do not add anything else to this file.

package lightsail
