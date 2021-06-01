# -*- coding: UTF-8 -*-

from context import Context

TracingIdKey = "__internal_tracing_id__"
JobIdKey = "__internal_job_id__"
ReqIdKey = "__internal_request_id__"
StoreIdKey = "__internal_store_id__"
TargetTypeKey = "__internal_target_type__"
IndexKey = "__internal_index__"

def isTargetTypeSupported(ctx:Context, msg)->bool:
    # TODO: convert msg to type string
    typs = (ctx.get(TargetTypeKey) or "").split(",")
    return msg in typs