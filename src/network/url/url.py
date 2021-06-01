# -*- coding: UTF-8 -*-

from urllib.parse import urlparse,urlencode,parse_qs

class Values(object):
    """ Values """

    def __init__(self, rawquery):
        if rawquery:
            self._vals = parse_qs(rawquery)
        else:
            self._vals = dict()

    def get(self, key:str)->str:
        vs = self._vals.get(key)
        if len(vs) > 0:
            return vs[0]
        return ""
    
    def add(self, key:str, val:str):
        self._vals[key] = self._vals.get(key) or list()
        self._vals[key].append(val)

    def set(self, key:str, val:str):
        self._vals[key] = [val]

    def delete(self, key:str):
        if self._vals.get(key):
            del self._vals[key]

    def encode(self):
        q = []
        for (k, vs) in self._vals.items():
            for v in vs:
                q.append((k,v,))
        return urlencode(q)

class UserInfo(object):
    
    def __init__(self, username:str, password:str=None):
        self.username = username
        self.password = password or ""

    @property
    def username(self):
        return self._username

    @username.setter
    def username(self, val:str):
        self._username = val

    @property
    def password(self):
        return self._password

    @password.setter
    def password(self, val:str):
        self._password = val

    def encode(self)->str:
        """ encode """
        if not self.username:
            return ""
        return "{}:{}".format(self.username, self.password)

class URL(object):
    """ URL """

    def __init__(self, rawurl):
        self._u = urlparse(rawurl)
        self.scheme = self._u.scheme
        self.userinfo = UserInfo(self._u.username, self._u.password)
        if self._u.port:
            self.host = "{}:{}".format(self._u.hostname, self._u.port)
        else:
            self.host = self._u.hostname
        self.path = self._u.path
        self.raw_query = self._u.query
        self.fragment = self._u.fragment

    @property
    def scheme(self)->str:
        return self._scheme
    
    @scheme.setter
    def scheme(self, val)->str:
        self._scheme = val

    @property
    def userinfo(self)->UserInfo:
        return self._userinfo

    @userinfo.setter
    def userinfo(self, val:UserInfo):
        self._userinfo = val

    @property
    def host(self)->str:
        return self._host
    
    @host.setter
    def host(self, val:str):
        self._host = val

    @property
    def hostname(self)->str:
        return self._u.hostname

    @property
    def port(self)->str:
        return self._u.port
    
    @property
    def path(self)->str:
        return self._path

    @path.setter
    def path(self, val:str):
        self._path = val
    
    @property
    def raw_query(self)->str:
        return self._raw_query
    
    @raw_query.setter
    def raw_query(self, val):
        self._raw_query = val

    def query(self)->Values:
        return Values(self._raw_query)

    @property
    def fragment(self)->str:
        return self._fragment

    @fragment.setter
    def fragment(self, val:str):
        self._fragment = val

    @property
    def path(self)->str:
        return self._path
    
    @path.setter
    def path(self, val:str):
        self._path = val

    def __str__(self) -> str:
        u = self._host + self._path
        if self._raw_query:
            u = u +  "?" + self._raw_query
        if self._fragment:
            u = u + "#" + self._fragment
        if self._userinfo:
            info = self._userinfo.encode()
            if info:
                u =  + "@" + u
        if self._scheme:
            u = self._scheme + "://" + u
        return u
        