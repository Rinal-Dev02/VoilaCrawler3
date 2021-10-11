# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: protobuf/options/auth.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor.FileDescriptor(
  name='protobuf/options/auth.proto',
  package='protobuf.options',
  syntax='proto3',
  serialized_options=b'Z\030protobuf/options;options',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n\x1bprotobuf/options/auth.proto\x12\x10protobuf.options\"\x85\x01\n\x08\x41uthRule\x12*\n\x05level\x18\x01 \x01(\x0e\x32\x1b.protobuf.options.AuthLevel\x12\r\n\x05scope\x18\x02 \x01(\t\x12)\n\x04verb\x18\x03 \x01(\x0e\x32\x1b.protobuf.options.ScopeVerb\x12\x13\n\x0brequireUser\x18\x06 \x01(\x08*=\n\tAuthLevel\x12\n\n\x06NoAuth\x10\x00\x12\x10\n\x0cOptionalAuth\x10\x01\x12\x12\n\x0eRestrictedAuth\x10\x02*B\n\tScopeVerb\x12\x08\n\x04\x41UTO\x10\x00\x12\x07\n\x03GET\x10\x01\x12\n\n\x06\x43REATE\x10\x03\x12\n\n\x06UPDATE\x10\x04\x12\n\n\x06\x44\x45LETE\x10\x06\x42\x1aZ\x18protobuf/options;optionsb\x06proto3'
)

_AUTHLEVEL = _descriptor.EnumDescriptor(
  name='AuthLevel',
  full_name='protobuf.options.AuthLevel',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='NoAuth', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='OptionalAuth', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='RestrictedAuth', index=2, number=2,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=185,
  serialized_end=246,
)
_sym_db.RegisterEnumDescriptor(_AUTHLEVEL)

AuthLevel = enum_type_wrapper.EnumTypeWrapper(_AUTHLEVEL)
_SCOPEVERB = _descriptor.EnumDescriptor(
  name='ScopeVerb',
  full_name='protobuf.options.ScopeVerb',
  filename=None,
  file=DESCRIPTOR,
  create_key=_descriptor._internal_create_key,
  values=[
    _descriptor.EnumValueDescriptor(
      name='AUTO', index=0, number=0,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='GET', index=1, number=1,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='CREATE', index=2, number=3,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='UPDATE', index=3, number=4,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
    _descriptor.EnumValueDescriptor(
      name='DELETE', index=4, number=6,
      serialized_options=None,
      type=None,
      create_key=_descriptor._internal_create_key),
  ],
  containing_type=None,
  serialized_options=None,
  serialized_start=248,
  serialized_end=314,
)
_sym_db.RegisterEnumDescriptor(_SCOPEVERB)

ScopeVerb = enum_type_wrapper.EnumTypeWrapper(_SCOPEVERB)
NoAuth = 0
OptionalAuth = 1
RestrictedAuth = 2
AUTO = 0
GET = 1
CREATE = 3
UPDATE = 4
DELETE = 6



_AUTHRULE = _descriptor.Descriptor(
  name='AuthRule',
  full_name='protobuf.options.AuthRule',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='level', full_name='protobuf.options.AuthRule.level', index=0,
      number=1, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='scope', full_name='protobuf.options.AuthRule.scope', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='verb', full_name='protobuf.options.AuthRule.verb', index=2,
      number=3, type=14, cpp_type=8, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='requireUser', full_name='protobuf.options.AuthRule.requireUser', index=3,
      number=6, type=8, cpp_type=7, label=1,
      has_default_value=False, default_value=False,
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
  serialized_start=50,
  serialized_end=183,
)

_AUTHRULE.fields_by_name['level'].enum_type = _AUTHLEVEL
_AUTHRULE.fields_by_name['verb'].enum_type = _SCOPEVERB
DESCRIPTOR.message_types_by_name['AuthRule'] = _AUTHRULE
DESCRIPTOR.enum_types_by_name['AuthLevel'] = _AUTHLEVEL
DESCRIPTOR.enum_types_by_name['ScopeVerb'] = _SCOPEVERB
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

AuthRule = _reflection.GeneratedProtocolMessageType('AuthRule', (_message.Message,), {
  'DESCRIPTOR' : _AUTHRULE,
  '__module__' : 'protobuf.options.auth_pb2'
  # @@protoc_insertion_point(class_scope:protobuf.options.AuthRule)
  })
_sym_db.RegisterMessage(AuthRule)


DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)
