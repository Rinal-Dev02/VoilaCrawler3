# encoding=utf8
# extension plugin generated codes. DO NOT EDIT!!!

from google.protobuf.descriptor_pb2 import FileDescriptorProto

from chameleon.api.invocation import ServiceDesc

from .service_pb2 import DESCRIPTOR

FileDescriptor = FileDescriptorProto()
DESCRIPTOR.CopyToProto(FileDescriptor)



from .service_pb2_grpc import CrawlerNodeServicer as _CrawlerNodeServicer, CrawlerNodeStub as _CrawlerNodeStub, add_CrawlerNodeServicer_to_server

class CrawlerNodeServicer(_CrawlerNodeServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerNode" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def Version(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Version")
                if not methodDesc:
                    raise RuntimeError("Method [Version] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Version)
            return self.servicer.Version(request, context)
	
        def CrawlerOptions(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("CrawlerOptions")
                if not methodDesc:
                    raise RuntimeError("Method [CrawlerOptions] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.CrawlerOptions)
            return self.servicer.CrawlerOptions(request, context)
	
        def AllowedDomains(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AllowedDomains")
                if not methodDesc:
                    raise RuntimeError("Method [AllowedDomains] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AllowedDomains)
            return self.servicer.AllowedDomains(request, context)
	
        def CanonicalUrl(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("CanonicalUrl")
                if not methodDesc:
                    raise RuntimeError("Method [CanonicalUrl] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.CanonicalUrl)
            return self.servicer.CanonicalUrl(request, context)
	
        def Parse(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Parse")
                if not methodDesc:
                    raise RuntimeError("Method [Parse] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Parse)
            return self.servicer.Parse(request, context)
	
        def Call(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Call")
                if not methodDesc:
                    raise RuntimeError("Method [Call] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Call)
            return self.servicer.Call(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_CrawlerNodeServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class CrawlerNodeStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new CrawlerNodeStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerNode" ][0], FileDescriptor.package)
        self.____stub = _CrawlerNodeStub(channel)
        self.____interceptor = interceptor

	
    def Version(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Version")
            if not methodDesc:
                raise RuntimeError("Method [Version] description not found")
            return self.____interceptor(self.____stub.Version, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Version(request, timeout, metadata, credentials)
	
    def CrawlerOptions(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("CrawlerOptions")
            if not methodDesc:
                raise RuntimeError("Method [CrawlerOptions] description not found")
            return self.____interceptor(self.____stub.CrawlerOptions, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.CrawlerOptions(request, timeout, metadata, credentials)
	
    def AllowedDomains(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AllowedDomains")
            if not methodDesc:
                raise RuntimeError("Method [AllowedDomains] description not found")
            return self.____interceptor(self.____stub.AllowedDomains, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AllowedDomains(request, timeout, metadata, credentials)
	
    def CanonicalUrl(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("CanonicalUrl")
            if not methodDesc:
                raise RuntimeError("Method [CanonicalUrl] description not found")
            return self.____interceptor(self.____stub.CanonicalUrl, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.CanonicalUrl(request, timeout, metadata, credentials)
	
    def Parse(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Parse")
            if not methodDesc:
                raise RuntimeError("Method [Parse] description not found")
            return self.____interceptor(self.____stub.Parse, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Parse(request, timeout, metadata, credentials)
	
    def Call(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Call")
            if not methodDesc:
                raise RuntimeError("Method [Call] description not found")
            return self.____interceptor(self.____stub.Call, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Call(request, timeout, metadata, credentials)
	



from .service_pb2_grpc import CrawlerRegisterServicer as _CrawlerRegisterServicer, CrawlerRegisterStub as _CrawlerRegisterStub, add_CrawlerRegisterServicer_to_server

class CrawlerRegisterServicer(_CrawlerRegisterServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerRegister" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def Connect(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Connect")
                if not methodDesc:
                    raise RuntimeError("Method [Connect] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Connect)
            return self.servicer.Connect(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_CrawlerRegisterServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class CrawlerRegisterStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new CrawlerRegisterStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerRegister" ][0], FileDescriptor.package)
        self.____stub = _CrawlerRegisterStub(channel)
        self.____interceptor = interceptor

	
    def Connect(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Connect")
            if not methodDesc:
                raise RuntimeError("Method [Connect] description not found")
            return self.____interceptor(self.____stub.Connect, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Connect(request, timeout, metadata, credentials)
	



from .service_pb2_grpc import CrawlerManagerServicer as _CrawlerManagerServicer, CrawlerManagerStub as _CrawlerManagerStub, add_CrawlerManagerServicer_to_server

class CrawlerManagerServicer(_CrawlerManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def GetCrawlers(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawlers")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawlers] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawlers)
            return self.servicer.GetCrawlers(request, context)
	
        def GetCrawler(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawler")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawler] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawler)
            return self.servicer.GetCrawler(request, context)
	
        def GetCrawlerOptions(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawlerOptions")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawlerOptions] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawlerOptions)
            return self.servicer.GetCrawlerOptions(request, context)
	
        def GetCanonicalUrl(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCanonicalUrl")
                if not methodDesc:
                    raise RuntimeError("Method [GetCanonicalUrl] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCanonicalUrl)
            return self.servicer.GetCanonicalUrl(request, context)
	
        def DoParse(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DoParse")
                if not methodDesc:
                    raise RuntimeError("Method [DoParse] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DoParse)
            return self.servicer.DoParse(request, context)
	
        def RemoteCall(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("RemoteCall")
                if not methodDesc:
                    raise RuntimeError("Method [RemoteCall] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.RemoteCall)
            return self.servicer.RemoteCall(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_CrawlerManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class CrawlerManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new CrawlerManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "CrawlerManager" ][0], FileDescriptor.package)
        self.____stub = _CrawlerManagerStub(channel)
        self.____interceptor = interceptor

	
    def GetCrawlers(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawlers")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawlers] description not found")
            return self.____interceptor(self.____stub.GetCrawlers, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawlers(request, timeout, metadata, credentials)
	
    def GetCrawler(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawler")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawler] description not found")
            return self.____interceptor(self.____stub.GetCrawler, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawler(request, timeout, metadata, credentials)
	
    def GetCrawlerOptions(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawlerOptions")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawlerOptions] description not found")
            return self.____interceptor(self.____stub.GetCrawlerOptions, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawlerOptions(request, timeout, metadata, credentials)
	
    def GetCanonicalUrl(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCanonicalUrl")
            if not methodDesc:
                raise RuntimeError("Method [GetCanonicalUrl] description not found")
            return self.____interceptor(self.____stub.GetCanonicalUrl, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCanonicalUrl(request, timeout, metadata, credentials)
	
    def DoParse(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DoParse")
            if not methodDesc:
                raise RuntimeError("Method [DoParse] description not found")
            return self.____interceptor(self.____stub.DoParse, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DoParse(request, timeout, metadata, credentials)
	
    def RemoteCall(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("RemoteCall")
            if not methodDesc:
                raise RuntimeError("Method [RemoteCall] description not found")
            return self.____interceptor(self.____stub.RemoteCall, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.RemoteCall(request, timeout, metadata, credentials)
	



from .service_pb2_grpc import GatewayServicer as _GatewayServicer, GatewayStub as _GatewayStub, add_GatewayServicer_to_server

class GatewayServicer(_GatewayServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "Gateway" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def GetCrawlers(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawlers")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawlers] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawlers)
            return self.servicer.GetCrawlers(request, context)
	
        def GetCrawler(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawler")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawler] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawler)
            return self.servicer.GetCrawler(request, context)
	
        def GetCanonicalUrl(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCanonicalUrl")
                if not methodDesc:
                    raise RuntimeError("Method [GetCanonicalUrl] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCanonicalUrl)
            return self.servicer.GetCanonicalUrl(request, context)
	
        def Fetch(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Fetch")
                if not methodDesc:
                    raise RuntimeError("Method [Fetch] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Fetch)
            return self.servicer.Fetch(request, context)
	
        def GetRequest(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetRequest")
                if not methodDesc:
                    raise RuntimeError("Method [GetRequest] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetRequest)
            return self.servicer.GetRequest(request, context)
	
        def GetCrawlerLogs(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetCrawlerLogs")
                if not methodDesc:
                    raise RuntimeError("Method [GetCrawlerLogs] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetCrawlerLogs)
            return self.servicer.GetCrawlerLogs(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_GatewayServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class GatewayStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new GatewayStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "Gateway" ][0], FileDescriptor.package)
        self.____stub = _GatewayStub(channel)
        self.____interceptor = interceptor

	
    def GetCrawlers(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawlers")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawlers] description not found")
            return self.____interceptor(self.____stub.GetCrawlers, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawlers(request, timeout, metadata, credentials)
	
    def GetCrawler(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawler")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawler] description not found")
            return self.____interceptor(self.____stub.GetCrawler, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawler(request, timeout, metadata, credentials)
	
    def GetCanonicalUrl(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCanonicalUrl")
            if not methodDesc:
                raise RuntimeError("Method [GetCanonicalUrl] description not found")
            return self.____interceptor(self.____stub.GetCanonicalUrl, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCanonicalUrl(request, timeout, metadata, credentials)
	
    def Fetch(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Fetch")
            if not methodDesc:
                raise RuntimeError("Method [Fetch] description not found")
            return self.____interceptor(self.____stub.Fetch, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Fetch(request, timeout, metadata, credentials)
	
    def GetRequest(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetRequest")
            if not methodDesc:
                raise RuntimeError("Method [GetRequest] description not found")
            return self.____interceptor(self.____stub.GetRequest, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetRequest(request, timeout, metadata, credentials)
	
    def GetCrawlerLogs(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetCrawlerLogs")
            if not methodDesc:
                raise RuntimeError("Method [GetCrawlerLogs] description not found")
            return self.____interceptor(self.____stub.GetCrawlerLogs, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetCrawlerLogs(request, timeout, metadata, credentials)
	


