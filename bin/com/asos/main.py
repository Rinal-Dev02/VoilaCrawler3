#!/usr/bin/env python3.9
# -*- coding: UTF-8 -*-

import sys
from os import path

projectPath = path.abspath(path.normpath(path.dirname(__file__)))
index = projectPath.rfind("/bin")
if index < 0:
    index = projectPath.rfind("/releases")
importpath = projectPath[0:index] + "/src"
if path.isdir(importpath):
    sys.path[0] = importpath

from app import getArgs, Application
from proxy import ProxyClient
from crawler import Crawler
from com.asos import ASOS

args = getArgs()

def newCrawler(httpClient:ProxyClient)->Crawler:
    return ASOS(httpClient, label="test")

app = Application(args, newCrawler)
app.run()
