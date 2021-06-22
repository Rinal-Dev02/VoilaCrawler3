#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

from json.decoder import JSONDecodeError
import re
import json
import logging
import traceback

from typing import Any, Generator, List
from chameleon.smelter.v1.crawl.item.product_pb2 import Sku, SkuSpecOption
from context.context import Context
from crawler.context import IndexKey, TracingIdKey
from crawler import Crawler, CrawlerOptions, Error, ErrAbort, ErrUnsupportedPath
from network.url import URL
from network.http import Request,Response
from proxy import RequestOptions, ProxyClient
from util.proto import newImageMedia
from util.random import newRandomId

from chameleon.api.regulation import USD
from chameleon.smelter.v1.crawl.proxy import ProxyReliability, ReliabilityDefault, ReliabilityLow, ReliabilityMedium, ReliabilityHigh, ReliabilityIntelligent
from chameleon.smelter.v1.crawl.item import Product, SkuSpecColor, SkuSpecSize, Stock

class ASOS(Crawler):
    logger = logging.getLogger("com.asos")

    def __init__(self, httpClient:ProxyClient, **kwargs):
        options = CrawlerOptions()
        options.enableHeadless = True
        options.enableSessionInit = True
        options.reliability = ReliabilityMedium

        options.addCookie("geocountry", "US", path="/")
        options.addCookie("browseCountry", "US", path="/")
        options.addCookie("browseCurrency", "USD", path="/")
        options.addCookie("browseLanguage", "en-US", path="/")
        options.addCookie("browseSizeSchema", "US", path="/")
        options.addCookie("storeCode", "US", path="/")
        options.addCookie("currency", "2", path="/")

        super(ASOS, self).__init__("701fdaa85a5a18866ccbb357ad2ccff9", 1, ["*.asos.com"], options)

        self._httpClient = httpClient
        self._categoryPathMatcher = re.compile('^(/[a-z0-9_-]+)?/(women|men)(/[a-z0-9_-]+){1,6}/cat/?$')
        self._categoryJsonMatcher = re.compile('^/api/product/search/v2/categories/([a-z0-9]+)')
        self._productGroupMatcher = re.compile('^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/grp/[0-9]+/?$')
        self._productPathMatcher = re.compile('^(/[a-z0-9_-]+)?(/[a-z0-9_-]+){2}/prd/[0-9]+/?$')

    def CanonicalUrl(self, rawurl:str) -> str:
        """ get canonical url of provided url """
        u = URL(rawurl)
        if not u.scheme:
            u.scheme = "https"
        if not u.host:
            u.host = "www.asos.com"
        if self._productPathMatcher.match(u.path) or self._productGroupMatcher.match(u.path):
            u.raw_query = ""
            u.fragment = ""
            return str(u)
        return rawurl

    def Parse(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        path = resp.rawurl.path.rstrip("/")
        if path=="" or path == "/us/women" or path == "/us/men":
            yield from self.parseCategories(ctx, resp)
        elif self._categoryPathMatcher.match(path):
            yield from self.parseProductsHTML(ctx, resp)
        elif self._categoryJsonMatcher.match(path):
            yield from self.parseProductsJSON(ctx, resp)
        elif self._productGroupMatcher.match(path):
            yield from self.parseProductGroup(ctx, resp)
        elif self._productPathMatcher.match(path):
            yield from self.parseProduct(ctx, resp)
        return ErrUnsupportedPath

    def parseCategories(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        for node in resp.selector().css('#chrome-sticky-header nav[data-testid="primarynav-large"] button[data-id]'):
            dataid = node.attrib.get("data-id")
            if not dataid: continue
            cate = node.xpath("text()").extract_first()

            links = resp.selector().css('#{0} ul[data-id="{0}"]>li>ul>li>a[href]'.format(dataid))
            for link in links:
                subCate = link.xpath("text()").extract_first()
                href = link.attrib.get("href")
                if not href: continue

                try:
                    u = URL(href)
                    if "/gift-vouchers" in u.path: continue

                    mainCate = "women"
                    if u.path.startswith("/us/men"):
                        mainCate = "men"
                    if self._categoryPathMatcher.match(u.path):
                        nctx = Context(ctx, TracingIdKey, newRandomId())
                        nctx = Context(nctx, "MainCategory", mainCate)
                        nctx = Context(nctx, "Category", cate)
                        nctx = Context(nctx, "SubCategory", subCate)

                        req = Request(ctx,"GET", str(u))
                        yield nctx, req
                except:
                    self.logger.error("parse url %s fialed", href)
            
    productsDataReg = re.compile("window\.asos\.plp\._data\s*=\s*JSON\.parse\('(.*?)'\);", re.I|re.U)

    def parseProductsHTML(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        matched = ASOS.productsDataReg.findall(resp.body.decode())
        if len(matched) < 1:
            raise Error("extract json from product list {} failed".format(resp.rawurl))
        r = json.loads(matched[0])

        lastIndex = ctx.get_int(IndexKey)+1
        cid = str(r.get("search", {}).get("query", {}).get("cid", ""))
        for prod in r.get("search", {}).get("products", list()):
            href = "/us{}i&cid={}".format(prod.get("url"), cid)
            try:
                nctx = Context(ctx, IndexKey, lastIndex)
                req = Request(nctx, "GET", href)
                yield nctx, req
            except Exception as e:
                self.logger.error(traceback.format_exc())
                yield Error(str(e))
            finally:
                lastIndex += 1
        u = resp.url
        u.path = "/api/product/search/v2/categories/{}".format(cid)
        vals = u.query()
        for (k,v) in r.get("search", dict()).get("query", dict()).items():
            if k in ["cid", "page"]:
                continue
            vals.set(k, v)
        vals.set("offset", str(len(r["search"]["products"])))
        vals.set("limit", "72")
        u.raw_query = vals.encode()
        nctx = Context(ctx, IndexKey, lastIndex)
        req = Request(nctx, "get", str(u))
        yield nctx, req

    def parseProductsJSON(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        pass

    def parseProductGroup(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        pass

    productDetailDataReg = re.compile("window\.asos\.pdp\.config\.product\s*=\s*({[^;]+});", re.I|re.U)
    stockPriceReg = re.compile("window\.asos\.pdp\.config\.stockPriceApiUrl\s*=\s*'(/api/product/catalogue/[^;]+)'\s*;", re.I|re.U)
    appVersionReg = re.compile("window\.asos\.pdp\.config\.appVersion\s*=\s*'([a-z0-9-.]+)';", re.I|re.U)
    ratingReg = re.compile("window\.asos\.pdp\.config\.ratings\s*=\s*({.*?});", re.I|re.U)
    descReg = re.compile('<script\s+id="split\-structured\-data"\s+type="application/ld\+json">(.*?)</script>', re.I|re.U)

    def parseProduct(self, ctx:Context, resp:Response) -> Generator[Any,None,None]:
        respBody = resp.body.decode()

        i, sp, rating, desc, variants = None,None,None,None,dict()
        try:
            matched = self.productDetailDataReg.findall(respBody)
            if len(matched) < 1:
                raise Error("extract product detail from {} failed".format(resp.rawurl))
            i = json.loads(matched[0])
        except JSONDecodeError:
            raise Error("decode product detial failed")

        try:
            matchedRating = self.ratingReg.findall(respBody)
            if len(matchedRating) > 0:
                rating = json.loads(matchedRating[0])
            else:
                rating = dict()
        except JSONDecodeError: 
            raise Error("decode product detial failed")

        try:
            matchedDesc = self.descReg.findall(respBody)
            if len(matchedDesc) > 0:
                desc = json.loads(matchedDesc[0])
            else:
                desc = dict()
        except JSONDecodeError:
            raise Error("decode product detial failed")

        matchedStock = self.stockPriceReg.findall(respBody)
        matchedApiVer = self.appVersionReg.findall(respBody)
        if len(matchedStock) < 1 or len(matchedApiVer) < 1:
            self.logger.error("stock: %d, apiversion: %d", len(matchedStock), len(matchedApiVer))
            raise Error("extract product stock url or api version from {} failed".format(resp.rawurl))
        # get stock info by api
        stockUrl = "{}://{}{}".format(resp.url.scheme, resp.url.host, matchedStock[0])
        req = Request(ctx, "get", stockUrl)
        vals = req.url.query()
        vals.set("store", "US")
        vals.set("currency", "USD")
        req.url.raw_query = vals.encode()

        opts = self.CrawlOptions(req.url)
        for k in opts.headers.keys():
            req.headers.set(k, opts.headers.get(k))
        req.headers.set("accept-encoding", "gzip, deflate, br")
        req.headers.set("accept", "*/*")
        req.headers.set("referer", str(resp.url))
        req.headers.set("user-agent", resp.request.headers.get("user-agent"))
        req.headers.set("asos-c-name", "asos-web-productpage")
        req.headers.set("asos-c-version", matchedApiVer[0])

        cookie = req.headers.get("cookie")
        for c in opts.cookies:
            if c.path == "" or  req.url.path.startswith(c.path):
                val = "{}={}".format(c.name, c.value)
                if cookie == "":
                    cookie = val
                else:
                    cookie = cookie + "; " + val
        if cookie != "":
            req.headers.set("cookie", cookie)
        try:
            opts = RequestOptions(enable_proxy=True, enable_headless=opts.enableHeadless, relibility=opts.reliability)
            subresp = self._httpClient.do(ctx, req, opts)
            if subresp.status_code != 200:
                raise Error("access %s failed with status code {}",subresp.status_code)
            stocks = json.loads(subresp.body.decode())
            if not isinstance(stocks, list) or len(stocks) == 0:
                raise Error("got not valid stock price")
            sp = stocks[0]
            for variant in sp.get("variants", list()):
                if not (variant or dict()).get("variantId"):
                    self.logger.error("missing variantId field for {}".format(req.url))
                    continue
                variants[variant["variantId"]] = variant
        except Error as e:
            raise
        except Exception as e:
            raise Error("get stock info of {} failed, error={}", str(req.url), str(e))

        sel = resp.selector()
        
        canUrl = sel.css('link[rel="canonical"]').attrib.get("href")
        if not canUrl:
            canUrl = self.CanonicalUrl(str(resp.url))
        item = Product()
        source = item.source
        source.id = str(i["id"])
        source.crawlUrl = str(resp.rawurl)
        source.canonicalUrl = canUrl
        item.title = i["name"]
        item.description = desc.get("description", "")
        item.brandName = i["brandName"]
        item.crowdType = i["gender"]
        price = item.price
        price.currency = USD
        price.current = int(sp["productPrice"]["current"]["value"] * 100)
        stats = item.stats
        stats.rating = float(rating.get("averageOverallRating", 0))
        stats.reviewCount = int(rating.get("totalReviewCount", 0))
        stock = item.stock
        stock.stockStatus = Stock.InStock if i.get("isInStock") else Stock.OutStock

        if ctx.get_str("MainCategory") and ctx.get_str("Category"):
            item.crowdType = ctx.get_str("MainCategory")
            item.category = ctx.get_str("Category")
            item.subCategory = ctx.get_str("SubCategory")
            item.subCategory2 = ctx.get_str("SubCategory2")
            item.subCategory3 = ctx.get_str("SubCategory3")
        else:
            nodes = sel.css('nav[aria-label="breadcrumbs"]>ol>li>a::text')
            for index in range(len(nodes)):
                if index == len(nodes) - 1:
                    break
                node = nodes[index]
                if index == 1:
                    item.category = (node.get() or "").strip()
                elif index == 2:
                    item.subCategory = (node.get() or "").strip()
                elif index == 3:
                    item.subCategory2 = (node.get() or "").strip()
                elif index == 4:
                    item.subCategory3 = (node.get() or "").strip()
                elif index == 5:
                    item.subCategory4 = (node.get() or "").strip()

        for img in i["images"]:
            u = img["url"]
            media = newImageMedia("", u, u+"?wid=1000&fit=constrain", u+"?wid=650&fit=constrain", u+"?wid=650&fit=constrain", "", img["isPrimary"])
            item.medias.append(media)

        for rawsku in i["variants"]:
            variant = variants.get(rawsku["variantId"])
            if not variant:
                self.logger.error("no stock/price info found for sku {}".format(rawsku["variantId"]))
                continue
            sku = Sku()
            sku.sourceId = str(rawsku["variantId"])
            sku.medias.MergeFrom(item.medias)
            price = sku.price
            price.currency = USD
            price.current = int(variant["price"]["current"]["value"]*100)
            price.msrp = int(variant["price"]["previous"]["value"]*100)
            stock = sku.stock
            stock.stockStatus = Stock.InStock if variant["isInStock"] else Stock.OutOfStock

            # sku spec
            # color
            skuSpec = SkuSpecOption()
            skuSpec.type = SkuSpecColor
            skuSpec.id = str(rawsku["colourWayId"])
            skuSpec.name = rawsku["colour"]
            skuSpec.value = str(rawsku["colourWayId"])
            sku.specs.append(skuSpec)
            # size
            skuSpec = SkuSpecOption()
            skuSpec.type = SkuSpecSize
            skuSpec.id = str(rawsku["sizeId"])
            skuSpec.name = rawsku["size"]
            skuSpec.value = str(rawsku["sizeId"])
            sku.specs.append(skuSpec)

            item.skuItems.append(sku)
        
        self.logger.info("yield item")
        yield ctx, item

    def NewTestRequest(self, ctx:Context) -> Generator[Request,None,None]:
        yield super().NewTestRequest(ctx)

    def CheckTestResponse(self, ctx:Context, resp: Response) -> bool:
        return super().CheckTestResponse(ctx, resp)
