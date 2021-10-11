# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chameleon/api/http/data.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='chameleon/api/http/data.proto',
  package='chameleon.api.http',
  syntax='proto3',
  serialized_options=b'Z\027chameleon/api/http;http',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x1d\x63hameleon/api/http/data.proto\x12\x12\x63hameleon.api.http\"\x1b\n\tListValue\x12\x0e\n\x06values\x18\x01 \x03(\t\"\x97\x01\n\x06\x43ookie\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t\x12\x0e\n\x06\x64omain\x18\x03 \x01(\t\x12\x0c\n\x04path\x18\x04 \x01(\t\x12\x0f\n\x07\x65xpires\x18\x05 \x01(\x03\x12\x0c\n\x04size\x18\x06 \x01(\x05\x12\x10\n\x08httpOnly\x18\x07 \x01(\x08\x12\x0f\n\x07session\x18\x08 \x01(\x08\x12\x10\n\x08sameSite\x18\t \x01(\x05*G\n\x06Method\x12\x07\n\x03GET\x10\x00\x12\x08\n\x04POST\x10\x01\x12\x07\n\x03PUT\x10\x02\x12\t\n\x05PATCH\x10\x03\x12\n\n\x06\x44\x45LETE\x10\x04\x12\n\n\x06OPTION\x10\x06\x42\x19Z\x17\x63hameleon/api/http;httpb\x06proto3'
)

_METHOD = _descriptor.EnumDescriptor(
  name='Method',
  full_name='chameleon.api.http.Method',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='GET', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='POST', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PUT', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='PATCH', index=3, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DELETE', index=4, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='OPTION', index=5, number=6,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=236,
  serialized_end=307,
)
_sym_db.RegisterEnumDescriptor(_METHOD)

Method = enum_type_wrapper.EnumTypeWrapper(_METHOD)
GET = 0
POST = 1
PUT = 2
PATCH = 3
DELETE = 4
OPTION = 6



_LISTVALUE = _descriptor.Descriptor(
  name='ListValue',
  full_name='chameleon.api.http.ListValue',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='values', full_name='chameleon.api.http.ListValue.values', index=0,
      number=1, type=9, cpp_type=9, label=3,
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
  serialized_start=53,
  serialized_end=80,
)


_COOKIE = _descriptor.Descriptor(
  name='Cookie',
  full_name='chameleon.api.http.Cookie',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='chameleon.api.http.Cookie.name', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='chameleon.api.http.Cookie.value', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='domain', full_name='chameleon.api.http.Cookie.domain', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='path', full_name='chameleon.api.http.Cookie.path', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='expires', full_name='chameleon.api.http.Cookie.expires', index=4,
      number=5, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='size', full_name='chameleon.api.http.Cookie.size', index=5,
      number=6, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='httpOnly', full_name='chameleon.api.http.Cookie.httpOnly', index=6,
      number=7, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='session', full_name='chameleon.api.http.Cookie.session', index=7,
      number=8, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='sameSite', full_name='chameleon.api.http.Cookie.sameSite', index=8,
      number=9, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
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
  serialized_start=83,
  serialized_end=234,
)

DESCRIPTOR.message_types_by_name['ListValue'] = _LISTVALUE
DESCRIPTOR.message_types_by_name['Cookie'] = _COOKIE
DESCRIPTOR.enum_types_by_name['Method'] = _METHOD
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

ListValue = _reflection.GeneratedProtocolMessageType('ListValue', (_message.Message,), {
  'DESCRIPTOR' : _LISTVALUE,
  '__module__' : 'chameleon.api.http.data_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.api.http.ListValue)
  })
_sym_db.RegisterMessage(ListValue)

Cookie = _reflection.GeneratedProtocolMessageType('Cookie', (_message.Message,), {
  'DESCRIPTOR' : _COOKIE,
  '__module__' : 'chameleon.api.http.data_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.api.http.Cookie)
  })
_sym_db.RegisterMessage(Cookie)


DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)