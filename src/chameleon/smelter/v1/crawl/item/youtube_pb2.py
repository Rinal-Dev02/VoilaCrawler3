# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chameleon/smelter/v1/crawl/item/youtube.proto
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
  name='chameleon/smelter/v1/crawl/item/youtube.proto',
  package='chameleon.smelter.v1.crawl.item',
  syntax='proto3',
  serialized_options=b'Z$chameleon/smelter/v1/crawl/item;item',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n-chameleon/smelter/v1/crawl/item/youtube.proto\x12\x1f\x63hameleon.smelter.v1.crawl.item\x1a\x1e\x63hameleon/api/media/data.proto\x1a\x1d\x63hameleon/api/http/data.proto\"\xbe\x07\n\x07Youtube\x1aM\n\x06Source\x12\n\n\x02id\x18\x01 \x01(\t\x12\x10\n\x08\x63rawlUrl\x18\x02 \x01(\t\x12\x11\n\tsourceUrl\x18\x05 \x01(\t\x12\x12\n\npublishUtc\x18\x0f \x01(\x03\x1a\xbd\x02\n\x07\x43hannel\x12\n\n\x02id\x18\x01 \x01(\t\x12\x14\n\x0c\x63\x61nonicalUrl\x18\x02 \x01(\t\x12\x10\n\x08username\x18\x03 \x01(\t\x12\r\n\x05title\x18\x04 \x01(\t\x12\x0e\n\x06\x61vatar\x18\x05 \x01(\t\x12\x13\n\x0b\x64\x65scription\x18\x06 \x01(\t\x12\x0f\n\x07\x63ountry\x18\x07 \x01(\t\x12\x45\n\x05stats\x18\x0b \x01(\x0b\x32\x36.chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats\x12\x14\n\x0cpublishedUtc\x18\x0f \x01(\x03\x1a\\\n\x05Stats\x12\x16\n\x0esubscribeCount\x18\x01 \x01(\x05\x12\x12\n\nvideoCount\x18\x02 \x01(\x05\x12\x11\n\tviewCount\x18\x03 \x01(\x05\x12\x14\n\x0c\x63ommentCount\x18\x05 \x01(\x05\x1a\\\n\x05Stats\x12\x16\n\x0esubscribeCount\x18\x01 \x01(\x05\x12\x12\n\nvideoCount\x18\x02 \x01(\x05\x12\x11\n\tviewCount\x18\x03 \x01(\x05\x12\x14\n\x0c\x63ommentCount\x18\x05 \x01(\x05\x1a\xc5\x03\n\x05Video\x12?\n\x06source\x18\x02 \x01(\x0b\x32/.chameleon.smelter.v1.crawl.item.Youtube.Source\x12\r\n\x05title\x18\x04 \x01(\t\x12\x13\n\x0b\x64\x65scription\x18\x05 \x01(\t\x12\x41\n\x07\x63hannel\x18\x06 \x01(\x0b\x32\x30.chameleon.smelter.v1.crawl.item.Youtube.Channel\x12/\n\x05video\x18\x0b \x01(\x0b\x32 .chameleon.api.media.Media.Video\x12L\n\x07headers\x18\x0f \x03(\x0b\x32;.chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry\x12=\n\x05stats\x18\x15 \x01(\x0b\x32..chameleon.smelter.v1.crawl.item.Youtube.Stats\x12\x12\n\ncrawledUtc\x18\x1f \x01(\x03\x12\x12\n\nexpiresUtc\x18  \x01(\x03\x1a.\n\x0cHeadersEntry\x12\x0b\n\x03key\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t:\x02\x38\x01\x42&Z$chameleon/smelter/v1/crawl/item;itemb\x06proto3'
  ,
  dependencies=[chameleon_dot_api_dot_media_dot_data__pb2.DESCRIPTOR,chameleon_dot_api_dot_http_dot_data__pb2.DESCRIPTOR,])




_YOUTUBE_SOURCE = _descriptor.Descriptor(
  name='Source',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Source',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='chameleon.smelter.v1.crawl.item.Youtube.Source.id', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='crawlUrl', full_name='chameleon.smelter.v1.crawl.item.Youtube.Source.crawlUrl', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='sourceUrl', full_name='chameleon.smelter.v1.crawl.item.Youtube.Source.sourceUrl', index=2,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='publishUtc', full_name='chameleon.smelter.v1.crawl.item.Youtube.Source.publishUtc', index=3,
      number=15, type=3, cpp_type=2, label=1,
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
  serialized_start=157,
  serialized_end=234,
)

_YOUTUBE_CHANNEL_STATS = _descriptor.Descriptor(
  name='Stats',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='subscribeCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats.subscribeCount', index=0,
      number=1, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='videoCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats.videoCount', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='viewCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats.viewCount', index=2,
      number=3, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='commentCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats.commentCount', index=3,
      number=5, type=5, cpp_type=1, label=1,
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
  serialized_start=462,
  serialized_end=554,
)

_YOUTUBE_CHANNEL = _descriptor.Descriptor(
  name='Channel',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='id', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.id', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='canonicalUrl', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.canonicalUrl', index=1,
      number=2, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='username', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.username', index=2,
      number=3, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='title', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.title', index=3,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='avatar', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.avatar', index=4,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='description', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.description', index=5,
      number=6, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='country', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.country', index=6,
      number=7, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='stats', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.stats', index=7,
      number=11, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='publishedUtc', full_name='chameleon.smelter.v1.crawl.item.Youtube.Channel.publishedUtc', index=8,
      number=15, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_YOUTUBE_CHANNEL_STATS, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=237,
  serialized_end=554,
)

_YOUTUBE_STATS = _descriptor.Descriptor(
  name='Stats',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Stats',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='subscribeCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Stats.subscribeCount', index=0,
      number=1, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='videoCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Stats.videoCount', index=1,
      number=2, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='viewCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Stats.viewCount', index=2,
      number=3, type=5, cpp_type=1, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='commentCount', full_name='chameleon.smelter.v1.crawl.item.Youtube.Stats.commentCount', index=3,
      number=5, type=5, cpp_type=1, label=1,
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
  serialized_start=462,
  serialized_end=554,
)

_YOUTUBE_VIDEO_HEADERSENTRY = _descriptor.Descriptor(
  name='HeadersEntry',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='key', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry.key', index=0,
      number=1, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='value', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry.value', index=1,
      number=2, type=9, cpp_type=9, label=1,
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
  serialized_options=b'8\001',
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=1058,
  serialized_end=1104,
)

_YOUTUBE_VIDEO = _descriptor.Descriptor(
  name='Video',
  full_name='chameleon.smelter.v1.crawl.item.Youtube.Video',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
    _descriptor.FieldDescriptor(
      name='source', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.source', index=0,
      number=2, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='title', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.title', index=1,
      number=4, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='description', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.description', index=2,
      number=5, type=9, cpp_type=9, label=1,
      has_default_value=False, default_value=b"".decode('utf-8'),
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='channel', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.channel', index=3,
      number=6, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='video', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.video', index=4,
      number=11, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='headers', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.headers', index=5,
      number=15, type=11, cpp_type=10, label=3,
      has_default_value=False, default_value=[],
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='stats', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.stats', index=6,
      number=21, type=11, cpp_type=10, label=1,
      has_default_value=False, default_value=None,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='crawledUtc', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.crawledUtc', index=7,
      number=31, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
    _descriptor.FieldDescriptor(
      name='expiresUtc', full_name='chameleon.smelter.v1.crawl.item.Youtube.Video.expiresUtc', index=8,
      number=32, type=3, cpp_type=2, label=1,
      has_default_value=False, default_value=0,
      message_type=None, enum_type=None, containing_type=None,
      is_extension=False, extension_scope=None,
      serialized_options=None, file=DESCRIPTOR,  create_key=_descriptor._internal_create_key),
  ],
  extensions=[
  ],
  nested_types=[_YOUTUBE_VIDEO_HEADERSENTRY, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=651,
  serialized_end=1104,
)

_YOUTUBE = _descriptor.Descriptor(
  name='Youtube',
  full_name='chameleon.smelter.v1.crawl.item.Youtube',
  filename=None,
  file=DESCRIPTOR,
  containing_type=None,
  create_key=_descriptor._internal_create_key,
  fields=[
  ],
  extensions=[
  ],
  nested_types=[_YOUTUBE_SOURCE, _YOUTUBE_CHANNEL, _YOUTUBE_STATS, _YOUTUBE_VIDEO, ],
  enum_types=[
  ],
  serialized_options=None,
  is_extendable=False,
  syntax='proto3',
  extension_ranges=[],
  oneofs=[
  ],
  serialized_start=146,
  serialized_end=1104,
)

_YOUTUBE_SOURCE.containing_type = _YOUTUBE
_YOUTUBE_CHANNEL_STATS.containing_type = _YOUTUBE_CHANNEL
_YOUTUBE_CHANNEL.fields_by_name['stats'].message_type = _YOUTUBE_CHANNEL_STATS
_YOUTUBE_CHANNEL.containing_type = _YOUTUBE
_YOUTUBE_STATS.containing_type = _YOUTUBE
_YOUTUBE_VIDEO_HEADERSENTRY.containing_type = _YOUTUBE_VIDEO
_YOUTUBE_VIDEO.fields_by_name['source'].message_type = _YOUTUBE_SOURCE
_YOUTUBE_VIDEO.fields_by_name['channel'].message_type = _YOUTUBE_CHANNEL
_YOUTUBE_VIDEO.fields_by_name['video'].message_type = chameleon_dot_api_dot_media_dot_data__pb2._MEDIA_VIDEO
_YOUTUBE_VIDEO.fields_by_name['headers'].message_type = _YOUTUBE_VIDEO_HEADERSENTRY
_YOUTUBE_VIDEO.fields_by_name['stats'].message_type = _YOUTUBE_STATS
_YOUTUBE_VIDEO.containing_type = _YOUTUBE
DESCRIPTOR.message_types_by_name['Youtube'] = _YOUTUBE
_sym_db.RegisterFileDescriptor(DESCRIPTOR)

Youtube = _reflection.GeneratedProtocolMessageType('Youtube', (_message.Message,), {

  'Source' : _reflection.GeneratedProtocolMessageType('Source', (_message.Message,), {
    'DESCRIPTOR' : _YOUTUBE_SOURCE,
    '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Source)
    })
  ,

  'Channel' : _reflection.GeneratedProtocolMessageType('Channel', (_message.Message,), {

    'Stats' : _reflection.GeneratedProtocolMessageType('Stats', (_message.Message,), {
      'DESCRIPTOR' : _YOUTUBE_CHANNEL_STATS,
      '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
      # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Channel.Stats)
      })
    ,
    'DESCRIPTOR' : _YOUTUBE_CHANNEL,
    '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Channel)
    })
  ,

  'Stats' : _reflection.GeneratedProtocolMessageType('Stats', (_message.Message,), {
    'DESCRIPTOR' : _YOUTUBE_STATS,
    '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Stats)
    })
  ,

  'Video' : _reflection.GeneratedProtocolMessageType('Video', (_message.Message,), {

    'HeadersEntry' : _reflection.GeneratedProtocolMessageType('HeadersEntry', (_message.Message,), {
      'DESCRIPTOR' : _YOUTUBE_VIDEO_HEADERSENTRY,
      '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
      # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Video.HeadersEntry)
      })
    ,
    'DESCRIPTOR' : _YOUTUBE_VIDEO,
    '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
    # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube.Video)
    })
  ,
  'DESCRIPTOR' : _YOUTUBE,
  '__module__' : 'chameleon.smelter.v1.crawl.item.youtube_pb2'
  # @@protoc_insertion_point(class_scope:chameleon.smelter.v1.crawl.item.Youtube)
  })
_sym_db.RegisterMessage(Youtube)
_sym_db.RegisterMessage(Youtube.Source)
_sym_db.RegisterMessage(Youtube.Channel)
_sym_db.RegisterMessage(Youtube.Channel.Stats)
_sym_db.RegisterMessage(Youtube.Stats)
_sym_db.RegisterMessage(Youtube.Video)
_sym_db.RegisterMessage(Youtube.Video.HeadersEntry)


DESCRIPTOR._options = None
_YOUTUBE_VIDEO_HEADERSENTRY._options = None
# @@protoc_insertion_point(module_scope)
