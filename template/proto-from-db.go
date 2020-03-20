package template

// TmplProtoFromDb for template from db
var TmplProtoFromDb = `syntax = "proto3";

option go_package = "{{ .Name }}";
package {{ .Name }};

import "google/api/annotations.proto";
import "github.com/zokypesch/proto-lib/proto/options.proto";
import "google/protobuf/timestamp.proto";

service {{ .Name }} {{ unescape "{" }}
{{- range $table := .Tables }}
	rpc Create{{ ucfirst $table.Name }}(Create{{ ucfirst $table.Name }}Request) returns({{ ucfirst $table.Name }}) {
		option (google.api.http) = {
			post: "/api/v1/{{ $table.DBName }}/{{ $table.NameOriginal }}",
			body: "*"
		};
		option(httpMode) = "post";
		option(agregator) = "{{ ucfirst $table.Name }}.Create";
	};

	rpc GetAll{{ ucfirst $table.Name }}(GetAll{{ ucfirst $table.Name }}Request) returns(GetAll{{ ucfirst $table.Name }}Response) {
		option (google.api.http) = {
			get: "/api/v1/{{ $table.DBName }}/{{ $table.NameOriginal }}"
		};
		option(httpMode) = "get";
		option(agregator) = "{{ ucfirst $table.Name }}.GetAll";
	};

	rpc Update{{ ucfirst $table.Name }}(Update{{ ucfirst $table.Name }}Request) returns({{ ucfirst $table.Name }}) {
		option (google.api.http) = {
			put: "/api/v1/{{ $table.DBName }}/{{ $table.NameOriginal }}/{{ unescape "{"}}{{ $table.PrimaryKeyName }}{{ unescape "}"}}",
			body: "*"
		};
		option(httpMode) = "put";
		option(agregator) = "{{ ucfirst $table.Name }}.Update";
	};

	rpc Delete{{ ucfirst $table.Name }}(GetByIdRequest) returns(DeleteResponse) {
		option (google.api.http) = {
			delete: "/api/v1/{{ $table.DBName }}/{{ $table.NameOriginal }}/{{ unescape "{"}}{{ $table.PrimaryKeyName }}{{ unescape "}"}}"
		};
		option(httpMode) = "delete";
		option(agregator) = "{{ ucfirst $table.Name }}.Delete";
	};

	rpc GetById{{ ucfirst $table.Name }}(GetByIdRequest) returns({{ ucfirst $table.Name }}) {
		option (google.api.http) = {
			get: "/api/v1/{{ $table.DBName }}/{{ $table.NameOriginal }}/{{ unescape "{"}}{{ $table.PrimaryKeyName }}{{ unescape "}"}}"
		};
		option(httpMode) = "get";
		option(agregator) = "{{ ucfirst $table.Name }}.GetBy{{ ucfirst $table.PrimaryKeyName }}";
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
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.RequestPosition }} {{ unescape $field.Option }}
{{- end}}
{{- end}}
{{ unescape "}" }}

message Update{{ ucfirst $table.Name }}Request {{ unescape "{" }}
{{- range $field := $table.Fields }}
{{- if allowRequestWithId $field.Name }}
{{- if $field.PrimaryKey }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.RequestUpdatePosition }} [(required) = true,(required_type)="required",json_name="{{ $field.NameProto }}"];
{{- else }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.RequestUpdatePosition }} {{ unescape $field.OptionJSON }}
{{- end}}
{{- end }}
{{- end}}
{{ unescape "}" }}

message GetAll{{ ucfirst $table.Name }}Request {{ unescape "{" }}
{{- range $field := $table.Fields }}
{{- if allowRequest $field.Name }}
	{{ $field.DataTypeProto}} {{ $field.Name}} = {{ $field.RequestPosition }};
{{- end}}
{{- end}}
	int64 page = {{ $table.GetAll.Page }} [json_name="page"];
	int64 per_page = {{ $table.GetAll.PerPage }} [json_name="perPage"];
{{ unescape "}" }}

message GetAll{{ ucfirst $table.Name }}Response {{ unescape "{" }}
	repeated {{ ucfirst $table.Name }} items = 1 [json_name = "items"];
	int64 total = 2 [json_name = "total"];
	int64 page = 3 [json_name = "page"];
	int64 per_page = 4 [json_name = "perPage"];
{{ unescape "}" }}

{{- end}}

message GetByIdRequest {{ unescape "{" }}
	int64 id = 1 [(required) = true,(required_type)="required"];
{{ unescape "}" }}

message DeleteResponse {{ unescape "{" }}
	int64 id = 1;
{{ unescape "}" }}
`
