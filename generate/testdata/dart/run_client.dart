import 'dart:convert';
import 'dart:io';
import '../../../example/names/gen/name_service.gen.client.dart';

main(List<String> args) async {
  if (args.length < 1) {
    print('USAGE:');
    print('  dart run_client.dart {TEST_CASE}');
    exit(1);
  }
  await runTestCase(args[0]);
  exit(0);
}

runTestCase(String name) async {
  switch (name) {
    case "NotConnected":
      return testNotConnected();
    case "Success":
      return testSuccess();
    case "SuccessRaw":
      return testSuccessRaw();
    case "SuccessRawHeaders":
      return testSuccessRawHeaders();
    case "ValidationFailure":
      return testValidationFailure();
    case "AuthFailureClient":
      return testAuthFailureClient();
    case "AuthFailureCall":
      return testAuthFailureCall();
    case "AuthFailureCallOverride":
      return testAuthFailureCallOverride();
    default:
      print('Unknown test case: "$name"');
      exit(1);
  }
}

testNotConnected() async {
  var client = new NameServiceClient("http://localhost:9999");
  await output(client.FirstName(new FirstNameRequest(Name: "Jeff Lebowski")));
  await output(client.Download(new DownloadRequest(Name: "Jeff Lebowski")));
}

testSuccess() async {
  var client = new NameServiceClient("http://localhost:9100");
  await output(client.Split(SplitRequest(Name: 'Jeff Lebowski')));
  await output(client.FirstName(FirstNameRequest(Name: 'Jeff Lebowski')));
  await output(client.LastName(LastNameRequest(Name: 'Jeff Lebowski')));
  await output(client.SortName(SortNameRequest(Name: 'Jeff Lebowski')));
  await output(client.SortName(SortNameRequest(Name: 'Dude')));
}

testSuccessRaw() async {
  var client = new NameServiceClient("http://localhost:9100");
  await outputRaw(client.Download(DownloadRequest(Name: 'Jeff Lebowski')));
}

testSuccessRawHeaders() async {
  var client = new NameServiceClient("http://localhost:9100");
  await outputRaw(client.DownloadExt(DownloadExtRequest(Name: 'Jeff Lebowski', Ext: 'csv')));
  await outputRaw(client.DownloadExt(DownloadExtRequest(Name: 'Jeff Lebowski', Ext: 'txt')));
  await outputRaw(client.DownloadExt(DownloadExtRequest(Name: 'Jeff Lebowski', Ext: 't"x"t')));
}

testValidationFailure() async {
  var client = new NameServiceClient("http://localhost:9100");
  await output(client.Split(SplitRequest(Name: '')));
  await output(client.Split(SplitRequest()));
  await output(client.FirstName(FirstNameRequest(Name: '')));
  await output(client.FirstName(FirstNameRequest()));
  await output(client.LastName(LastNameRequest(Name: '')));
  await output(client.LastName(LastNameRequest()));
  await output(client.SortName(SortNameRequest(Name: '')));
  await output(client.SortName(SortNameRequest()));
}

testAuthFailureClient() async {
  var client = NameServiceClient('http://localhost:9100', authorization: 'Donny');
  await output(client.Split(SplitRequest(Name: 'Dude')));
  await output(client.FirstName(FirstNameRequest(Name: 'Dude')));
  await output(client.LastName(LastNameRequest(Name: 'Dude')));
  await output(client.SortName(SortNameRequest(Name: 'Dude')));
}

testAuthFailureCall() async {
  var client = NameServiceClient('http://localhost:9100');
  await output(client.Split(SplitRequest(Name: 'Dude'), authorization: 'Donny'));
  await output(client.FirstName(FirstNameRequest(Name: 'Dude'), authorization: 'Donny'));
  await output(client.LastName(LastNameRequest(Name: 'Dude'), authorization: 'Donny'));
  await output(client.SortName(SortNameRequest(Name: 'Dude'), authorization: 'Donny'));
}

testAuthFailureCallOverride() async {
  var client = NameServiceClient('http://localhost:9100', authorization: 'Donny');
  await output(client.Split(SplitRequest(Name: 'Dude'), authorization: 'ok'));
  await output(client.FirstName(FirstNameRequest(Name: 'Dude'), authorization: 'ok'));
  await output(client.LastName(LastNameRequest(Name: 'Dude'), authorization: 'ok'));
  await output(client.SortName(SortNameRequest(Name: 'Dude'), authorization: 'ok'));
}

output(Future<NameServiceModelJSON> model) async {
  try {
    var jsonString = jsonEncode(await model);
    print('OK ${jsonString}');
  }
  on NameServiceException catch (err) {
    var message = err.message.replaceAll('"', '\'');
    print('FAIL {"status":${err.status}, "message": "${message}"}');
  }
  catch (err) {
    print('FAIL {"message": "$err"}');
  }
}

outputRaw(Future<NameServiceModelJSON> model) async {
  try {
    var modelJson = (await model).toJson();
    modelJson['Content'] = await modelJson['Content'];
    print('OK ${jsonEncode(modelJson)}');
  }
  on NameServiceException catch (err) {
    var message = err.message.replaceAll('"', '\'');
    print('FAIL {"status":${err.status}, "message": "${message}"}');
  }
  catch (err) {
    print('FAIL {"message": "$err"}');
  }
}
