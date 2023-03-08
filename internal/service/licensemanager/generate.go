//go:generate go run ../../generate/tags/main.go -ServiceTagsSlice -UpdateTags -ContextOnly
//go:generate go run ../../generate/listpages/main.go -ListOps=ListLicenseConfigurations,ListLicenseSpecificationsForResource -ContextOnly
// ONLY generate directives and package declaration! Do not add anything else to this file.

package licensemanager
