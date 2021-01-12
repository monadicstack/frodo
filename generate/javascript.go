package generate

import "text/template"

// Once Go 1.16 comes out and we can embed files in the Go binary, I should pull this out
// into a separate template file and just embed that in the binary fs.
var TemplateClientJS = template.Must(template.New("client.js").Funcs(templateFuncs).Parse(`// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from {{ .Path }}
// !!!!!!! DO NOT EDIT !!!!!!!
const http = require('http');

{{ $ctx := . }}
{{ range .Services }}
class {{ .Name }}Client {
    baseURL;
    httpClient;

    constructor(baseURL, {pathPrefix}) {
        this.baseURL = baseURL;
        this.pathPrefix = pathPrefix || '';
        this.httpClient = http;
    }

    {{ $service := . }}
    {{ range .Methods }}
    async {{ .Name }}(serviceRequest) {
        return new Promise((resolve, reject) => {
            const serviceRequestJSON = JSON.stringify(serviceRequest);
            const options = {
                protocol: 'http:',
                hostname: 'localhost',
                port: '8080',
                path: buildRequestPath('{{ .HTTPPath }}', serviceRequest),
                method: '{{ .HTTPMethod }}',
                headers: {
                    'Content-Type': 'application/json',
                    {{ if .HTTPMethod | HTTPMethodSupportsBody }}
                    'Content-Length': serviceRequestJSON.length,
                    {{ end }}
                }
            };
            console.info(">>>>> CALLING: ", options.path);

            const req = this.httpClient.request(options, (res) => {
                let responseJSON = '';
                res.on('data', (chunk) => {
                    responseJSON += chunk;
                });
                res.on('end', () => {
                    resolve(JSON.parse(responseJSON));
                });
            });
            req.on('error', reject);
            {{ if .HTTPMethod | HTTPMethodSupportsBody }}
            req.write(serviceRequestJSON);
            {{ end }}
            req.end();
        });
    }
    {{ end }}
}
{{ end }}

/**
 * Fills in a router path pattern such as "/user/:id", with the appropriate attribute from
 * the 'serviceRequest' instance.
 *
 * @param {string} path The path pattern to populate w/ runtime values (e.g. "/user/:id")
 * @param {Object} serviceRequest The input struct for the service call
 * @returns {string} The fully-populate URL path (e.g. "/user/aCx31s")
 */
function buildRequestPath(path, serviceRequest) {
    const pathSegments = path.split("/").map(segment => {
        return segment.startsWith(":")
            ? attributeValue(serviceRequest, segment.substring(1))
            : segment;
    });
    return pathSegments.join("/");
}

/**
 *
 * @param {Object} struct The data structure whose value you're trying to peel off.
 * @param {string} attributeName The name of the input value you're trying to send
 * @returns {*}
 */
function attributeValue(struct, attributeName) {
    const normalized = attributeName.toLowerCase();
    for (const key in struct) {
        if (key.toLowerCase() === normalized) {
            return struct[key];
        }
    }
    return null;
}

module.exports = {
    {{ range .Services }}{{ .Name }}Client,{{ end }}
};

(async () => {
    console.info(">>>>> Making Call");
    const client = new GroupServiceClient("http://localhost:8080");
    const result = await client.GetByID({ id: "xyz", flag: true, moo: "cow" });
    console.info(">>>>>> RESULT: ", result);
})().catch(err => {
    console.error(err);
});
`))
