# encoding=utf8
# extension plugin generated codes. DO NOT EDIT!!!

from google.protobuf.descriptor_pb2 import FileDescriptorProto

from chameleon.api.invocation import ServiceDesc

from .service_pb2 import DESCRIPTOR

FileDescriptor = FileDescriptorProto()
DESCRIPTOR.CopyToProto(FileDescriptor)



from .service_pb2_grpc import RoleManagerServicer as _RoleManagerServicer, RoleManagerStub as _RoleManagerStub, add_RoleManagerServicer_to_server

class RoleManagerServicer(_RoleManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "RoleManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def Get(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Get")
                if not methodDesc:
                    raise RuntimeError("Method [Get] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Get)
            return self.servicer.Get(request, context)
	
        def List(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("List")
                if not methodDesc:
                    raise RuntimeError("Method [List] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.List)
            return self.servicer.List(request, context)
	
        def Create(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Create")
                if not methodDesc:
                    raise RuntimeError("Method [Create] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Create)
            return self.servicer.Create(request, context)
	
        def Update(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Update")
                if not methodDesc:
                    raise RuntimeError("Method [Update] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Update)
            return self.servicer.Update(request, context)
	
        def AddRule(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AddRule")
                if not methodDesc:
                    raise RuntimeError("Method [AddRule] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AddRule)
            return self.servicer.AddRule(request, context)
	
        def ResetRule(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("ResetRule")
                if not methodDesc:
                    raise RuntimeError("Method [ResetRule] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.ResetRule)
            return self.servicer.ResetRule(request, context)
	
        def DeleteRule(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DeleteRule")
                if not methodDesc:
                    raise RuntimeError("Method [DeleteRule] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DeleteRule)
            return self.servicer.DeleteRule(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_RoleManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class RoleManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new RoleManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "RoleManager" ][0], FileDescriptor.package)
        self.____stub = _RoleManagerStub(channel)
        self.____interceptor = interceptor

	
    def Get(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Get")
            if not methodDesc:
                raise RuntimeError("Method [Get] description not found")
            return self.____interceptor(self.____stub.Get, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Get(request, timeout, metadata, credentials)
	
    def List(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("List")
            if not methodDesc:
                raise RuntimeError("Method [List] description not found")
            return self.____interceptor(self.____stub.List, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.List(request, timeout, metadata, credentials)
	
    def Create(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Create")
            if not methodDesc:
                raise RuntimeError("Method [Create] description not found")
            return self.____interceptor(self.____stub.Create, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Create(request, timeout, metadata, credentials)
	
    def Update(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Update")
            if not methodDesc:
                raise RuntimeError("Method [Update] description not found")
            return self.____interceptor(self.____stub.Update, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Update(request, timeout, metadata, credentials)
	
    def AddRule(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AddRule")
            if not methodDesc:
                raise RuntimeError("Method [AddRule] description not found")
            return self.____interceptor(self.____stub.AddRule, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AddRule(request, timeout, metadata, credentials)
	
    def ResetRule(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("ResetRule")
            if not methodDesc:
                raise RuntimeError("Method [ResetRule] description not found")
            return self.____interceptor(self.____stub.ResetRule, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.ResetRule(request, timeout, metadata, credentials)
	
    def DeleteRule(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DeleteRule")
            if not methodDesc:
                raise RuntimeError("Method [DeleteRule] description not found")
            return self.____interceptor(self.____stub.DeleteRule, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DeleteRule(request, timeout, metadata, credentials)
	



from .service_pb2_grpc import UserManagerServicer as _UserManagerServicer, UserManagerStub as _UserManagerStub, add_UserManagerServicer_to_server

class UserManagerServicer(_UserManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "UserManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def Count(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Count")
                if not methodDesc:
                    raise RuntimeError("Method [Count] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Count)
            return self.servicer.Count(request, context)
	
        def List(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("List")
                if not methodDesc:
                    raise RuntimeError("Method [List] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.List)
            return self.servicer.List(request, context)
	
        def Who(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Who")
                if not methodDesc:
                    raise RuntimeError("Method [Who] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Who)
            return self.servicer.Who(request, context)
	
        def Exist(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Exist")
                if not methodDesc:
                    raise RuntimeError("Method [Exist] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Exist)
            return self.servicer.Exist(request, context)
	
        def Exists(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Exists")
                if not methodDesc:
                    raise RuntimeError("Method [Exists] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Exists)
            return self.servicer.Exists(request, context)
	
        def Get(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Get")
                if not methodDesc:
                    raise RuntimeError("Method [Get] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Get)
            return self.servicer.Get(request, context)
	
        def Gets(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Gets")
                if not methodDesc:
                    raise RuntimeError("Method [Gets] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Gets)
            return self.servicer.Gets(request, context)
	
        def Create(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Create")
                if not methodDesc:
                    raise RuntimeError("Method [Create] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Create)
            return self.servicer.Create(request, context)
	
        def Update(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Update")
                if not methodDesc:
                    raise RuntimeError("Method [Update] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Update)
            return self.servicer.Update(request, context)
	
        def UpdateField(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("UpdateField")
                if not methodDesc:
                    raise RuntimeError("Method [UpdateField] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.UpdateField)
            return self.servicer.UpdateField(request, context)
	
        def GetRoles(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetRoles")
                if not methodDesc:
                    raise RuntimeError("Method [GetRoles] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetRoles)
            return self.servicer.GetRoles(request, context)
	
        def AddRole(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AddRole")
                if not methodDesc:
                    raise RuntimeError("Method [AddRole] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AddRole)
            return self.servicer.AddRole(request, context)
	
        def DeleteRole(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DeleteRole")
                if not methodDesc:
                    raise RuntimeError("Method [DeleteRole] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DeleteRole)
            return self.servicer.DeleteRole(request, context)
	
        def Delete(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Delete")
                if not methodDesc:
                    raise RuntimeError("Method [Delete] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Delete)
            return self.servicer.Delete(request, context)
	
        def Restore(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Restore")
                if not methodDesc:
                    raise RuntimeError("Method [Restore] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Restore)
            return self.servicer.Restore(request, context)
	
        def Verify(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Verify")
                if not methodDesc:
                    raise RuntimeError("Method [Verify] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Verify)
            return self.servicer.Verify(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_UserManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class UserManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new UserManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "UserManager" ][0], FileDescriptor.package)
        self.____stub = _UserManagerStub(channel)
        self.____interceptor = interceptor

	
    def Count(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Count")
            if not methodDesc:
                raise RuntimeError("Method [Count] description not found")
            return self.____interceptor(self.____stub.Count, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Count(request, timeout, metadata, credentials)
	
    def List(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("List")
            if not methodDesc:
                raise RuntimeError("Method [List] description not found")
            return self.____interceptor(self.____stub.List, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.List(request, timeout, metadata, credentials)
	
    def Who(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Who")
            if not methodDesc:
                raise RuntimeError("Method [Who] description not found")
            return self.____interceptor(self.____stub.Who, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Who(request, timeout, metadata, credentials)
	
    def Exist(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Exist")
            if not methodDesc:
                raise RuntimeError("Method [Exist] description not found")
            return self.____interceptor(self.____stub.Exist, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Exist(request, timeout, metadata, credentials)
	
    def Exists(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Exists")
            if not methodDesc:
                raise RuntimeError("Method [Exists] description not found")
            return self.____interceptor(self.____stub.Exists, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Exists(request, timeout, metadata, credentials)
	
    def Get(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Get")
            if not methodDesc:
                raise RuntimeError("Method [Get] description not found")
            return self.____interceptor(self.____stub.Get, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Get(request, timeout, metadata, credentials)
	
    def Gets(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Gets")
            if not methodDesc:
                raise RuntimeError("Method [Gets] description not found")
            return self.____interceptor(self.____stub.Gets, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Gets(request, timeout, metadata, credentials)
	
    def Create(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Create")
            if not methodDesc:
                raise RuntimeError("Method [Create] description not found")
            return self.____interceptor(self.____stub.Create, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Create(request, timeout, metadata, credentials)
	
    def Update(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Update")
            if not methodDesc:
                raise RuntimeError("Method [Update] description not found")
            return self.____interceptor(self.____stub.Update, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Update(request, timeout, metadata, credentials)
	
    def UpdateField(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("UpdateField")
            if not methodDesc:
                raise RuntimeError("Method [UpdateField] description not found")
            return self.____interceptor(self.____stub.UpdateField, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.UpdateField(request, timeout, metadata, credentials)
	
    def GetRoles(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetRoles")
            if not methodDesc:
                raise RuntimeError("Method [GetRoles] description not found")
            return self.____interceptor(self.____stub.GetRoles, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetRoles(request, timeout, metadata, credentials)
	
    def AddRole(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AddRole")
            if not methodDesc:
                raise RuntimeError("Method [AddRole] description not found")
            return self.____interceptor(self.____stub.AddRole, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AddRole(request, timeout, metadata, credentials)
	
    def DeleteRole(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DeleteRole")
            if not methodDesc:
                raise RuntimeError("Method [DeleteRole] description not found")
            return self.____interceptor(self.____stub.DeleteRole, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DeleteRole(request, timeout, metadata, credentials)
	
    def Delete(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Delete")
            if not methodDesc:
                raise RuntimeError("Method [Delete] description not found")
            return self.____interceptor(self.____stub.Delete, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Delete(request, timeout, metadata, credentials)
	
    def Restore(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Restore")
            if not methodDesc:
                raise RuntimeError("Method [Restore] description not found")
            return self.____interceptor(self.____stub.Restore, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Restore(request, timeout, metadata, credentials)
	
    def Verify(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Verify")
            if not methodDesc:
                raise RuntimeError("Method [Verify] description not found")
            return self.____interceptor(self.____stub.Verify, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Verify(request, timeout, metadata, credentials)
	



from .service_pb2_grpc import ApplicationManagerServicer as _ApplicationManagerServicer, ApplicationManagerStub as _ApplicationManagerStub, add_ApplicationManagerServicer_to_server

class ApplicationManagerServicer(_ApplicationManagerServicer):
    class Wrapper(object):
        """The servicer wrapper object
        """
        def __init__(self, servicer, interceptor=None):
            """Create a new Wrapper
            """
            self.desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "ApplicationManager" ][0], FileDescriptor.package)
            self.servicer = servicer
            self.interceptor = interceptor

	
        def Count(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Count")
                if not methodDesc:
                    raise RuntimeError("Method [Count] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Count)
            return self.servicer.Count(request, context)
	
        def List(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("List")
                if not methodDesc:
                    raise RuntimeError("Method [List] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.List)
            return self.servicer.List(request, context)
	
        def Exist(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Exist")
                if not methodDesc:
                    raise RuntimeError("Method [Exist] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Exist)
            return self.servicer.Exist(request, context)
	
        def Exists(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Exists")
                if not methodDesc:
                    raise RuntimeError("Method [Exists] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Exists)
            return self.servicer.Exists(request, context)
	
        def Get(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Get")
                if not methodDesc:
                    raise RuntimeError("Method [Get] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Get)
            return self.servicer.Get(request, context)
	
        def Gets(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Gets")
                if not methodDesc:
                    raise RuntimeError("Method [Gets] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Gets)
            return self.servicer.Gets(request, context)
	
        def Create(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Create")
                if not methodDesc:
                    raise RuntimeError("Method [Create] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Create)
            return self.servicer.Create(request, context)
	
        def Update(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Update")
                if not methodDesc:
                    raise RuntimeError("Method [Update] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Update)
            return self.servicer.Update(request, context)
	
        def Delete(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Delete")
                if not methodDesc:
                    raise RuntimeError("Method [Delete] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Delete)
            return self.servicer.Delete(request, context)
	
        def Restore(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("Restore")
                if not methodDesc:
                    raise RuntimeError("Method [Restore] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.Restore)
            return self.servicer.Restore(request, context)
	
        def SetOptions(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("SetOptions")
                if not methodDesc:
                    raise RuntimeError("Method [SetOptions] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.SetOptions)
            return self.servicer.SetOptions(request, context)
	
        def SetTags(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("SetTags")
                if not methodDesc:
                    raise RuntimeError("Method [SetTags] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.SetTags)
            return self.servicer.SetTags(request, context)
	
        def ResetSecret(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("ResetSecret")
                if not methodDesc:
                    raise RuntimeError("Method [ResetSecret] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.ResetSecret)
            return self.servicer.ResetSecret(request, context)
	
        def GetSecretKey(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetSecretKey")
                if not methodDesc:
                    raise RuntimeError("Method [GetSecretKey] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetSecretKey)
            return self.servicer.GetSecretKey(request, context)
	
        def GetSecretPublicKey(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetSecretPublicKey")
                if not methodDesc:
                    raise RuntimeError("Method [GetSecretPublicKey] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetSecretPublicKey)
            return self.servicer.GetSecretPublicKey(request, context)
	
        def AddSecretKey(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AddSecretKey")
                if not methodDesc:
                    raise RuntimeError("Method [AddSecretKey] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AddSecretKey)
            return self.servicer.AddSecretKey(request, context)
	
        def SetDefaultSecretKey(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("SetDefaultSecretKey")
                if not methodDesc:
                    raise RuntimeError("Method [SetDefaultSecretKey] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.SetDefaultSecretKey)
            return self.servicer.SetDefaultSecretKey(request, context)
	
        def DeleteSecretKey(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DeleteSecretKey")
                if not methodDesc:
                    raise RuntimeError("Method [DeleteSecretKey] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DeleteSecretKey)
            return self.servicer.DeleteSecretKey(request, context)
	
        def GetDefaultRedirectURI(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetDefaultRedirectURI")
                if not methodDesc:
                    raise RuntimeError("Method [GetDefaultRedirectURI] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetDefaultRedirectURI)
            return self.servicer.GetDefaultRedirectURI(request, context)
	
        def SetDefaultRedirectURI(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("SetDefaultRedirectURI")
                if not methodDesc:
                    raise RuntimeError("Method [SetDefaultRedirectURI] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.SetDefaultRedirectURI)
            return self.servicer.SetDefaultRedirectURI(request, context)
	
        def GetWhiteRedirectURIs(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("GetWhiteRedirectURIs")
                if not methodDesc:
                    raise RuntimeError("Method [GetWhiteRedirectURIs] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.GetWhiteRedirectURIs)
            return self.servicer.GetWhiteRedirectURIs(request, context)
	
        def AddWhiteRedirectURI(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AddWhiteRedirectURI")
                if not methodDesc:
                    raise RuntimeError("Method [AddWhiteRedirectURI] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AddWhiteRedirectURI)
            return self.servicer.AddWhiteRedirectURI(request, context)
	
        def DeleteWhiteRedirectURI(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("DeleteWhiteRedirectURI")
                if not methodDesc:
                    raise RuntimeError("Method [DeleteWhiteRedirectURI] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.DeleteWhiteRedirectURI)
            return self.servicer.DeleteWhiteRedirectURI(request, context)
	
        def ClearWhiteRedirectURIs(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("ClearWhiteRedirectURIs")
                if not methodDesc:
                    raise RuntimeError("Method [ClearWhiteRedirectURIs] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.ClearWhiteRedirectURIs)
            return self.servicer.ClearWhiteRedirectURIs(request, context)
	
        def AddApplicationRole(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("AddApplicationRole")
                if not methodDesc:
                    raise RuntimeError("Method [AddApplicationRole] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.AddApplicationRole)
            return self.servicer.AddApplicationRole(request, context)
	
        def RemoveApplicationRole(self, request, context):
            if self.interceptor:
                methodDesc = self.desc.methods.get("RemoveApplicationRole")
                if not methodDesc:
                    raise RuntimeError("Method [RemoveApplicationRole] description not found")
                return self.interceptor(request, context, methodDesc, self.servicer.RemoveApplicationRole)
            return self.servicer.RemoveApplicationRole(request, context)
	

    @classmethod
    def addToServer(cls, servicer, server, *args, **kwargs):
        """Add this servicer to server
        Args:
            servicer(object): The servicer
            server(gRPC server): The gRPC server
        """
        return add_ApplicationManagerServicer_to_server(cls.Wrapper(servicer, *args, **kwargs), server)

class ApplicationManagerStub(object):
    def __init__(self, channel, interceptor = None):
        """Create a new ApplicationManagerStub object
        """
        self.____desc = ServiceDesc([ x for x in FileDescriptor.service if x.name == "ApplicationManager" ][0], FileDescriptor.package)
        self.____stub = _ApplicationManagerStub(channel)
        self.____interceptor = interceptor

	
    def Count(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Count")
            if not methodDesc:
                raise RuntimeError("Method [Count] description not found")
            return self.____interceptor(self.____stub.Count, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Count(request, timeout, metadata, credentials)
	
    def List(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("List")
            if not methodDesc:
                raise RuntimeError("Method [List] description not found")
            return self.____interceptor(self.____stub.List, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.List(request, timeout, metadata, credentials)
	
    def Exist(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Exist")
            if not methodDesc:
                raise RuntimeError("Method [Exist] description not found")
            return self.____interceptor(self.____stub.Exist, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Exist(request, timeout, metadata, credentials)
	
    def Exists(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Exists")
            if not methodDesc:
                raise RuntimeError("Method [Exists] description not found")
            return self.____interceptor(self.____stub.Exists, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Exists(request, timeout, metadata, credentials)
	
    def Get(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Get")
            if not methodDesc:
                raise RuntimeError("Method [Get] description not found")
            return self.____interceptor(self.____stub.Get, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Get(request, timeout, metadata, credentials)
	
    def Gets(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Gets")
            if not methodDesc:
                raise RuntimeError("Method [Gets] description not found")
            return self.____interceptor(self.____stub.Gets, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Gets(request, timeout, metadata, credentials)
	
    def Create(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Create")
            if not methodDesc:
                raise RuntimeError("Method [Create] description not found")
            return self.____interceptor(self.____stub.Create, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Create(request, timeout, metadata, credentials)
	
    def Update(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Update")
            if not methodDesc:
                raise RuntimeError("Method [Update] description not found")
            return self.____interceptor(self.____stub.Update, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Update(request, timeout, metadata, credentials)
	
    def Delete(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Delete")
            if not methodDesc:
                raise RuntimeError("Method [Delete] description not found")
            return self.____interceptor(self.____stub.Delete, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Delete(request, timeout, metadata, credentials)
	
    def Restore(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("Restore")
            if not methodDesc:
                raise RuntimeError("Method [Restore] description not found")
            return self.____interceptor(self.____stub.Restore, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.Restore(request, timeout, metadata, credentials)
	
    def SetOptions(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("SetOptions")
            if not methodDesc:
                raise RuntimeError("Method [SetOptions] description not found")
            return self.____interceptor(self.____stub.SetOptions, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.SetOptions(request, timeout, metadata, credentials)
	
    def SetTags(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("SetTags")
            if not methodDesc:
                raise RuntimeError("Method [SetTags] description not found")
            return self.____interceptor(self.____stub.SetTags, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.SetTags(request, timeout, metadata, credentials)
	
    def ResetSecret(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("ResetSecret")
            if not methodDesc:
                raise RuntimeError("Method [ResetSecret] description not found")
            return self.____interceptor(self.____stub.ResetSecret, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.ResetSecret(request, timeout, metadata, credentials)
	
    def GetSecretKey(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetSecretKey")
            if not methodDesc:
                raise RuntimeError("Method [GetSecretKey] description not found")
            return self.____interceptor(self.____stub.GetSecretKey, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetSecretKey(request, timeout, metadata, credentials)
	
    def GetSecretPublicKey(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetSecretPublicKey")
            if not methodDesc:
                raise RuntimeError("Method [GetSecretPublicKey] description not found")
            return self.____interceptor(self.____stub.GetSecretPublicKey, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetSecretPublicKey(request, timeout, metadata, credentials)
	
    def AddSecretKey(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AddSecretKey")
            if not methodDesc:
                raise RuntimeError("Method [AddSecretKey] description not found")
            return self.____interceptor(self.____stub.AddSecretKey, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AddSecretKey(request, timeout, metadata, credentials)
	
    def SetDefaultSecretKey(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("SetDefaultSecretKey")
            if not methodDesc:
                raise RuntimeError("Method [SetDefaultSecretKey] description not found")
            return self.____interceptor(self.____stub.SetDefaultSecretKey, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.SetDefaultSecretKey(request, timeout, metadata, credentials)
	
    def DeleteSecretKey(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DeleteSecretKey")
            if not methodDesc:
                raise RuntimeError("Method [DeleteSecretKey] description not found")
            return self.____interceptor(self.____stub.DeleteSecretKey, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DeleteSecretKey(request, timeout, metadata, credentials)
	
    def GetDefaultRedirectURI(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetDefaultRedirectURI")
            if not methodDesc:
                raise RuntimeError("Method [GetDefaultRedirectURI] description not found")
            return self.____interceptor(self.____stub.GetDefaultRedirectURI, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetDefaultRedirectURI(request, timeout, metadata, credentials)
	
    def SetDefaultRedirectURI(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("SetDefaultRedirectURI")
            if not methodDesc:
                raise RuntimeError("Method [SetDefaultRedirectURI] description not found")
            return self.____interceptor(self.____stub.SetDefaultRedirectURI, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.SetDefaultRedirectURI(request, timeout, metadata, credentials)
	
    def GetWhiteRedirectURIs(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("GetWhiteRedirectURIs")
            if not methodDesc:
                raise RuntimeError("Method [GetWhiteRedirectURIs] description not found")
            return self.____interceptor(self.____stub.GetWhiteRedirectURIs, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.GetWhiteRedirectURIs(request, timeout, metadata, credentials)
	
    def AddWhiteRedirectURI(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AddWhiteRedirectURI")
            if not methodDesc:
                raise RuntimeError("Method [AddWhiteRedirectURI] description not found")
            return self.____interceptor(self.____stub.AddWhiteRedirectURI, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AddWhiteRedirectURI(request, timeout, metadata, credentials)
	
    def DeleteWhiteRedirectURI(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("DeleteWhiteRedirectURI")
            if not methodDesc:
                raise RuntimeError("Method [DeleteWhiteRedirectURI] description not found")
            return self.____interceptor(self.____stub.DeleteWhiteRedirectURI, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.DeleteWhiteRedirectURI(request, timeout, metadata, credentials)
	
    def ClearWhiteRedirectURIs(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("ClearWhiteRedirectURIs")
            if not methodDesc:
                raise RuntimeError("Method [ClearWhiteRedirectURIs] description not found")
            return self.____interceptor(self.____stub.ClearWhiteRedirectURIs, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.ClearWhiteRedirectURIs(request, timeout, metadata, credentials)
	
    def AddApplicationRole(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("AddApplicationRole")
            if not methodDesc:
                raise RuntimeError("Method [AddApplicationRole] description not found")
            return self.____interceptor(self.____stub.AddApplicationRole, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.AddApplicationRole(request, timeout, metadata, credentials)
	
    def RemoveApplicationRole(self, request, timeout=None, metadata=None, credentials=None):
        if self.____interceptor:
            methodDesc = self.____desc.methods.get("RemoveApplicationRole")
            if not methodDesc:
                raise RuntimeError("Method [RemoveApplicationRole] description not found")
            return self.____interceptor(self.____stub.RemoveApplicationRole, methodDesc, request, timeout, metadata, credentials)
        return self.____stub.RemoveApplicationRole(request, timeout, metadata, credentials)
	


