//go:generate go run ../../generate/tags/main.go -ListTags -ServiceTagsSlice -UpdateTags -AWSSDKVersion=2
// ONLY generate directives and package declaration! Do not add anything else to this file.

// this uses correct sesv2 service in sdk but sets package wrong in tags_gen...
package sesv2
