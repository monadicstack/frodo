package generate

// TemplateClientSwagger is the text template that generates the Swagger JSON for your API
var TemplateClientSwagger = parseArtifactTemplate("swagger.yaml", `{{ range $service := .Services }}
openapi: 3.0.0
info:
    title: {{ .Name }}
    version: "{{ .Version }}"

servers:
    - url: {{ .HTTPPathPrefix | LeadingSlash }}

paths:
{{ range $method := .Methods }}
    "{{ .HTTPPath | OpenAPIPath}}":
        {{ .HTTPMethod | ToLower }}:
            description: > {{ range .Documentation }}
                {{ . }}
{{ end }}
            {{ if .HTTPMethod | HTTPMethodSupportsBody }}
            requestBody:
                content:
                     application/json:
                         schema:
                             $ref: '#/components/schemas/{{ .Request.Name }}'
            {{ end }}
                
            responses:
                {{ .HTTPStatus }}:
                    description: Success
                    content:
                        application/json:
                            schema:
                                $ref: '#/components/schemas/{{ .Response.Name }}'
{{ end }}

{{ end }}

components:
    schemas:
{{ range $model := .Models }}
        {{ .Name }}:
            type: object
            {{ if .Fields.NotEmpty }}properties:
{{ range $field := .Fields }}
                {{ .Name }}:
                    type: {{ .Type.JSONType }}
                    {{ if .Documentation.NotEmpty }}description: > {{ range .Documentation }} 
                        {{ . }}{{ end }}
                    {{ end }}
            {{ end }}
{{ end }}
{{ end }}
`)
