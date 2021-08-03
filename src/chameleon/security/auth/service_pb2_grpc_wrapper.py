# encoding=utf8
# extension plugin generated codes. DO NOT EDIT!!!

from google.protobuf.descriptor_pb2 import FileDescriptorProto

from chameleon.api.invocation import ServiceDesc

from .service_pb2 import DESCRIPTOR

FileDescriptor = FileDescriptorProto()
DESCRIPTOR.CopyToProto(FileDescriptor)



from .service_pb2_grpc import AuthorizerServicer as _AuthorizerServicer, AuthorizerStub as _AuthorizerStub, add_AuthorizerServicer_to_server

class AuthorizerServicer(_AuthorizerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "Authorizer" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def GetAppInfo(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetAppInfo")
                if not methodDesc:
                    raise RuntimeError("Method [GetAppInfo] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetAppInfo)
            return self.servicer.GetAppInfo(request, context)
	
        def GetUserInfo(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetUserInfo")
                if not methodDesc:
                    raise RuntimeError("Method [GetUserInfo] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetUserInfo)
            return self.servicer.GetUserInfo(request, context)
	
        def Register(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Register")
                if not methodDesc:
                    raise RuntimeError("Method [Register] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Register)
            return self.servicer.Register(request, context)
	
        def Authorize(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Authorize")
                if not methodDesc:
                    raise RuntimeError("Method [Authorize] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Authorize)
            return self.servicer.Authorize(request, context)
	
        def ValidateAccessToken(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("ValidateAccessToken")
                if not methodDesc:
                    raise RuntimeError("Method [ValidateAccessToken] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.ValidateAccessToken)
            return self.servicer.ValidateAccessToken(request, context)
	
        def AuthorizeAccess(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AuthorizeAccess")
                if not methodDesc:
                    raise RuntimeError("Method [AuthorizeAccess] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AuthorizeAccess)
            return self.servicer.AuthorizeAccess(request, context)
	
        def PermitApplication(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("PermitApplication")
                if not methodDesc:
                    raise RuntimeError("Method [PermitApplication] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.PermitApplication)
            return self.servicer.PermitApplication(request, context)
	
        def RevokeApplication(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("RevokeApplication")
                if not methodDesc:
                    raise RuntimeError("Method [RevokeApplication] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.RevokeApplication)
            return self.servicer.RevokeApplication(request, context)
	
        def DenyApplication(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DenyApplication")
                if not methodDesc:
                    raise RuntimeError("Method [DenyApplication] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DenyApplication)
            return self.servicer.DenyApplication(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_AuthorizerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class AuthorizerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new AuthorizerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "Authorizer" ][0], FileDescriptor.package)
        self.____stub = _AuthorizerStub(channel)
        self.____interceptor = interceptor

	
    def GetAppInfo(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetAppInfo")
            if not methodDesc:
                raise RuntimeError("Method [GetAppInfo] description not found")
            return self.____interceptor(self.____stub.GetAppInfo, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetAppInfo(request, timeout, metadata, credentials)
	
    def GetUserInfo(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetUserInfo")
            if not methodDesc:
                raise RuntimeError("Method [GetUserInfo] description not found")
            return self.____interceptor(self.____stub.GetUserInfo, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetUserInfo(request, timeout, metadata, credentials)
	
    def Register(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Register")
            if not methodDesc:
                raise RuntimeError("Method [Register] description not found")
            return self.____interceptor(self.____stub.Register, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Register(request, timeout, metadata, credentials)
	
    def Authorize(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Authorize")
            if not methodDesc:
                raise RuntimeError("Method [Authorize] description not found")
            return self.____interceptor(self.____stub.Authorize, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Authorize(request, timeout, metadata, credentials)
	
    def ValidateAccessToken(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("ValidateAccessToken")
            if not methodDesc:
                raise RuntimeError("Method [ValidateAccessToken] description not found")
            return self.____interceptor(self.____stub.ValidateAccessToken, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.ValidateAccessToken(request, timeout, metadata, credentials)
	
    def AuthorizeAccess(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AuthorizeAccess")
            if not methodDesc:
                raise RuntimeError("Method [AuthorizeAccess] description not found")
            return self.____interceptor(self.____stub.AuthorizeAccess, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AuthorizeAccess(request, timeout, metadata, credentials)
	
    def PermitApplication(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("PermitApplication")
            if not methodDesc:
                raise RuntimeError("Method [PermitApplication] description not found")
            return self.____interceptor(self.____stub.PermitApplication, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.PermitApplication(request, timeout, metadata, credentials)
	
    def RevokeApplication(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("RevokeApplication")
            if not methodDesc:
                raise RuntimeError("Method [RevokeApplication] description not found")
            return self.____interceptor(self.____stub.RevokeApplication, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.RevokeApplication(request, timeout, metadata, credentials)
	
    def DenyApplication(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DenyApplication")
            if not methodDesc:
                raise RuntimeError("Method [DenyApplication] description not found")
            return self.____interceptor(self.____stub.DenyApplication, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DenyApplication(request, timeout, metadata, credentials)
	


