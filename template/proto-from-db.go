package template

// TmplProtoFromDb for template from db
var TmplProtoFromDb = `syntax = "proto3";

option go_package = "{{ .Name }}";
package {{ .Name }};

import "google/protobuf/empty.proto";
import "google/api/annotations.proto";
import "github.com/zokypesch/proto-lib/proto/options.proto";
import "google/protobuf/timestamp.proto";

service {{ .Name }} {{ unescape "{" }}
{{- range $table := .Tables }}
	rpc Create{{ ucfirst $table.Name }}(Create{{ ucfirst $table.Name }}Request) returns(General{{ ucfirst $table.Name }}Response) {
		option (google.api.http) = {
			post: "/v1/{{ $table.Name }}",
			body: "*"
		};
		option(httpMode) = "post";
		option(agregator) = "{{ ucfirst $table.Name }}.Create";
	};

	rpc GetAll{{ ucfirst $table.Name }}({{ ucfirst $table.Name }}) returns(GetAll{{ ucfirst $table.Name }}Response) {
		option (google.api.http) = {
			get: "/v1/{{ $table.Name }}"
		};
		option(httpMode) = "get";
		option(agregator) = "{{ ucfirst $table.Name }}.GetAll";
	};

	rpc Update{{ ucfirst $table.Name }}(Create{{ ucfirst $table.Name }}Request) returns(General{{ ucfirst $table.Name }}Response) {
		option (google.api.http) = {
			put: "/v1/{{ $table.Name }}",
			body: "*"
		};
		option(httpMode) = "put";
		option(agregator) = "{{ ucfirst $table.Name }}.Update";
	};

	rpc Delete{{ ucfirst $table.Name }}(GetByIdRequest) returns(DeleteResponse) {
		option (google.api.http) = {
			delete: "/v1/{{ $table.Name }}/{{ unescape "{"}}{{ $table.PrimaryKeyName }}{{ unescape "}"}}",
			body: "*"
		};
		option(httpMode) = "delete";
		option(agregator) = "{{ ucfirst $table.Name }}.Delete";
	};

	rpc GetById{{ ucfirst $table.Name }}(GetByIdRequest) returns(General{{ ucfirst $table.Name }}Response) {
		option (google.api.http) = {
			get: "/v1/{{ $table.Name }}/{{ unescape "{"}}{{ $table.PrimaryKeyName }}{{ unescape "}"}}"
		};
		option(httpMode) = "get";
		option(agregator) = "{{ ucfirst $table.Name }}.GetById";
	};
{{- end}}
{{ unescape "}" }}

// Table refference
{{- range $table := .Tables }}
message {{ ucfirst $table.Name }} {{ unescape "{" }}
	option (isRepo) = true;
{{- range $field := $table.Fields }}
{{- if $field.PrimaryKey }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.OrdinalPosition }} [(isPrimaryKey) = true];
{{- else}}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.OrdinalPosition }};
{{- end}}
{{- end}}
{{- range $join := $table.Joins }}
{{- if $join.Repeated }}
	repeated {{ $join.ReferencedTableName }} {{ $join.ReferencedTableOriginal }} = {{ $join.OrdinalPosition }} {{ unescape $join.Option }};
{{- else}}
	{{ $join.ReferencedTableName }} {{ $join.ReferencedTableOriginal }} = {{ $join.OrdinalPosition }} {{ unescape $join.Option }};
{{- end}}
{{- end}}
{{ unescape "}" }}
{{- end}}

// Table Request
{{- range $table := .Tables }}
message Create{{ ucfirst $table.Name }}Request {{ unescape "{" }}
{{- range $field := $table.Fields }}
{{- if allowRequest $field.Name }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.OrdinalPosition }} {{ unescape $field.Option }}
{{- end}}
{{- end}}
{{ unescape "}" }}

message General{{ ucfirst $table.Name }}Response {{ unescape "{" }}
{{- range $field := $table.Fields }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.OrdinalPosition }} [(json_name) = "{{ $field.NameProto }}"];
{{- end}}
{{- range $join := $table.Joins }}
{{- if $join.Repeated }}
	repeated {{ $join.ReferencedTableName }} {{ $join.ReferencedTableOriginal }} = {{ $join.OrdinalPosition }} [(json_name) = "{{ $join.ReferencedColumnNameProto }}"];
{{- else}}
	{{ $join.ReferencedTableName }} {{ $join.ReferencedTableOriginal }} = {{ $join.OrdinalPosition }} [(json_name) = "{{ $join.ReferencedColumnNameProto }}"];
{{- end}}
{{- end}}
{{ unescape "}" }}

message GetAll{{ ucfirst $table.Name }}Response {{ unescape "{" }}
	repeated General{{ ucfirst $table.Name }}Response items = 1 [(json_name) = "items"];
	int64 total = 2 [(json_name) = "total"];
	int64 page = 3 [(json_name) = "page"];
	int64 per_page = 4 [(json_name) = "perPage"];
	int64 last_page = 5 [(json_name) = "lastPage"];
{{ unescape "}" }}

{{- end}}

message GetByIdRequest {{ unescape "{" }}
	int64 id = 1;
{{ unescape "}" }}

message DeleteResponse {{ unescape "{" }}
	int64 id = 1;
{{ unescape "}" }}
`
