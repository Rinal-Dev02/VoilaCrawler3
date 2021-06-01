# -*- coding: UTF-8 -*-

from .header import Header
from context import Context
from network.url import URL
from parsel import Selector

class Request:
    pass

class Response(object):
    """ golang like http Response """

    def __init__(self, status_code, header, body, request):
        self._status_code = status_code
        self._header     = header or Header()
        self._body       = body
        self._request    = request
        self._selector   = None

    @property
    def request(self)->Request:
        """ request of this response """
        return self._request

    @property
    def status_code(self)->int:
        return self._status_code

    @property
    def header(self)->Header:
        return self._header

    @property
    def body(self)->bytes:
        return self._body

    @property
    def url(self)->URL:
        """ returns the url which generates the response """
        return self.request.url

    @property
    def rawurl(self)->URL:
        """ returns raw url start the request. which may different from the url function """
        req = self.request
        while req.response != None:
            req = req.response.req
        return req.url

    @property
    def context(self)->Context:
        return self._context or self._request.context or Context()

    @context.setter
    def context(self, val:Context):
        self._context = val

    def selector(self, encoding:str="utf8")->Selector:
        """
        returned a selector when the returned data is html content with default utf8 encoding.

        this selector is same as the selector in Scrapy.
        """

        if self._selector:
            return self._selector
        ctype = self._header.get("content-type")
        if self._body and ("text/html" in ctype or "application/xhtml+xml" in ctype or "application/xml" in ctype):
            self._selector = Selector(text=self._body.decode(encoding=encoding))
        else: self._selector = Selector(text='')

        return self._selector

class Request(object):
    """
    golang like http request
    """
    def __init__(self, ctx:Context, method:str, url:str, body:bytes=None, header:Header=Header(), resp:Response=None):
        self._ctx     = ctx
        self._method  = (method or "get").upper()
        self._url     = URL(url)
        self._body    = body
        self._headers = header
        self._response = resp

    @property
    def context(self)->Context:
        return self._ctx or Context()

    @context.setter
    def context(self, ctx:Context):
        self._ctx = ctx

    @property
    def method(self)->str:
        return self._method or "GET"

    @property
    def url(self)->URL:
        return self._url or URL()

    @property
    def headers(self)->Header:
        """ return header """
        return self._headers

    @property
    def body(self)->bytes:
        return self._body

    @property
    def response(self)->Response:
        """ parent response """
        return self._response or None
