package generate

// TemplateClientSwagger is the text template that generates the Swagger JSON for your API
var TemplateClientSwagger = parseArtifactTemplate("swagger.yaml", `{{ range $service := .Services }}
openapi: 3.0.0
info:
    title: {{ .Name }}
    version: "{{ .Version }}"

servers:
    - url: {{ .Gateway.PathPrefix | LeadingSlash }}

paths:
{{ range $method := .Functions }}
{{ $pathFields := .Gateway.PathParameters }}
{{ $queryFields := .Gateway.QueryParameters }}
    "{{ .Gateway.Path | OpenAPIPath }}":
        {{ .Gateway.Method | ToLower }}:
            description: > {{ range .Documentation }}
                {{ . }}{{ end }}
            {{ if or $pathFields.NotEmpty $queryFields.NotEmpty }}
            parameters:
                {{ range $pathFields }}
                - in: path
                  name: {{ .Name }}
                  required: true
                  {{ if .Field.Documentation.NotEmpty }}
                  description:  > {{ range .Field.Documentation }} 
                      {{ . }}{{ end }}
                  {{ end }}
                  schema:
                      type: {{ .Field.Type.JSONType }}
                {{ end }}
                {{ range $queryFields }}
                - in: query
                  name: {{ .Name }}
                  {{ if .Field.Documentation.NotEmpty }}
                  description:  > {{ range .Field.Documentation }} 
                      {{ . }}{{ end }}
                  {{ end }}
                  schema:
                      type: {{ .Field.Type.JSONType }}
                {{ end }}
            {{ end }}
            {{ if .Gateway.SupportsBody }}
            requestBody:
                content:
                     application/json:
                         schema:
                             $ref: '#/components/schemas/{{ .Request.Name }}'
            {{ end }}
                
            responses:
                {{ .Gateway.Status }}:
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
