# -*- coding: UTF-8 -*-

from abc import ABC, abstractmethod
from typing import Any, Generator

from context import Context
from network.url import URL
from network.http.request import Response
from proxy.proxy import RequestOptions

from .option import CrawlerOptions
from .error import Error, Unimplemented
from network.http import Request
from chameleon.smelter.v1.crawl import Item,Error

# Crawler
class Crawler(ABC):

    def __init__(self, id:str, version:int, allowedDomains:list[str], options:CrawlerOptions):
        if not id:
            raise ValueError("invalid crawler binding id")
        if version < 0:
            raise ValueError("invalid crawler number version")
        if not allowedDomains:
            raise ValueError("crawler allowed domains required")
        if not options:
            raise ValueError("crawler options required")
        super(Crawler, self).__init__()

        self._id = id
        self._version = version
        self._allowedDomains = allowedDomains
        self._options = options

    def ID(self):
        """
        returns crawler unique id, this commonly should be the hosted id of this site called store Id.
        """
        return self._id

    def Version(self):
        """
        the version of current this crawler, which should be an active number.
        """
        return self._version

    def CrawlOptions(self, u:URL)->CrawlerOptions:
        """
        crawler action requirement, if need to get options by url, overwrite this function
        """
        return self._options

    def AllowedDomains(self)->list:
        """
        the domains this crawler supportes
        """
        return self._allowedDomains

    def CanonicalUrl(self, u:str)->str:
        """
        returns canonical url the proviced url
        """
        return u

    @abstractmethod
    def Parse(self, ctx:Context, resp:Response)->Generator[Any,None,None]:
        """
        used to parse http request parse.
            param ctx used to share info between parent and child. and it can set the max ttl for parse job.
            param resp represents the http response, with act as a real http response.
        this function will yield results
        """
        yield Error(msg="Parse function not implemented", code=Unimplemented)

    @abstractmethod
    def NewTestRequest(self, ctx:Context) -> Generator[Request,None,None]:
        pass

    @abstractmethod
    def CheckTestResponse(self, ctx:Context, resp:Response)->bool:
        return True