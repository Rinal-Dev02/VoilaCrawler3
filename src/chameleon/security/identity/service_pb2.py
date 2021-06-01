# encoding=utf8
# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: chameleon/security/identity/service.proto
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2
from protobuf import annotations_pb2 as protobuf_dot_annotations__pb2
from protobuf.google.api import annotations_pb2 as protobuf_dot_google_dot_api_dot_annotations__pb2
from chameleon.security.identity import data_pb2 as chameleon_dot_security_dot_identity_dot_data__pb2
from chameleon.security.identity import service_message_pb2 as chameleon_dot_security_dot_identity_dot_service__message__pb2


DESCRIPTOR = _descriptor.FileDescriptor(
  name='chameleon/security/identity/service.proto',
  package='chameleon.security.identity',
  syntax='proto3',
  serialized_options=b'\n\037com.chameleon.security.identityB\014ServiceProtoP\001Z$chameleon/security/identity;identity\210\001\001',
  create_key=_descriptor._internal_create_key,
  serialized_pb=b'\n)chameleon/security/identity/service.proto\x12\x1b\x63hameleon.security.identity\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1aprotobuf/annotations.proto\x1a%protobuf/google/api/annotations.proto\x1a&chameleon/security/identity/data.proto\x1a\x31\x63hameleon/security/identity/service_message.proto2\x94\x11\n\x0bUserManager\x12\x9e\x01\n\x05\x43ount\x12-.chameleon.security.identity.CountUserRequest\x1a..chameleon.security.identity.CountUserResponse\"6\x82\xd3\xe4\x93\x02\"\"\x1d/security/identity/user/count:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x9a\x01\n\x04List\x12,.chameleon.security.identity.ListUserRequest\x1a-.chameleon.security.identity.ListUserResponse\"5\x82\xd3\xe4\x93\x02!\"\x1c/security/identity/user/list:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12l\n\x03Who\x12\x16.google.protobuf.Empty\x1a!.chameleon.security.identity.User\"*\x82\xd3\xe4\x93\x02\x1d\x12\x1b/security/identity/user/who\x82\x82\x87\x03\x02\x08\x02\x12\x8a\x01\n\x05\x45xist\x12/.chameleon.security.identity.IsUserExistRequest\x1a\x16.google.protobuf.Empty\"8\x82\xd3\xe4\x93\x02$\x12\"/security/identity/user/exist/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xa6\x01\n\x06\x45xists\x12\x30.chameleon.security.identity.IsUserExistsRequest\x1a\x31.chameleon.security.identity.IsUserExistsResponse\"7\x82\xd3\xe4\x93\x02#\"\x1e/security/identity/user/exists:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x89\x01\n\x03Get\x12+.chameleon.security.identity.GetUserRequest\x1a!.chameleon.security.identity.User\"2\x82\xd3\xe4\x93\x02\x1e\x12\x1c/security/identity/user/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x9a\x01\n\x04Gets\x12,.chameleon.security.identity.GetsUserRequest\x1a-.chameleon.security.identity.GetsUserResponse\"5\x82\xd3\xe4\x93\x02!\"\x1c/security/identity/user/gets:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x80\x01\n\x06\x43reate\x12!.chameleon.security.identity.User\x1a!.chameleon.security.identity.User\"0\x82\xd3\xe4\x93\x02\x1c\"\x17/security/identity/user:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x82\x01\n\x06Update\x12..chameleon.security.identity.UpdateUserRequest\x1a\x16.google.protobuf.Empty\"0\x82\xd3\xe4\x93\x02\x1c\x32\x17/security/identity/user:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xa1\x01\n\x08GetRoles\x12\x30.chameleon.security.identity.GetUserRolesRequest\x1a\x31.chameleon.security.identity.GetUserRolesResponse\"0\x82\xd3\xe4\x93\x02#\x12!/security/identity/user/{id}/role\x82\x82\x87\x03\x02\x08\x02\x12\x8b\x01\n\x07\x41\x64\x64Role\x12/.chameleon.security.identity.AddUserRoleRequest\x1a\x16.google.protobuf.Empty\"7\x82\xd3\xe4\x93\x02#\"!/security/identity/user/{id}/role\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x93\x01\n\x0bReplaceRole\x12\x33.chameleon.security.identity.ReplaceUserRoleRequest\x1a\x16.google.protobuf.Empty\"7\x82\xd3\xe4\x93\x02#\x1a!/security/identity/user/{id}/role\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x91\x01\n\nDeleteRole\x12\x32.chameleon.security.identity.DeleteUserRoleRequest\x1a\x16.google.protobuf.Empty\"7\x82\xd3\xe4\x93\x02#*!/security/identity/user/{id}/role\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x84\x01\n\x06\x44\x65lete\x12..chameleon.security.identity.DeleteUserRequest\x1a\x16.google.protobuf.Empty\"2\x82\xd3\xe4\x93\x02\x1e*\x1c/security/identity/user/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x8e\x01\n\x07Restore\x12/.chameleon.security.identity.RestoreUserRequest\x1a\x16.google.protobuf.Empty\":\x82\xd3\xe4\x93\x02&\"$/security/identity/user/{id}/restore\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min2\x96\'\n\x12\x41pplicationManager\x12\xb3\x01\n\x05\x43ount\x12\x34.chameleon.security.identity.CountApplicationRequest\x1a\x35.chameleon.security.identity.CountApplicationResponse\"=\x82\xd3\xe4\x93\x02)\"$/security/identity/application/count:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xaf\x01\n\x04List\x12\x33.chameleon.security.identity.ListApplicationRequest\x1a\x34.chameleon.security.identity.ListApplicationResponse\"<\x82\xd3\xe4\x93\x02(\"#/security/identity/application/list:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x98\x01\n\x05\x45xist\x12\x36.chameleon.security.identity.IsApplicationExistRequest\x1a\x16.google.protobuf.Empty\"?\x82\xd3\xe4\x93\x02+\x12)/security/identity/application/exist/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xbb\x01\n\x06\x45xists\x12\x37.chameleon.security.identity.IsApplicationExistsRequest\x1a\x38.chameleon.security.identity.IsApplicationExistsResponse\">\x82\xd3\xe4\x93\x02*\"%/security/identity/application/exists:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x9e\x01\n\x03Get\x12\x32.chameleon.security.identity.GetApplicationRequest\x1a(.chameleon.security.identity.Application\"9\x82\xd3\xe4\x93\x02%\x12#/security/identity/application/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xaf\x01\n\x04Gets\x12\x33.chameleon.security.identity.GetsApplicationRequest\x1a\x34.chameleon.security.identity.GetsApplicationResponse\"<\x82\xd3\xe4\x93\x02(\"#/security/identity/application/gets:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x95\x01\n\x06\x43reate\x12(.chameleon.security.identity.Application\x1a(.chameleon.security.identity.Application\"7\x82\xd3\xe4\x93\x02#\"\x1e/security/identity/application:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x90\x01\n\x06Update\x12\x35.chameleon.security.identity.UpdateApplicationRequest\x1a\x16.google.protobuf.Empty\"7\x82\xd3\xe4\x93\x02#2\x1e/security/identity/application:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x92\x01\n\x06\x44\x65lete\x12\x35.chameleon.security.identity.DeleteApplicationRequest\x1a\x16.google.protobuf.Empty\"9\x82\xd3\xe4\x93\x02%*#/security/identity/application/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x9c\x01\n\x07Restore\x12\x36.chameleon.security.identity.RestoreApplicationRequest\x1a\x16.google.protobuf.Empty\"A\x82\xd3\xe4\x93\x02-\x12+/security/identity/application/restore/{id}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xa5\x01\n\nSetOptions\x12\x39.chameleon.security.identity.SetApplicationOptionsRequest\x1a\x16.google.protobuf.Empty\"D\x82\xd3\xe4\x93\x02\x30\x1a+/security/identity/application/{id}/options:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\x9c\x01\n\x07SetTags\x12\x36.chameleon.security.identity.SetApplicationTagsRequest\x1a\x16.google.protobuf.Empty\"A\x82\xd3\xe4\x93\x02-\x1a(/security/identity/application/{id}/tags:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xce\x01\n\x0bResetSecret\x12:.chameleon.security.identity.ResetApplicationSecretRequest\x1a;.chameleon.security.identity.ResetApplicationSecretResponse\"F\x82\xd3\xe4\x93\x02\x32\x1a\x30/security/identity/application/{id}/secret/reset\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xc3\x01\n\x0cGetSecretKey\x12;.chameleon.security.identity.GetApplicationSecretKeyRequest\x1a\x31.chameleon.security.identity.ApplicationSecretKey\"C\x82\xd3\xe4\x93\x02/\x12-/security/identity/application/{id}/secretkey\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xc8\x01\n\x12GetSecretPublicKey\x12\x41.chameleon.security.identity.GetApplicationSecretPublicKeyRequest\x1a\x31.chameleon.security.identity.ApplicationSecretKey\"<\x82\xd3\xe4\x93\x02\x36\x12\x34/security/identity/application/{id}/secretkey/public\x12\xc3\x01\n\x0c\x41\x64\x64SecretKey\x12;.chameleon.security.identity.AddApplicationSecretKeyRequest\x1a\x31.chameleon.security.identity.ApplicationSecretKey\"C\x82\xd3\xe4\x93\x02/\"-/security/identity/application/{id}/secretkey\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xbe\x01\n\x13SetDefaultSecretKey\x12\x42.chameleon.security.identity.SetApplicationDefaultSecretKeyRequest\x1a\x16.google.protobuf.Empty\"K\x82\xd3\xe4\x93\x02\x37\"5/security/identity/application/{id}/secretkey/default\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xb6\x01\n\x0f\x44\x65leteSecretKey\x12>.chameleon.security.identity.DeleteApplicationSecretKeyRequest\x1a\x16.google.protobuf.Empty\"K\x82\xd3\xe4\x93\x02\x37*5/security/identity/application/{id}/secretkey/{keyID}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xe5\x01\n\x15GetDefaultRedirectURI\x12\x44.chameleon.security.identity.GetApplicationDefaultRedirectURIRequest\x1a\x45.chameleon.security.identity.GetApplicationDefaultRedirectURIResponse\"?\x82\xd3\xe4\x93\x02\x39\x12\x37/security/identity/application/{id}/redirecturi/default\x12\xc4\x01\n\x15SetDefaultRedirectURI\x12\x44.chameleon.security.identity.SetApplicationDefaultRedirectURIRequest\x1a\x16.google.protobuf.Empty\"M\x82\xd3\xe4\x93\x02\x39\"7/security/identity/application/{id}/redirecturi/default\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xe8\x01\n\x14GetWhiteRedirectURIs\x12\x43.chameleon.security.identity.GetApplicationWhiteRedirectURIsRequest\x1a\x44.chameleon.security.identity.GetApplicationWhiteRedirectURIsResponse\"E\x82\xd3\xe4\x93\x02\x31\x12//security/identity/application/{id}/redirecturi\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xdd\x01\n\x13\x41\x64\x64WhiteRedirectURI\x12\x42.chameleon.security.identity.AddApplicationWhiteRedirectURIRequest\x1a\x38.chameleon.security.identity.ApplicationWhiteRedirectURI\"H\x82\xd3\xe4\x93\x02\x34\"//security/identity/application/{id}/redirecturi:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xc6\x01\n\x16\x44\x65leteWhiteRedirectURI\x12\x45.chameleon.security.identity.DeleteApplicationWhiteRedirectURIRequest\x1a\x16.google.protobuf.Empty\"M\x82\xd3\xe4\x93\x02\x39*7/security/identity/application/{id}/redirecturi/{uriID}\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xbe\x01\n\x16\x43learWhiteRedirectURIs\x12\x45.chameleon.security.identity.ClearApplicationWhiteRedirectURIsRequest\x1a\x16.google.protobuf.Empty\"E\x82\xd3\xe4\x93\x02\x31*//security/identity/application/{id}/redirecturi\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xae\x01\n\x13\x41\x64\x64\x41pplicationScope\x12\x37.chameleon.security.identity.AddApplicationScopeRequest\x1a\x16.google.protobuf.Empty\"F\x82\xd3\xe4\x93\x02\x32\"-/security/identity/application/{id}/scope/add:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xb4\x01\n\x15ResetApplicationScope\x12\x39.chameleon.security.identity.ResetApplicationScopeRequest\x1a\x16.google.protobuf.Empty\"H\x82\xd3\xe4\x93\x02\x34\"//security/identity/application/{id}/scope/reset:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64min\x12\xb7\x01\n\x16RemoveApplicationScope\x12:.chameleon.security.identity.RemoveApplicationScopeRequest\x1a\x16.google.protobuf.Empty\"I\x82\xd3\xe4\x93\x02\x35\"0/security/identity/application/{id}/scope/remove:\x01*\x82\x82\x87\x03\t\x08\x02\x12\x05\x61\x64minBZ\n\x1f\x63om.chameleon.security.identityB\x0cServiceProtoP\x01Z$chameleon/security/identity;identity\x88\x01\x01\x62\x06proto3'
  ,
  dependencies=[google_dot_protobuf_dot_empty__pb2.DESCRIPTOR,protobuf_dot_annotations__pb2.DESCRIPTOR,protobuf_dot_google_dot_api_dot_annotations__pb2.DESCRIPTOR,chameleon_dot_security_dot_identity_dot_data__pb2.DESCRIPTOR,chameleon_dot_security_dot_identity_dot_service__message__pb2.DESCRIPTOR,])



_sym_db.RegisterFileDescriptor(DESCRIPTOR)


DESCRIPTOR._options = None

_USERMANAGER = _descriptor.ServiceDescriptor(
  name='UserManager',
  full_name='chameleon.security.identity.UserManager',
  file=DESCRIPTOR,
  index=0,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=262,
  serialized_end=2458,
  methods=[
  _descriptor.MethodDescriptor(
    name='Count',
    full_name='chameleon.security.identity.UserManager.Count',
    index=0,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._COUNTUSERREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._COUNTUSERRESPONSE,
    serialized_options=b'\202\323\344\223\002\"\"\035/security/identity/user/count:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='List',
    full_name='chameleon.security.identity.UserManager.List',
    index=1,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._LISTUSERREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._LISTUSERRESPONSE,
    serialized_options=b'\202\323\344\223\002!\"\034/security/identity/user/list:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Who',
    full_name='chameleon.security.identity.UserManager.Who',
    index=2,
    containing_service=None,
    input_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._USER,
    serialized_options=b'\202\323\344\223\002\035\022\033/security/identity/user/who\202\202\207\003\002\010\002',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Exist',
    full_name='chameleon.security.identity.UserManager.Exist',
    index=3,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISUSEREXISTREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002$\022\"/security/identity/user/exist/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Exists',
    full_name='chameleon.security.identity.UserManager.Exists',
    index=4,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISUSEREXISTSREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISUSEREXISTSRESPONSE,
    serialized_options=b'\202\323\344\223\002#\"\036/security/identity/user/exists:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Get',
    full_name='chameleon.security.identity.UserManager.Get',
    index=5,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETUSERREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._USER,
    serialized_options=b'\202\323\344\223\002\036\022\034/security/identity/user/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Gets',
    full_name='chameleon.security.identity.UserManager.Gets',
    index=6,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETSUSERREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETSUSERRESPONSE,
    serialized_options=b'\202\323\344\223\002!\"\034/security/identity/user/gets:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Create',
    full_name='chameleon.security.identity.UserManager.Create',
    index=7,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_data__pb2._USER,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._USER,
    serialized_options=b'\202\323\344\223\002\034\"\027/security/identity/user:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Update',
    full_name='chameleon.security.identity.UserManager.Update',
    index=8,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._UPDATEUSERREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002\0342\027/security/identity/user:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GetRoles',
    full_name='chameleon.security.identity.UserManager.GetRoles',
    index=9,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETUSERROLESREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETUSERROLESRESPONSE,
    serialized_options=b'\202\323\344\223\002#\022!/security/identity/user/{id}/role\202\202\207\003\002\010\002',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='AddRole',
    full_name='chameleon.security.identity.UserManager.AddRole',
    index=10,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ADDUSERROLEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002#\"!/security/identity/user/{id}/role\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ReplaceRole',
    full_name='chameleon.security.identity.UserManager.ReplaceRole',
    index=11,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._REPLACEUSERROLEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002#\032!/security/identity/user/{id}/role\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='DeleteRole',
    full_name='chameleon.security.identity.UserManager.DeleteRole',
    index=12,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._DELETEUSERROLEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002#*!/security/identity/user/{id}/role\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Delete',
    full_name='chameleon.security.identity.UserManager.Delete',
    index=13,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._DELETEUSERREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002\036*\034/security/identity/user/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Restore',
    full_name='chameleon.security.identity.UserManager.Restore',
    index=14,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._RESTOREUSERREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002&\"$/security/identity/user/{id}/restore\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_USERMANAGER)

DESCRIPTOR.services_by_name['UserManager'] = _USERMANAGER


_APPLICATIONMANAGER = _descriptor.ServiceDescriptor(
  name='ApplicationManager',
  full_name='chameleon.security.identity.ApplicationManager',
  file=DESCRIPTOR,
  index=1,
  serialized_options=None,
  create_key=_descriptor._internal_create_key,
  serialized_start=2461,
  serialized_end=7475,
  methods=[
  _descriptor.MethodDescriptor(
    name='Count',
    full_name='chameleon.security.identity.ApplicationManager.Count',
    index=0,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._COUNTAPPLICATIONREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._COUNTAPPLICATIONRESPONSE,
    serialized_options=b'\202\323\344\223\002)\"$/security/identity/application/count:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='List',
    full_name='chameleon.security.identity.ApplicationManager.List',
    index=1,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._LISTAPPLICATIONREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._LISTAPPLICATIONRESPONSE,
    serialized_options=b'\202\323\344\223\002(\"#/security/identity/application/list:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Exist',
    full_name='chameleon.security.identity.ApplicationManager.Exist',
    index=2,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISAPPLICATIONEXISTREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002+\022)/security/identity/application/exist/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Exists',
    full_name='chameleon.security.identity.ApplicationManager.Exists',
    index=3,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISAPPLICATIONEXISTSREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ISAPPLICATIONEXISTSRESPONSE,
    serialized_options=b'\202\323\344\223\002*\"%/security/identity/application/exists:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Get',
    full_name='chameleon.security.identity.ApplicationManager.Get',
    index=4,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATION,
    serialized_options=b'\202\323\344\223\002%\022#/security/identity/application/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Gets',
    full_name='chameleon.security.identity.ApplicationManager.Gets',
    index=5,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETSAPPLICATIONREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETSAPPLICATIONRESPONSE,
    serialized_options=b'\202\323\344\223\002(\"#/security/identity/application/gets:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Create',
    full_name='chameleon.security.identity.ApplicationManager.Create',
    index=6,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATION,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATION,
    serialized_options=b'\202\323\344\223\002#\"\036/security/identity/application:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Update',
    full_name='chameleon.security.identity.ApplicationManager.Update',
    index=7,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._UPDATEAPPLICATIONREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002#2\036/security/identity/application:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Delete',
    full_name='chameleon.security.identity.ApplicationManager.Delete',
    index=8,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._DELETEAPPLICATIONREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002%*#/security/identity/application/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='Restore',
    full_name='chameleon.security.identity.ApplicationManager.Restore',
    index=9,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._RESTOREAPPLICATIONREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002-\022+/security/identity/application/restore/{id}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='SetOptions',
    full_name='chameleon.security.identity.ApplicationManager.SetOptions',
    index=10,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._SETAPPLICATIONOPTIONSREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0020\032+/security/identity/application/{id}/options:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='SetTags',
    full_name='chameleon.security.identity.ApplicationManager.SetTags',
    index=11,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._SETAPPLICATIONTAGSREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\002-\032(/security/identity/application/{id}/tags:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ResetSecret',
    full_name='chameleon.security.identity.ApplicationManager.ResetSecret',
    index=12,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._RESETAPPLICATIONSECRETREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._RESETAPPLICATIONSECRETRESPONSE,
    serialized_options=b'\202\323\344\223\0022\0320/security/identity/application/{id}/secret/reset\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GetSecretKey',
    full_name='chameleon.security.identity.ApplicationManager.GetSecretKey',
    index=13,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONSECRETKEYREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATIONSECRETKEY,
    serialized_options=b'\202\323\344\223\002/\022-/security/identity/application/{id}/secretkey\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GetSecretPublicKey',
    full_name='chameleon.security.identity.ApplicationManager.GetSecretPublicKey',
    index=14,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONSECRETPUBLICKEYREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATIONSECRETKEY,
    serialized_options=b'\202\323\344\223\0026\0224/security/identity/application/{id}/secretkey/public',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='AddSecretKey',
    full_name='chameleon.security.identity.ApplicationManager.AddSecretKey',
    index=15,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ADDAPPLICATIONSECRETKEYREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATIONSECRETKEY,
    serialized_options=b'\202\323\344\223\002/\"-/security/identity/application/{id}/secretkey\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='SetDefaultSecretKey',
    full_name='chameleon.security.identity.ApplicationManager.SetDefaultSecretKey',
    index=16,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._SETAPPLICATIONDEFAULTSECRETKEYREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0027\"5/security/identity/application/{id}/secretkey/default\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='DeleteSecretKey',
    full_name='chameleon.security.identity.ApplicationManager.DeleteSecretKey',
    index=17,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._DELETEAPPLICATIONSECRETKEYREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0027*5/security/identity/application/{id}/secretkey/{keyID}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GetDefaultRedirectURI',
    full_name='chameleon.security.identity.ApplicationManager.GetDefaultRedirectURI',
    index=18,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONDEFAULTREDIRECTURIREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONDEFAULTREDIRECTURIRESPONSE,
    serialized_options=b'\202\323\344\223\0029\0227/security/identity/application/{id}/redirecturi/default',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='SetDefaultRedirectURI',
    full_name='chameleon.security.identity.ApplicationManager.SetDefaultRedirectURI',
    index=19,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._SETAPPLICATIONDEFAULTREDIRECTURIREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0029\"7/security/identity/application/{id}/redirecturi/default\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='GetWhiteRedirectURIs',
    full_name='chameleon.security.identity.ApplicationManager.GetWhiteRedirectURIs',
    index=20,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONWHITEREDIRECTURISREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._GETAPPLICATIONWHITEREDIRECTURISRESPONSE,
    serialized_options=b'\202\323\344\223\0021\022//security/identity/application/{id}/redirecturi\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='AddWhiteRedirectURI',
    full_name='chameleon.security.identity.ApplicationManager.AddWhiteRedirectURI',
    index=21,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ADDAPPLICATIONWHITEREDIRECTURIREQUEST,
    output_type=chameleon_dot_security_dot_identity_dot_data__pb2._APPLICATIONWHITEREDIRECTURI,
    serialized_options=b'\202\323\344\223\0024\"//security/identity/application/{id}/redirecturi:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='DeleteWhiteRedirectURI',
    full_name='chameleon.security.identity.ApplicationManager.DeleteWhiteRedirectURI',
    index=22,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._DELETEAPPLICATIONWHITEREDIRECTURIREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0029*7/security/identity/application/{id}/redirecturi/{uriID}\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ClearWhiteRedirectURIs',
    full_name='chameleon.security.identity.ApplicationManager.ClearWhiteRedirectURIs',
    index=23,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._CLEARAPPLICATIONWHITEREDIRECTURISREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0021*//security/identity/application/{id}/redirecturi\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='AddApplicationScope',
    full_name='chameleon.security.identity.ApplicationManager.AddApplicationScope',
    index=24,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._ADDAPPLICATIONSCOPEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0022\"-/security/identity/application/{id}/scope/add:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='ResetApplicationScope',
    full_name='chameleon.security.identity.ApplicationManager.ResetApplicationScope',
    index=25,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._RESETAPPLICATIONSCOPEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0024\"//security/identity/application/{id}/scope/reset:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
  _descriptor.MethodDescriptor(
    name='RemoveApplicationScope',
    full_name='chameleon.security.identity.ApplicationManager.RemoveApplicationScope',
    index=26,
    containing_service=None,
    input_type=chameleon_dot_security_dot_identity_dot_service__message__pb2._REMOVEAPPLICATIONSCOPEREQUEST,
    output_type=google_dot_protobuf_dot_empty__pb2._EMPTY,
    serialized_options=b'\202\323\344\223\0025\"0/security/identity/application/{id}/scope/remove:\001*\202\202\207\003\t\010\002\022\005admin',
    create_key=_descriptor._internal_create_key,
  ),
])
_sym_db.RegisterServiceDescriptor(_APPLICATIONMANAGER)

DESCRIPTOR.services_by_name['ApplicationManager'] = _APPLICATIONMANAGER

# @@protoc_insertion_point(module_scope)
