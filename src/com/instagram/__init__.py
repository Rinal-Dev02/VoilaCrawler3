#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

from chameleon.api import http
from crawler import Crawler
from proxy import ProxyClient

class Instagram(Crawler):
    def __init__(self, httpClient:ProxyClient):
        self._id = "xxx"
        super(Instagram, self).__init__(id)
        self._httpClient = httpClient
    
    def ID(self):
        return super().ID()
