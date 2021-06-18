# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chameleon/smelter/v1/crawl/proxy/data.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from chameleon.api.http import data_pb2 as chameleon_dot_api_dot_http_dot_data__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='chameleon/smelter/v1/crawl/proxy/data.proto',
  package='chameleon.smelter.v1.crawl.proxy',
  syntax='proto3',
  serialized_options=b'Z&chameleon/smelter/v1/crawl/proxy;proxy',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n+chameleon/smelter/v1/crawl/proxy/data.proto\x12 chameleon.smelter.v1.crawl.proxy\x1a\x1d\x63hameleon/api/http/data.proto\"\xb1\x05\n\x07Request\x12\x11\n\ttracingId\x18\x01 \x01(\t\x12\r\n\x05jobId\x18\x02 \x01(\t\x12\r\n\x05reqId\x18\x03 \x01(\t\x12\x0e\n\x06method\x18\x06 \x01(\t\x12\x0b\n\x03url\x18\x07 \x01(\t\x12G\n\x07headers\x18\x08 \x03(\x0b\x32\x36.chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry\x12\x0c\n\x04\x62ody\x18\t \x01(\x0c\x12\x42\n\x07options\x18\x0b \x01(\x0b\x32\x31.chameleon.smelter.v1.crawl.proxy.Request.Options\x12<\n\x08response\x18\x0f \x01(\x0b\x32*.chameleon.smelter.v1.crawl.proxy.Response\x1a\xaf\x02\n\x07Options\x12\x13\n\x0b\x65nableProxy\x18\x01 \x01(\x08\x12G\n\x0breliability\x18\x02 \x01(\x0e\x32\x32.chameleon.smelter.v1.crawl.proxy.ProxyReliability\x12\x16\n\x0e\x65nableHeadless\x18\x03 \x01(\x08\x12\x16\n\x0ejsWaitDuration\x18\x04 \x01(\x03\x12\x19\n\x11\x65nableSessionInit\x18\x05 \x01(\x08\x12\x13\n\x0bkeepSession\x18\x06 \x01(\x08\x12\x18\n\x10\x64isableCookieJar\x18\x07 \x01(\x08\x12\x18\n\x10maxTtlPerRequest\x18\x08 \x01(\x03\x12\x17\n\x0f\x64isableRedirect\x18\x0b \x01(\x08\x12\x19\n\x11requestFilterKeys\x18\x0f \x03(\t\x1aM\n\x0cHeadersEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12,\n\x05value\x18\x02 \x01(\x0b\x32\x1d.chameleon.api.http.ListValue:\x02\x38\x01\"\x8a\x03\n\x08Response\x12\x12\n\nstatusCode\x18\x01 \x01(\x05\x12\x0e\n\x06status\x18\x02 \x01(\t\x12\r\n\x05proto\x18\x03 \x01(\t\x12\x12\n\nprotoMajor\x18\x04 \x01(\x05\x12\x12\n\nprotoMinor\x18\x05 \x01(\x05\x12H\n\x07headers\x18\x06 \x03(\x0b\x32\x37.chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry\x12\x0c\n\x04\x62ody\x18\t \x01(\x0c\x12\x15\n\rbodyCacheLink\x18\x0b \x01(\t\x12\x10\n\x08\x64uration\x18\x0c \x01(\x03\x12\x17\n\x0f\x61verageDuration\x18\r \x01(\x03\x12:\n\x07request\x18\x0f \x01(\x0b\x32).chameleon.smelter.v1.crawl.proxy.Request\x1aM\n\x0cHeadersEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12,\n\x05value\x18\x02 \x01(\x0b\x32\x1d.chameleon.api.http.ListValue:\x02\x38\x01\"\xec\x01\n\x0bRequestWrap\x12\r\n\x05reqId\x18\x01 \x01(\t\x12:\n\x07request\x18\x02 \x01(\x0b\x32).chameleon.smelter.v1.crawl.proxy.Request\x12\x10\n\x08\x64\x65\x61\x64line\x18\x06 \x01(\x03\x12\x11\n\texecCount\x18\x07 \x01(\x05\x12\x10\n\x08\x64uration\x18\x0b \x01(\x03\x12\x17\n\x0f\x61verageDuration\x18\x0c \x01(\x03\x12\x42\n\x07options\x18\x0f \x01(\x0b\x32\x31.chameleon.smelter.v1.crawl.proxy.Request.Options*\x9f\x01\n\x10ProxyReliability\x12\x16\n\x12ReliabilityDefault\x10\x00\x12\x12\n\x0eReliabilityLow\x10\x01\x12\x15\n\x11ReliabilityMedium\x10\x02\x12\x13\n\x0fReliabilityHigh\x10\x03\x12\x17\n\x13ReliabilityRealtime\x10\t\x12\x1a\n\x16ReliabilityIntelligent\x10\nB(Z&chameleon/smelter/v1/crawl/proxy;proxyb\x06proto3'
  ,
  dependencies=[chameleon_dot_api_dot_http_dot_data__pb2.DESCRIPTOR,])

_PROXYRELIABILITY = _descriptor.EnumDescriptor(
  name='ProxyReliability',
  full_name='chameleon.smelter.v1.crawl.proxy.ProxyReliability',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='ReliabilityDefault', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ReliabilityLow', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ReliabilityMedium', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ReliabilityHigh', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ReliabilityRealtime', index=4, number=9,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='ReliabilityIntelligent', index=5, number=10,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=1441,
  serialized_end=1600,
)
_sym_db.RegisterEnumDescriptor(_PROXYRELIABILITY)

ProxyReliability = enum_type_wrapper.EnumTypeWrapper(_PROXYRELIABILITY)
ReliabilityDefault = 0
ReliabilityLow = 1
ReliabilityMedium = 2
ReliabilityHigh = 3
ReliabilityRealtime = 9
ReliabilityIntelligent = 10



_REQUEST_OPTIONS = _descriptor.Descriptor(
  name='Options',
  full_name='chameleon.smelter.v1.crawl.proxy.Request.Options',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='enableProxy', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.enableProxy', index=0,
      number=1, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='reliability', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.reliability', index=1,
      number=2, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='enableHeadless', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.enableHeadless', index=2,
      number=3, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='jsWaitDuration', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.jsWaitDuration', index=3,
      number=4, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='enableSessionInit', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.enableSessionInit', index=4,
      number=5, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='keepSession', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.keepSession', index=5,
      number=6, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='disableCookieJar', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.disableCookieJar', index=6,
      number=7, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='maxTtlPerRequest', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.maxTtlPerRequest', index=7,
      number=8, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='disableRedirect', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.disableRedirect', index=8,
      number=11, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='requestFilterKeys', full_name='chameleon.smelter.v1.crawl.proxy.Request.Options.requestFilterKeys', index=9,
      number=15, type=9, cpp_type=9, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=420,
  serialized_end=723,
)

_REQUEST_HEADERSENTRY = _descriptor.Descriptor(
  name='HeadersEntry',
  full_name='chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=b'8\001',
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=725,
  serialized_end=802,
)

_REQUEST = _descriptor.Descriptor(
  name='Request',
  full_name='chameleon.smelter.v1.crawl.proxy.Request',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='tracingId', full_name='chameleon.smelter.v1.crawl.proxy.Request.tracingId', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='jobId', full_name='chameleon.smelter.v1.crawl.proxy.Request.jobId', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='reqId', full_name='chameleon.smelter.v1.crawl.proxy.Request.reqId', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='method', full_name='chameleon.smelter.v1.crawl.proxy.Request.method', index=3,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='url', full_name='chameleon.smelter.v1.crawl.proxy.Request.url', index=4,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='headers', full_name='chameleon.smelter.v1.crawl.proxy.Request.headers', index=5,
      number=8, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='body', full_name='chameleon.smelter.v1.crawl.proxy.Request.body', index=6,
      number=9, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=b"",
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='options', full_name='chameleon.smelter.v1.crawl.proxy.Request.options', index=7,
      number=11, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='response', full_name='chameleon.smelter.v1.crawl.proxy.Request.response', index=8,
      number=15, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_REQUEST_OPTIONS, _REQUEST_HEADERSENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=113,
  serialized_end=802,
)


_RESPONSE_HEADERSENTRY = _descriptor.Descriptor(
  name='HeadersEntry',
  full_name='chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry.value', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=b'8\001',
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=725,
  serialized_end=802,
)

_RESPONSE = _descriptor.Descriptor(
  name='Response',
  full_name='chameleon.smelter.v1.crawl.proxy.Response',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='statusCode', full_name='chameleon.smelter.v1.crawl.proxy.Response.statusCode', index=0,
      number=1, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='status', full_name='chameleon.smelter.v1.crawl.proxy.Response.status', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='proto', full_name='chameleon.smelter.v1.crawl.proxy.Response.proto', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='protoMajor', full_name='chameleon.smelter.v1.crawl.proxy.Response.protoMajor', index=3,
      number=4, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='protoMinor', full_name='chameleon.smelter.v1.crawl.proxy.Response.protoMinor', index=4,
      number=5, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='headers', full_name='chameleon.smelter.v1.crawl.proxy.Response.headers', index=5,
      number=6, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='body', full_name='chameleon.smelter.v1.crawl.proxy.Response.body', index=6,
      number=9, type=12, cpp_type=9, label=1,
      has_default_value=False, default_value=b"",
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='bodyCacheLink', full_name='chameleon.smelter.v1.crawl.proxy.Response.bodyCacheLink', index=7,
      number=11, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='duration', full_name='chameleon.smelter.v1.crawl.proxy.Response.duration', index=8,
      number=12, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='averageDuration', full_name='chameleon.smelter.v1.crawl.proxy.Response.averageDuration', index=9,
      number=13, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='request', full_name='chameleon.smelter.v1.crawl.proxy.Response.request', index=10,
      number=15, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_RESPONSE_HEADERSENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=805,
  serialized_end=1199,
)


_REQUESTWRAP = _descriptor.Descriptor(
  name='RequestWrap',
  full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='reqId', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.reqId', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='request', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.request', index=1,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='deadline', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.deadline', index=2,
      number=6, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='execCount', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.execCount', index=3,
      number=7, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='duration', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.duration', index=4,
      number=11, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='averageDuration', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.averageDuration', index=5,
      number=12, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='options', full_name='chameleon.smelter.v1.crawl.proxy.RequestWrap.options', index=6,
      number=15, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1202,
  serialized_end=1438,
)

_REQUEST_OPTIONS.fields_by_name['reliability'].enum_type = _PROXYRELIABILITY
_REQUEST_OPTIONS.containing_type = _REQUEST
_REQUEST_HEADERSENTRY.fields_by_name['value'].message_type = chameleon_dot_api_dot_http_dot_data__pb2._LISTVALUE
_REQUEST_HEADERSENTRY.containing_type = _REQUEST
_REQUEST.fields_by_name['headers'].message_type = _REQUEST_HEADERSENTRY
_REQUEST.fields_by_name['options'].message_type = _REQUEST_OPTIONS
_REQUEST.fields_by_name['response'].message_type = _RESPONSE
_RESPONSE_HEADERSENTRY.fields_by_name['value'].message_type = chameleon_dot_api_dot_http_dot_data__pb2._LISTVALUE
_RESPONSE_HEADERSENTRY.containing_type = _RESPONSE
_RESPONSE.fields_by_name['headers'].message_type = _RESPONSE_HEADERSENTRY
_RESPONSE.fields_by_name['request'].message_type = _REQUEST
_REQUESTWRAP.fields_by_name['request'].message_type = _REQUEST
_REQUESTWRAP.fields_by_name['options'].message_type = _REQUEST_OPTIONS
DESCRIPTOR.message_types_by_name['Request'] = _REQUEST
DESCRIPTOR.message_types_by_name['Response'] = _RESPONSE
DESCRIPTOR.message_types_by_name['RequestWrap'] = _REQUESTWRAP
DESCRIPTOR.enum_types_by_name['ProxyReliability'] = _PROXYRELIABILITY
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Request = _reflection.GeneratedProtocolMessageType('Request', (_message.Message,), {

  'Options' : _reflection.GeneratedProtocolMessageType('Options', (_message.Message,), {
    'DESCRIPTOR' : _REQUEST_OPTIONS,
    '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.Request.Options)
    })
  ,

  'HeadersEntry' : _reflection.GeneratedProtocolMessageType('HeadersEntry', (_message.Message,), {
    'DESCRIPTOR' : _REQUEST_HEADERSENTRY,
    '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.Request.HeadersEntry)
    })
  ,
  'DESCRIPTOR' : _REQUEST,
  '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.Request)
  })
_sym_db.RegisterMessage(Request)
_sym_db.RegisterMessage(Request.Options)
_sym_db.RegisterMessage(Request.HeadersEntry)

Response = _reflection.GeneratedProtocolMessageType('Response', (_message.Message,), {

  'HeadersEntry' : _reflection.GeneratedProtocolMessageType('HeadersEntry', (_message.Message,), {
    'DESCRIPTOR' : _RESPONSE_HEADERSENTRY,
    '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.Response.HeadersEntry)
    })
  ,
  'DESCRIPTOR' : _RESPONSE,
  '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.Response)
  })
_sym_db.RegisterMessage(Response)
_sym_db.RegisterMessage(Response.HeadersEntry)

RequestWrap = _reflection.GeneratedProtocolMessageType('RequestWrap', (_message.Message,), {
  'DESCRIPTOR' : _REQUESTWRAP,
  '__module__' : 'chameleon.smelter.v1.crawl.proxy.data_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.proxy.RequestWrap)
  })
_sym_db.RegisterMessage(RequestWrap)


DESCRIPTOR._options = None
_REQUEST_HEADERSENTRY._options = None
_RESPONSE_HEADERSENTRY._options = None
# @@protoc_insertion_point(module_scope)
