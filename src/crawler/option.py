# -*- coding: UTF-8 -*-

from chameleon.api.http import Cookie,GET,POST,PUT,PATCH,DELETE,OPTION
from chameleon.smelter.v1.crawl.proxy import ProxyReliability,ReliabilityDefault,ReliabilityLow,ReliabilityMedium,ReliabilityHigh,ReliabilityIntelligent

class CrawlerOptions(object):
    """ CrawlerOptions """

    def __init__(self):
        self.enableHeadless    = False
        self.enableSessionInit = False
        self.keepSession       = False
        self.sessionTtl        = 0
        self.disableCookieJar  = False
        self.disableRedirect   = False
        self.headers        = dict()
        self.cookies        = list()
        self.reliability       = ReliabilityDefault

    @property
    def enableHeadless(self) -> bool:
        return self._enableHeadless or False

    @enableHeadless.setter
    def enableHeadless(self, val:bool):
        self._enableHeadless = val

    @property
    def enableSessionInit(self) -> bool:
        return self._enableSessionInit or False

    @enableSessionInit.setter
    def enableSessionInit(self, val:bool):
        if not isinstance(val, bool):
            raise TypeError("boolen value expected")
        self._enableSessionInit = val

    @property
    def keepSession(self)->bool:
        return self._keepSession or False

    @keepSession.setter
    def keepSession(self, val:bool):
        if not isinstance(val, bool):
            raise TypeError("boolen value expected")
        self._keepSession = val

    @property
    def sessionTTL(self):
        return self._sessionTTL or 0

    @sessionTTL.setter
    def keepSession(self, val:int):
        self._sessionTTL = val

    @property
    def disableCookieJar(self)->bool:
        return self._disableCookieJar or False

    @disableCookieJar.setter
    def disableCookieJar(self, val:bool):
        self._disableCookieJar = val

    @property
    def disableRedirect(self)->bool:
        return self._disableRedirect or False

    @disableRedirect.setter
    def disableRedirect(self, val:bool):
        self._disableRedirect = val

    @property
    def disableRedirect(self)->bool:
        return self._disableRedirect or False

    @disableRedirect.setter
    def disableRedirect(self, val:bool):
        self._disableRedirect = val

    @property
    def headers(self)->dict:
        return self._headers or dict()

    @headers.setter
    def headers(self, val:dict):
        self._headers = val

    def getHeader(self, key:str)->str:
        if not self._headers:
            return ""
        return self._headers.get(key) or ""

    def setHeader(self, key:str, val:str):
        if not self._headers:
            self._headers = dict()
        self._headers[key] = val

    def delHeader(self, key:str):
        if not self._headers:
            return
        if self._headers.get(key):
            del self._headers[key]

    @property
    def cookies(self)->list[Cookie]:
        return self._cookies or list()

    @cookies.setter
    def cookies(self, val:list[Cookie]):
        self._cookies = val

    def addCookie(self, name:str, val:str, domain:str="", path:str=""):
        """ add cookie """
        if not name:
            raise ValueError("invalid cookie name")
        cookie = Cookie()
        cookie.name = name
        cookie.value = val
        cookie.domain = domain
        cookie.path = path
        if self._cookies == None:
            self._cookies = list()
        self._cookies.append(cookie)

    @property
    def reliability(self)->ProxyReliability:
        return self._reliability or ReliabilityDefault

    @reliability.setter
    def reliability(self, val:ProxyReliability):
        self._reliability = val