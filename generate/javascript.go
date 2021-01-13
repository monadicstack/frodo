package generate

// TemplateClientJS is the text template that generates a JavaScript/Node based client for interacting
// with the remote service over RPC/REST.
var TemplateClientJS = parseArtifactTemplate("client.js", `// !!!!!!! DO NOT EDIT !!!!!!!
// Auto-generated client code from {{ .Path }}
// !!!!!!! DO NOT EDIT !!!!!!!
const defaultHTTPClient = require('http');
const defaultHTTPSClient = require('https');

{{ $ctx := . }}
{{ range .Services }}
/**
 * Exposes all of the standard operations for the remote {{ .Name }} service. These RPC calls will be
 * sent over http(s) to the backend service instances. 
 */
class {{ .Name }}Client {
    baseURL;
    httpClient;

    constructor(baseURL, {pathPrefix, http, https} = {}) {
        this.baseURL = new URL(baseURL);
        this.pathPrefix = forceLeadingSlash(pathPrefix) || '';
        this.httpClient = this.baseURL.protocol === 'https:'
            ? https || defaultHTTPSClient 
            : http || defaultHTTPClient;
    }

    {{ $service := . }}
    {{ range .Methods }}
    /**{{ range .Documentation }}
     * {{ . }} {{ end }}
     *
     * @param { {{ .Request.Name }} } serviceRequest The input parameters
     * @returns {Promise<{{ .Response.Name }}>} The JSON-encoded return value of the operation.
     */
    async {{ .Name }}(serviceRequest) {
        return new Promise((resolve, reject) => {
            {{ if .HTTPMethod | HTTPMethodSupportsBody }}
            const bodyJSON = JSON.stringify(serviceRequest);
            {{ end }}
            const options = {
                protocol: this.baseURL.protocol,
                hostname: this.baseURL.hostname,
                port: this.baseURL.port,
                path: buildRequestPath('{{ .HTTPMethod }}', this.pathPrefix + '{{ .HTTPPath }}', serviceRequest),
                method: '{{ .HTTPMethod }}',
                headers: {
                    'Content-Type': 'application/json; charset=utf-8',
                    {{ if .HTTPMethod | HTTPMethodSupportsBody }}
                    'Content-Length': bodyJSON.length,
                    {{ end }}
                }
            };
            const req = this.httpClient.request(options, (res) => {
                let responseData = '';
                res.on('data', (chunk) => {
                    responseData += chunk;
                });
                res.on('end', () => {
                    handleResponse({res, responseData, resolve, reject});
                });
            });
            req.on('error', reject);
            {{ if .HTTPMethod | HTTPMethodSupportsBody }}
            req.write(bodyJSON);
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
 * @param {string} method The HTTP method for this request (determines if we include a query string)
 * @param {string} path The path pattern to populate w/ runtime values (e.g. "/user/:id")
 * @param {Object} serviceRequest The input struct for the service call
 * @returns {string} The fully-populate URL path (e.g. "/user/aCx31s")
 */
function buildRequestPath(method, path, serviceRequest) {
    const pathSegments = path.split("/").map(segment => {
        return segment.startsWith(":")
            ? attributeValue(serviceRequest, segment.substring(1))
            : segment;
    });
    let resolvedPath = pathSegments.join("/");
    if (!supportsBody(method)) {
        resolvedPath += '?';
        for (const attr in serviceRequest) {
            resolvedPath += attr + '=' + encodeURLParam(serviceRequest[attr]) + '&';
        }
    }
    return resolvedPath;
}

function encodeURLParam(value) {
    if (!value) {
        return '';
    }
    switch (typeof value) {
    case 'string':
    case 'number':
    case 'boolean':
        return encodeURIComponent(value);
    case 'function':
        return encodeURLParam(value());
    default:
        const valueJSON = JSON.stringify(value);
        return encodeURIComponent(valueJSON);
    }
}

/**
 * Given a struct-style object, return the values of the matching attribute. This is meant
 * to match the server's loose matching where the attribute name "ID" will match the
 * field "id".
 *
 * @param {Object} struct The data structure whose value you're trying to peel off.
 * @param {string} attributeName The name of the input value you're trying to send
 * @returns {*}
 */
function attributeValue(struct, attributeName) {
    const normalized = attributeName.toLowerCase();
    for (const key in struct) {
        if (key.toLowerCase() === normalized) {
            return encodeURLParam(struct[key]);
        }
    }
    return null;
}

/**
 * Takes a URL path (or prefix) and ensures that it starts with a leading "/" character.
 *
 * @param {string} urlPath The path to normalize
 * @returns {string}
 */
function forceLeadingSlash(urlPath) {
    if (!urlPath) {
        return '';
    }
    if (urlPath.startsWith('/')) {
        return urlPath
    }
    return '/' + urlPath;
}

/**
 * Accepts the full response data and the request's promise resolve/reject and determines
 * which to invoke. This will also JSON-unmarshal the response data if need be.
 */
function handleResponse({resolve, reject, res, responseData}) {
    const contentType = res.headers['content-type'];
    const responseValue = contentType.startsWith('application/json')
        ? JSON.parse(responseData)
        : responseData;

    return res.statusCode >= 400
        ? reject(responseValue)
        : resolve(responseValue);
}

/**
 * Does the HTTP method given support supplying data in the body of the request? For instance this is
 * true for POST but not for GET.
 *
 * @param {string} method The HTTP method that you are processing (e.g. "GET", "POST", etc) 
 * @returns {boolean}
 */
function supportsBody(method) {
    return method === 'POST' || method === 'PUT' || method === 'PATCH';
}

{{ range .Models }}
/**
 * @typedef {Object|*} {{ .Name }}
 */
{{ end }}

module.exports = {
    {{ range .Services }}{{ .Name }}Client,{{ end }}
};

/*
(async () => {
    console.info(">>>>> Making Call");
    const client = new GroupServiceClient("http://localhost:8080", { pathPrefix: 'v2'});
    const resultA = await client.GetByID({ id: "abc xyz", flag: true, moo: "cow", dog: {reason: 'i like pugs'}});
    console.info(">>>>>> RESULT A: ", resultA);

    const resultB = await client.CreateGroup({ name: 'Dogs', description: 'Woof woof, baby.' });
    console.info(">>>>>> RESULT B: ", resultB);
})().catch(err => {
    console.info(">>>>> ERROR...");
    console.error(err);
});
*/
`)
