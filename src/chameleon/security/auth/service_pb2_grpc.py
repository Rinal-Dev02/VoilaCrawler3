# encoding=utf8
# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

from chameleon.security.auth import service_message_pb2 as chameleon_dot_security_dot_auth_dot_service__message__pb2
from google.protobuf import empty_pb2 as google_dot_protobuf_dot_empty__pb2


class AuthorizerStub(object):
    """认证授权服务
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GetAppInfo = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/GetAppInfo',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoResponse.FromString,
                )
        self.GetUserInfo = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/GetUserInfo',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoResponse.FromString,
                )
        self.Register = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/Register',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterResponse.FromString,
                )
        self.Authorize = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/Authorize',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeResponse.FromString,
                )
        self.ValidateAccessToken = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/ValidateAccessToken',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenResposne.FromString,
                )
        self.AuthorizeAccess = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/AuthorizeAccess',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessRequest.SerializeToString,
                response_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessResponse.FromString,
                )
        self.PermitApplication = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/PermitApplication',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.PermitApplicationRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                )
        self.RevokeApplication = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/RevokeApplication',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RevokeApplicationRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                )
        self.DenyApplication = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/DenyApplication',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.DenyApplicationRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                )
        self.Verify = channel.unary_unary(
                '/chameleon.security.auth.Authorizer/Verify',
                request_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.VerifyRequest.SerializeToString,
                response_deserializer=google_dot_protobuf_dot_empty__pb2.Empty.FromString,
                )


class AuthorizerServicer(object):
    """认证授权服务
    """

    def GetAppInfo(self, request, context):
        """获得应用信息
        这个API与`Application Manager`中的`Get`API的区别是这个接口不需要认证且严格限制了返回的内容
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def GetUserInfo(self, request, context):
        """获得用户信息
        这个API与`Application Manager`中的`Get`API的区别是这个接口不需要认证且严格限制了返回的内容
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Register(self, request, context):
        """Register
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Authorize(self, request, context):
        """执行授权
        该操作将认证提交的数据并返回一个可以用于访问系统的Token数据
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def ValidateAccessToken(self, request, context):
        """验证AccessToken
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def AuthorizeAccess(self, request, context):
        """执行授权访问
        返回授权之后的Impersonation可以用于在整个调用链中传递调用方信息。
        __注意__： 该身份信息只能在系统内部使用，可以由`iam secret`进行校验（返回值封装于jwt格式中）
        返回的信息同时包括了主体（用户、应用等）的相关信息
        注意：该API只做与调用的API无关的授权工作（譬如某API需要调用方具有特殊权限，则不在这个API中检查），其保证的是authorization内容的合法性以及授权主体（用户、应用等）的合法性
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def PermitApplication(self, request, context):
        """批准应用
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def RevokeApplication(self, request, context):
        """撤销授权应用
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def DenyApplication(self, request, context):
        """拒绝授权应用
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')

    def Verify(self, request, context):
        """邮箱验证
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_AuthorizerServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GetAppInfo': grpc.unary_unary_rpc_method_handler(
                    servicer.GetAppInfo,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoResponse.SerializeToString,
            ),
            'GetUserInfo': grpc.unary_unary_rpc_method_handler(
                    servicer.GetUserInfo,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoResponse.SerializeToString,
            ),
            'Register': grpc.unary_unary_rpc_method_handler(
                    servicer.Register,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterResponse.SerializeToString,
            ),
            'Authorize': grpc.unary_unary_rpc_method_handler(
                    servicer.Authorize,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeResponse.SerializeToString,
            ),
            'ValidateAccessToken': grpc.unary_unary_rpc_method_handler(
                    servicer.ValidateAccessToken,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenResposne.SerializeToString,
            ),
            'AuthorizeAccess': grpc.unary_unary_rpc_method_handler(
                    servicer.AuthorizeAccess,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessRequest.FromString,
                    response_serializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessResponse.SerializeToString,
            ),
            'PermitApplication': grpc.unary_unary_rpc_method_handler(
                    servicer.PermitApplication,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.PermitApplicationRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'RevokeApplication': grpc.unary_unary_rpc_method_handler(
                    servicer.RevokeApplication,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.RevokeApplicationRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'DenyApplication': grpc.unary_unary_rpc_method_handler(
                    servicer.DenyApplication,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.DenyApplicationRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
            'Verify': grpc.unary_unary_rpc_method_handler(
                    servicer.Verify,
                    request_deserializer=chameleon_dot_security_dot_auth_dot_service__message__pb2.VerifyRequest.FromString,
                    response_serializer=google_dot_protobuf_dot_empty__pb2.Empty.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'chameleon.security.auth.Authorizer', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Authorizer(object):
    """认证授权服务
    """

    @staticmethod
    def GetAppInfo(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/GetAppInfo',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.GetApplicationInfoResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def GetUserInfo(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/GetUserInfo',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.GetUserInfoResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Register(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/Register',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.RegisterResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Authorize(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/Authorize',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def ValidateAccessToken(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/ValidateAccessToken',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.ValidateAccessTokenResposne.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def AuthorizeAccess(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/AuthorizeAccess',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessRequest.SerializeToString,
            chameleon_dot_security_dot_auth_dot_service__message__pb2.AuthorizeAccessResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def PermitApplication(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/PermitApplication',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.PermitApplicationRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def RevokeApplication(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/RevokeApplication',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.RevokeApplicationRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def DenyApplication(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/DenyApplication',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.DenyApplicationRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)

    @staticmethod
    def Verify(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/chameleon.security.auth.Authorizer/Verify',
            chameleon_dot_security_dot_auth_dot_service__message__pb2.VerifyRequest.SerializeToString,
            google_dot_protobuf_dot_empty__pb2.Empty.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
