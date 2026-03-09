{{- define "Factory" -}}
// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// {{ template "Annotation" . }}
func {{ template "FactoryFunctionName" . }}() list.ListResourceWithConfigure {
	return &{{ template "ListResourceStructName" . }}{}
}
{{- end }}

{{- define "Annotation" -}}
@FrameworkListResource("{{ .ProviderResourceName }}")
{{- end }}

{{- define "ListResourceStruct" -}}
type {{ template "ListResourceStructName" . }} struct {
	{{ .ListResourceLowerCamel }}Resource
	framework.WithList
}
{{- end }}

{{- define "GoImports" -}}
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
{{- end }}

{{- define "ReadBody" -}}
	 		{{- template "ReadBodyLogging" . }}

			result := request.NewListResult(ctx)
			
			var data {{ .ListResourceLowerCamel }}ResourceModel
			{{ if .IncludeComments -}}
			// TIP: -- 6. Set the ID, arguments, and attributes
			// Using a field name prefix allows mapping fields such as `{{ .ListResource }}Id` to `ID`
			{{- end }}
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				result.Diagnostics.Append(r.flatten(ctx, bucket, &data)...)
				if result.Diagnostics.HasError() {
					return
				}

				{{ if .IncludeComments -}}
				// TIP: -- 7. Set the display name
				{{- end }}
				result.DisplayName = aws.ToString(item.{{ .ListResource }}Name)
			})
{{- end }}

{{- define "ResourceFlattenFunction" -}}
{{- if .IncludeComments -}}
// TIP: ==== RESOURCE FLATTENING FUNCTION ====
// This function should be placed in the resource type's source file ("{{ .ListResourceSnake }}.go"). It may already be present.
// It is intended to perform the flattening of the results of the API call or calls used to populate a resource's values.
// It should replace the flattening portion of the resource type's Read function (`{{ .ListResourceLowerCamel }}Resource.Read`) and take the API results
// as parameters.
// The replaced section of the Read function should be
//	response.Diagnostics.Append(r.flatten(ctx, output, &data)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
{{- end }}
// func (r *{{ .ListResourceLowerCamel }}Resource) flatten(ctx context.Context, {{ .ListResourceLowerCamel }} *awstypes.{{ .ListResourceAWS }}, data *{{ .ListResourceLowerCamel }}ResourceModel) (diags diag.Diagnostics) {
// 	diags.Append(fwflex.Flatten(ctx, {{ .ListResourceLowerCamel }}, data)...)
// 	return diags
// }
{{- end }}
