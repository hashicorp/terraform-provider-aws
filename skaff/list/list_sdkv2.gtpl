{{- define "Factory" -}}
// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// {{ template "Annotation" . }}
func {{ template "FactoryFunctionName" . }}() inttypes.ListResourceForSDK {
	l := {{ template "ListResourceStructName" . }}{}
	l.SetResourceSchema(resource{{ .ListResource }}())
	return &l
}
{{- end }}

{{- define "Annotation" -}}
@SDKListResource("{{ .ProviderResourceName }}")
{{- end }}

{{- define "ListResourceStruct" -}}
type {{ template "ListResourceStructName" . }} struct {
	framework.ListResourceWithSDKv2Resource
}
{{- end }}

{{- define "GoImports" -}}
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
{{- end }}

{{- define "ReadBody" -}}
	 		{{- template "ReadBodyLogging" . }}

			result := request.NewListResult(ctx)
			
			rd := l.ResourceData()
			rd.SetId(arn)
			{{ if .IncludeComments -}}
			// TIP: -- 6. Populate additional attributes needed for Resource Identity
			{{- end }}
			rd.Set(names.AttrName, name)

			if request.IncludeResource {
				if err := resource{{ .ListResource }}Flatten(ctx, l.Meta(), &item, rd); err != nil {
					tflog.Error(ctx, "Reading {{ .HumanFriendlyServiceShort }} {{ .HumanListResourceName }}", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			{{ if .IncludeComments -}}
			// TIP: -- 7. Set the display name
			{{- end }}
			result.DisplayName = aws.ToString(item.{{ .ListResource }}Name)

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}
{{- end }}

{{- define "ResourceFlattenFunction" -}}
{{- if .IncludeComments -}}
// TIP: ==== RESOURCE FLATTENING FUNCTION ====
// This function should be placed in the resource type's source file ("{{ .ListResourceSnake }}.go").
// It is intended to perform the flattening of the results of the API call or calls used to populate a resource's values.
// It should replace most of the body of the resource type's Read function (`resource{{ .ListResource }}Read`) and take the API results
// as parameters.
// The replaced section of the Read function should be
// 	if err := resource{{ .ListResource }}Flatten(ctx, meta.(*conns.AWSClient), {{ .ListResourceLowerCamel }}, d); err != nil {
// 		return sdkdiag.AppendFromErr(diags, err)
// 	}
{{- end }}
func resource{{ .ListResource }}Flatten(ctx context.Context, awsClient *conns.AWSClient, {{ .ListResourceLowerCamel }} *awstypes.{{ .ListResourceAWS }}, d *schema.ResourceData) error {
	d.Set(names.AttrARN, awsClient.RegionalARN(ctx, "{{ .ARNNamespace }}", "{{ .ListResourceLower }}/"+d.Id()))
	if err := d.Set("some_collection", flattenSomeCollection(someCollection)); err != nil {
		return fmt.Errorf("setting some_collection: %w", err)
	}

	return nil
}
{{- end }}
