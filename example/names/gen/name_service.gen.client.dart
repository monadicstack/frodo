// Code generated by Frodo from example/names/name_service.go - DO NOT EDIT
//
//   https://github.com/monadicstack/frodo
//
import 'dart:async';
import 'dart:convert';
import 'dart:io';

class NameServiceClient {
  static const String pathPrefix = '';

  final String baseURL;
  String authorization;
  HttpClient httpClient = HttpClient();

  NameServiceClient(this.baseURL, {
      this.authorization = '',
  });

  
  /// Download returns a raw CSV file containing the parsed name.
  Future<DownloadResponse> Download(DownloadRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.Download';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponseRaw(httpResponse, (json) => DownloadResponse.fromJson(json));
  }
  
  /// DownloadExt returns a raw CSV file containing the parsed name. This differs from Download
  /// by giving you the "Ext" knob which will let you exercise the content type and disposition
  /// interfaces that Frodo supports for raw responses.
  Future<DownloadExtResponse> DownloadExt(DownloadExtRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.DownloadExt';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponseRaw(httpResponse, (json) => DownloadExtResponse.fromJson(json));
  }
  
  /// FirstName extracts just the first name from a full name string.
  Future<FirstNameResponse> FirstName(FirstNameRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.FirstName';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponse(httpResponse, (json) => FirstNameResponse.fromJson(json));
    
  }
  
  /// LastName extracts just the last name from a full name string.
  Future<LastNameResponse> LastName(LastNameRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.LastName';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponse(httpResponse, (json) => LastNameResponse.fromJson(json));
    
  }
  
  /// SortName establishes the "phone book" name for the given full name.
  Future<SortNameResponse> SortName(SortNameRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.SortName';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponse(httpResponse, (json) => SortNameResponse.fromJson(json));
    
  }
  
  /// Split separates a first and last name.
  Future<SplitResponse> Split(SplitRequest serviceRequest, {String authorization = ''}) async {
    var requestJson = serviceRequest.toJson();
    var method = 'POST';
    var route = '/NameService.Split';
    var url = _joinUrl([baseURL, pathPrefix, _buildRequestPath(method, route, requestJson)]);

    var httpRequest = await httpClient.openUrl(method, Uri.parse(url));
    httpRequest.headers.set('Accept', 'application/json');
    httpRequest.headers.set('Authorization', _authorize(authorization));
    httpRequest.headers.set('Content-Type', 'application/json');
    httpRequest.write(jsonEncode(requestJson));

    var httpResponse = await httpRequest.close();
    return _handleResponse(httpResponse, (json) => SplitResponse.fromJson(json));
    
  }
  

  String _buildRequestPath(String method, String route, Map<String, dynamic> requestJson) {
    String stringify(Map<String, dynamic> json, String key) {
      return Uri.encodeComponent(json[key]?.toString() ?? '');
    }
    String stringifyAndRemove(Map<String, dynamic> json, String key) {
      return Uri.encodeComponent(json.remove(key)?.toString() ?? '');
    }

    // Since we're embedding values in a path or query string, we need to flatten "{a: {b: {c: 4}}}"
    // down to "a.b.c=4" for it to fit nicely into our URL-based binding.
    requestJson = _flattenJson(requestJson);

    var resolvedPath = route
      .split('/')
      .map((s) => s.startsWith(':') ? stringifyAndRemove(requestJson, s.substring(1)) : s)
      .join('/');

    // These encode the data in the body, so no need to shove it in the query string.
    if (method == 'POST' || method == 'PUT' || method == 'PATCH') {
      return resolvedPath;
    }

    // GET/DELETE/etc will pass all values through the query string.
    var queryValues = requestJson.keys
      .map((key) => key + '=' + stringify(requestJson, key))
      .join('&');

    return resolvedPath + '?' + queryValues;
  }

  Future<T> _handleResponse<T>(HttpClientResponse httpResponse, T Function(Map<String, dynamic>) factory) async {
    if (httpResponse.statusCode >= 400) {
      throw await NameServiceException.fromResponse(httpResponse);
    }

    var bodyJson = await _streamToString(httpResponse);
    return factory(jsonDecode(bodyJson));
  }

  Future<T> _handleResponseRaw<T>(HttpClientResponse httpResponse, T Function(Map<String, dynamic>) factory) async {
    if (httpResponse.statusCode >= 400) {
      throw await NameServiceException.fromResponse(httpResponse);
    }

    return factory({
      'Content': httpResponse,
      'ContentType': httpResponse.headers.value('Content-Type') ?? 'application/octet-stream',
      'ContentFileName': _dispositionFileName(httpResponse.headers.value('Content-Disposition')),
    });
  }

  String _authorize(String callAuthorization) {
    return callAuthorization.trim().isNotEmpty
      ? callAuthorization
      : authorization;
  }

  String _joinUrl(List<String> segments) {
    String stripLeadingSlash(String s) {
      while (s.startsWith('/')) {
        s = s.substring(1);
      }
      return s;
    }
    String stripTrailingSlash(String s) {
      while (s.endsWith('/')) {
        s = s.substring(0, s.length - 1);
      }
      return s;
    }
    bool notEmpty(String s) {
      return s.isNotEmpty;
    }

    return segments
        .map(stripLeadingSlash)
        .map(stripTrailingSlash)
        .where(notEmpty)
        .join('/');
  }

  Map<String, dynamic> _flattenJson(Map<String, dynamic> json) {
    // Adds the given json map entry to the accumulator map. The 'path' contains
    // the period-delimited path for all parent objects we've recurred down from.
    void flattenEntry(String path, String key, dynamic value, Map<String, dynamic> accum) {
      if (value == null) {
        return;
      }

      path = path == '' ? key : '$path.$key';
      if (value is Map<String, dynamic>) {
        value.keys.forEach((key) => flattenEntry(path, key, value[key], accum));
        return;
      }
      accum[path] = value;
    }

    Map<String, dynamic> result = Map<String, dynamic>();
    json.keys.forEach((key) => flattenEntry("", key, json[key], result));
    return result;
  }

  String _dispositionFileName(String? contentDisposition) {
    if (contentDisposition == null) {
      return '';
    }

    var fileNameAttrPos = contentDisposition.indexOf("filename=");
    if (fileNameAttrPos < 0) {
      return '';
    }

    var fileName = contentDisposition.substring(fileNameAttrPos + 9);
    fileName = fileName.startsWith('"') ? fileName.substring(1) : fileName;
    fileName = fileName.endsWith('"') ? fileName.substring(0, fileName.length - 1) : fileName;
    fileName = fileName.replaceAll('\\"', '\"');
    return fileName;
  }
}

class NameServiceException implements Exception {
  int status;
  String message;

  NameServiceException(this.status, this.message);

  static Future<NameServiceException> fromResponse(HttpClientResponse response) async {
    var body = await _streamToString(response);
    var message = '';
    try {
      Map<String, dynamic> json = jsonDecode(body);
      message = json['message'] ?? json['error'] ?? body;
    }
    catch (_) {
      message = body;
    }
    throw new NameServiceException(response.statusCode, message);
  }
}


/// LastNameRequest is the output for the LastName function.
class LastNameRequest implements NameServiceModelJSON { 
  String? Name;

  LastNameRequest({ 
    this.Name,
  });

  LastNameRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}

/// LastNameResponse is the output for the LastName function.
class LastNameResponse implements NameServiceModelJSON { 
  String? LastName;

  LastNameResponse({ 
    this.LastName,
  });

  LastNameResponse.fromJson(Map<String, dynamic> json) { 
    LastName = json['LastName'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'LastName': LastName,
    };
  }
}

/// DownloadExtRequest is the input for the DownloadExt function.
class DownloadExtRequest implements NameServiceModelJSON { 
  String? Name;
  String? Ext;

  DownloadExtRequest({ 
    this.Name,
    this.Ext,
  });

  DownloadExtRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
    Ext = json['Ext'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
      'Ext': Ext,
    };
  }
}

/// DownloadResponse is the output for the Download function.
class DownloadResponse implements NameServiceModelJSON { 
  Stream<List<int>>? Content;
  String? ContentType;
  String? ContentFileName;

  DownloadResponse({ 
    this.Content,
    this.ContentType,
    this.ContentFileName,
    
  });

  DownloadResponse.fromJson(Map<String, dynamic> json) { 
    Content = json['Content'] as Stream<List<int>>?;
    ContentType = json['ContentType'] ?? 'application/octet-stream';
    ContentFileName = json['ContentFileName'] ?? '';
    
  }

  Map<String, dynamic> toJson() {
    return { 
      'Content': _streamToString(Content),
      'ContentType': ContentType ?? 'application/octet-stream',
      'ContentFileName': ContentFileName ?? '',
      
    };
  }
}

/// DownloadExtResponse is the output for the DownloadExt function.
class DownloadExtResponse implements NameServiceModelJSON { 
  Stream<List<int>>? Content;
  String? ContentType;
  String? ContentFileName;

  DownloadExtResponse({ 
    this.Content,
    this.ContentType,
    this.ContentFileName,
    
  });

  DownloadExtResponse.fromJson(Map<String, dynamic> json) { 
    Content = json['Content'] as Stream<List<int>>?;
    ContentType = json['ContentType'] ?? 'application/octet-stream';
    ContentFileName = json['ContentFileName'] ?? '';
    
  }

  Map<String, dynamic> toJson() {
    return { 
      'Content': _streamToString(Content),
      'ContentType': ContentType ?? 'application/octet-stream',
      'ContentFileName': ContentFileName ?? '',
      
    };
  }
}

/// FirstNameRequest is the input for the FirstName function.
class FirstNameRequest implements NameServiceModelJSON { 
  String? Name;

  FirstNameRequest({ 
    this.Name,
  });

  FirstNameRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}

/// SortNameRequest is the input for the SortName function.
class SortNameRequest implements NameServiceModelJSON { 
  String? Name;

  SortNameRequest({ 
    this.Name,
  });

  SortNameRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}

/// SplitResponse is the output for the Split function.
class SplitResponse implements NameServiceModelJSON { 
  String? FirstName;
  String? LastName;

  SplitResponse({ 
    this.FirstName,
    this.LastName,
  });

  SplitResponse.fromJson(Map<String, dynamic> json) { 
    FirstName = json['FirstName'];
    LastName = json['LastName'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'FirstName': FirstName,
      'LastName': LastName,
    };
  }
}

/// NameRequest generalizes the data we pass to any of the name service functions.
class NameRequest implements NameServiceModelJSON { 
  String? Name;

  NameRequest({ 
    this.Name,
  });

  NameRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}

/// FirstNameResponse is the output for the FirstName function.
class FirstNameResponse implements NameServiceModelJSON { 
  String? FirstName;

  FirstNameResponse({ 
    this.FirstName,
  });

  FirstNameResponse.fromJson(Map<String, dynamic> json) { 
    FirstName = json['FirstName'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'FirstName': FirstName,
    };
  }
}

/// SortNameResponse is the output for the SortName function.
class SortNameResponse implements NameServiceModelJSON { 
  String? SortName;

  SortNameResponse({ 
    this.SortName,
  });

  SortNameResponse.fromJson(Map<String, dynamic> json) { 
    SortName = json['SortName'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'SortName': SortName,
    };
  }
}

class SplitRequest implements NameServiceModelJSON { 
  String? Name;

  SplitRequest({ 
    this.Name,
  });

  SplitRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}

/// DownloadRequest is the input for the Download function.
class DownloadRequest implements NameServiceModelJSON { 
  String? Name;

  DownloadRequest({ 
    this.Name,
  });

  DownloadRequest.fromJson(Map<String, dynamic> json) { 
    Name = json['Name'];
  }

  Map<String, dynamic> toJson() {
    return { 
      'Name': Name,
    };
  }
}


class NameServiceModelJSON {
  Map<String, dynamic> toJson() {
    throw new Exception('toJson not implemented');
  }
}

List<T>? _map<T>(List<dynamic>? jsonList, T Function(dynamic) mapping) {
  return jsonList == null ? null : jsonList.map(mapping).toList();
}

Future<String> _streamToString(Stream<List<int>>? stream) async {
  if (stream == null) {
    return '';
  }
  var bodyCompleter = new Completer<String>();
  stream.transform(utf8.decoder).listen(bodyCompleter.complete);
  return bodyCompleter.future;
}
