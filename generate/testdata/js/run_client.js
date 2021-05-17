/* global require,process */
const fetch = require('./node_modules/node-fetch/lib/index.js');
const {NameServiceClient} = require('../../../example/names/gen/name_service.gen.client.js');

async function main() {
    const suite = new TestSuite();
    const testFunctionName = 'test' + process.argv[2];
    return suite[testFunctionName]();
}

class TestSuite {
    async testNotConnected() {
        const client = new NameServiceClient('http://localhost:9999', {fetch});
        await output(client.Split({ Name: 'Jeff Lebowski' }));
        await output(client.Download({ Name: 'Jeff Lebowski' }));
    }

    async testBadFetch() {
        const client = new NameServiceClient('http://localhost:9100', {fetch: {}});
        await output(client.Split({ Name: 'Jeff Lebowski' }));
    }

    async testSuccess() {
        const client = new NameServiceClient('http://localhost:9100', {fetch});
        await output(client.Split({ Name: 'Jeff Lebowski' }));
        await output(client.FirstName({ Name: 'Jeff Lebowski' }));
        await output(client.LastName({ Name: 'Jeff Lebowski' }));
        await output(client.SortName({ Name: 'Jeff Lebowski' }));
        await output(client.SortName({ Name: 'Dude' }));
    }

    async testSuccessRaw() {
        const client = new NameServiceClient('http://localhost:9100', {fetch});
        await outputRaw(client.Download({ Name: 'Jeff Lebowski' }));
    }

    async testSuccessRawHeaders() {
        const client = new NameServiceClient('http://localhost:9100', {fetch});
        await outputRaw(client.DownloadExt({ Name: 'Jeff Lebowski', Ext: 'csv' }));
        await outputRaw(client.DownloadExt({ Name: 'Jeff Lebowski', Ext: 'txt' }));
        await outputRaw(client.DownloadExt({ Name: 'Jeff Lebowski', Ext: 't"x"t' }));
    }

    async testValidationFailure() {
        const client = new NameServiceClient('http://localhost:9100', {fetch});
        await output(client.Split({ Name: '' }));
        await output(client.Split({ }));
        await output(client.FirstName({ Name: '' }));
        await output(client.FirstName({ }));
        await output(client.LastName({ Name: '' }));
        await output(client.LastName({ }));
        await output(client.SortName({ Name: '' }));
        await output(client.SortName({ }));

        // Raw failures should be output as JSON, too.
        await outputRaw(client.Download({ }));
        await outputRaw(client.DownloadExt({ }));
    }

    async testAuthFailureClient() {
        const client = new NameServiceClient('http://localhost:9100', {fetch, authorization: 'Donny'});
        await output(client.Split({ Name: 'Dude' }));
        await output(client.FirstName({ Name: 'Dude' }));
        await output(client.LastName({ Name: 'Dude' }));
        await output(client.SortName({ Name: 'Dude' }));
    }

    async testAuthFailureCall() {
        const client = new NameServiceClient('http://localhost:9100', {fetch});
        await output(client.Split({Name: 'Dude'}, {authorization: 'Donny'}));
        await output(client.FirstName({Name: 'Dude'}, {authorization: 'Donny'}));
        await output(client.LastName({Name: 'Dude'}, {authorization: 'Donny'}));
        await output(client.SortName({Name: 'Dude'}, {authorization: 'Donny'}));
    }

    async testAuthFailureCallOverride() {
        const client = new NameServiceClient('http://localhost:9100', {fetch, authorization: 'Donny'});
        await output(client.Split({Name: 'Dude'}, {authorization: 'ok'}));
        await output(client.FirstName({Name: 'Dude'}, {authorization: 'ok'}));
        await output(client.LastName({Name: 'Dude'}, {authorization: 'ok'}));
        await output(client.SortName({Name: 'Dude'}, {authorization: 'ok'}));
    }
}

async function outputRaw(value) {
    return output(value, true);
}

async function output(value, raw = false) {
    try {
        value = await value;
        if (raw) {
            value.Content = await value.Content.text();
        }
        console.info('OK ' + JSON.stringify(await value));
    }
    catch (e) {
        const failure = await e;
        const failureJSON = typeof failure === 'string'
            ? JSON.stringify({message: failure})
            : JSON.stringify(failure);

        console.info('FAIL ' + failureJSON);
    }
}

main()
    .then()
    .catch((e) => console.info('FAILURE:' + e));
