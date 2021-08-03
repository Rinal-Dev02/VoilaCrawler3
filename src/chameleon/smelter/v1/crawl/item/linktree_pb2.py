# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chameleon/smelter/v1/crawl/item/linktree.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from chameleon.api.media import data_pb2 as chameleon_dot_api_dot_media_dot_data__pb2
from chameleon.api.http import data_pb2 as chameleon_dot_api_dot_http_dot_data__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='chameleon/smelter/v1/crawl/item/linktree.proto',
  package='chameleon.smelter.v1.crawl.item',
  syntax='proto3',
  serialized_options=b'Z$chameleon/smelter/v1/crawl/item;item',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n.chameleon/smelter/v1/crawl/item/linktree.proto\x12\x1f\x63hameleon.smelter.v1.crawl.item\x1a\x1e\x63hameleon/api/media/data.proto\x1a\x1d\x63hameleon/api/http/data.proto\"\xab\x02\n\x08Linktree\x1a\x9e\x02\n\x04Item\x12G\n\x07profile\x18\x05 \x01(\x0b\x32\x36.chameleon.smelter.v1.crawl.item.Linktree.Item.Profile\x12\x42\n\x05links\x18\x06 \x03(\x0b\x32\x33.chameleon.smelter.v1.crawl.item.Linktree.Item.Link\x1a<\n\x07Profile\x12\x0c\n\x04name\x18\x02 \x01(\t\x12\x0e\n\x06\x61vatar\x18\x03 \x01(\t\x12\x13\n\x0blinktreeUrl\x18\x0b \x01(\t\x1aK\n\x04Link\x12\n\n\x02id\x18\x01 \x01(\t\x12\r\n\x05title\x18\x02 \x01(\t\x12\x0b\n\x03url\x18\x03 \x01(\t\x12\x0c\n\x04icon\x18\x06 \x01(\t\x12\r\n\x05style\x18\x07 \x01(\tB&Z$chameleon/smelter/v1/crawl/item;itemb\x06proto3'
  ,
  dependencies=[chameleon_dot_api_dot_media_dot_data__pb2.DESCRIPTOR,chameleon_dot_api_dot_http_dot_data__pb2.DESCRIPTOR,])




_LINKTREE_ITEM_PROFILE = _descriptor.Descriptor(
  name='Profile',
  full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Profile',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='name', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Profile.name', index=0,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='avatar', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Profile.avatar', index=1,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='linktreeUrl', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Profile.linktreeUrl', index=2,
      number=11, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
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
  serialized_start=309,
  serialized_end=369,
)

_LINKTREE_ITEM_LINK = _descriptor.Descriptor(
  name='Link',
  full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link.id', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='title', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link.title', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='url', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link.url', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='icon', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link.icon', index=3,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='style', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.Link.style', index=4,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
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
  serialized_start=371,
  serialized_end=446,
)

_LINKTREE_ITEM = _descriptor.Descriptor(
  name='Item',
  full_name='chameleon.smelter.v1.crawl.item.Linktree.Item',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='profile', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.profile', index=0,
      number=5, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='links', full_name='chameleon.smelter.v1.crawl.item.Linktree.Item.links', index=1,
      number=6, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_LINKTREE_ITEM_PROFILE, _LINKTREE_ITEM_LINK, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=160,
  serialized_end=446,
)

_LINKTREE = _descriptor.Descriptor(
  name='Linktree',
  full_name='chameleon.smelter.v1.crawl.item.Linktree',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[_LINKTREE_ITEM, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=147,
  serialized_end=446,
)

_LINKTREE_ITEM_PROFILE.containing_type = _LINKTREE_ITEM
_LINKTREE_ITEM_LINK.containing_type = _LINKTREE_ITEM
_LINKTREE_ITEM.fields_by_name['profile'].message_type = _LINKTREE_ITEM_PROFILE
_LINKTREE_ITEM.fields_by_name['links'].message_type = _LINKTREE_ITEM_LINK
_LINKTREE_ITEM.containing_type = _LINKTREE
DESCRIPTOR.message_types_by_name['Linktree'] = _LINKTREE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Linktree = _reflection.GeneratedProtocolMessageType('Linktree', (_message.Message,), {

  'Item' : _reflection.GeneratedProtocolMessageType('Item', (_message.Message,), {

    'Profile' : _reflection.GeneratedProtocolMessageType('Profile', (_message.Message,), {
      'DESCRIPTOR' : _LINKTREE_ITEM_PROFILE,
      '__module__' : 'chameleon.smelter.v1.crawl.item.linktree_pb2'
      # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Linktree.Item.Profile)
      })
    ,

    'Link' : _reflection.GeneratedProtocolMessageType('Link', (_message.Message,), {
      'DESCRIPTOR' : _LINKTREE_ITEM_LINK,
      '__module__' : 'chameleon.smelter.v1.crawl.item.linktree_pb2'
      # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Linktree.Item.Link)
      })
    ,
    'DESCRIPTOR' : _LINKTREE_ITEM,
    '__module__' : 'chameleon.smelter.v1.crawl.item.linktree_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Linktree.Item)
    })
  ,
  'DESCRIPTOR' : _LINKTREE,
  '__module__' : 'chameleon.smelter.v1.crawl.item.linktree_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Linktree)
  })
_sym_db.RegisterMessage(Linktree)
_sym_db.RegisterMessage(Linktree.Item)
_sym_db.RegisterMessage(Linktree.Item.Profile)
_sym_db.RegisterMessage(Linktree.Item.Link)


DESCRIPTOR._options = None
# @@protoc_insertion_point(module_scope)
