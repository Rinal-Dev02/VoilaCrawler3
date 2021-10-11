# encoding=utf8
# extension plugin generated codes. DO NOT EDIT!!!

from google.protobuf.descriptor_pb2 import FileDescriptorProto

from chameleon.api.invocation import ServiceDesc

from .service_pb2 import DESCRIPTOR

FileDescriptor = FileDescriptorProto()
DESCRIPTOR.CopyToProto(FileDescriptor)



from .service_pb2_grpc import SessionManagerServicer as _SessionManagerServicer, SessionManagerStub as _SessionManagerStub, add_SessionManagerServicer_to_server

class SessionManagerServicer(_SessionManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "SessionManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def GetCookies(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCookies")
                if not methodDesc:
                    raise RuntimeError("Method [GetCookies] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCookies)
            return self.servicer.GetCookies(request, context)
	
        def SetCookies(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("SetCookies")
                if not methodDesc:
                    raise RuntimeError("Method [SetCookies] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.SetCookies)
            return self.servicer.SetCookies(request, context)
	
        def ClearCookies(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("ClearCookies")
                if not methodDesc:
                    raise RuntimeError("Method [ClearCookies] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.ClearCookies)
            return self.servicer.ClearCookies(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_SessionManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class SessionManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new SessionManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "SessionManager" ][0], FileDescriptor.package)
        self.____stub = _SessionManagerStub(channel)
        self.____interceptor = interceptor

	
    def GetCookies(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCookies")
            if not methodDesc:
                raise RuntimeError("Method [GetCookies] description not found")
            return self.____interceptor(self.____stub.GetCookies, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCookies(request, timeout, metadata, credentials)
	
    def SetCookies(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("SetCookies")
            if not methodDesc:
                raise RuntimeError("Method [SetCookies] description not found")
            return self.____interceptor(self.____stub.SetCookies, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.SetCookies(request, timeout, metadata, credentials)
	
    def ClearCookies(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("ClearCookies")
            if not methodDesc:
                raise RuntimeError("Method [ClearCookies] description not found")
            return self.____interceptor(self.____stub.ClearCookies, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.ClearCookies(request, timeout, metadata, credentials)
	


