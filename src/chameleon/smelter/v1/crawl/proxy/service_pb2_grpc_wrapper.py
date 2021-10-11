# encoding=utf8
# extension plugin generated codes. DO NOT EDIT!!!

from google.protobuf.descriptor_pb2 import FileDescriptorProto

from chameleon.api.invocation import ServiceDesc

from .service_pb2 import DESCRIPTOR

FileDescriptor = FileDescriptorProto()
DESCRIPTOR.CopyToProto(FileDescriptor)



from .service_pb2_grpc import ProxyManagerServicer as _ProxyManagerServicer, ProxyManagerStub as _ProxyManagerStub, add_ProxyManagerServicer_to_server

class ProxyManagerServicer(_ProxyManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "ProxyManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def DoRequest(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DoRequest")
                if not methodDesc:
                    raise RuntimeError("Method [DoRequest] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DoRequest)
            return self.servicer.DoRequest(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_ProxyManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class ProxyManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new ProxyManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "ProxyManager" ][0], FileDescriptor.package)
        self.____stub = _ProxyManagerStub(channel)
        self.____interceptor = interceptor

	
    def DoRequest(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DoRequest")
            if not methodDesc:
                raise RuntimeError("Method [DoRequest] description not found")
            return self.____interceptor(self.____stub.DoRequest, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DoRequest(request, timeout, metadata, credentials)
	


