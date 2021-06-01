# -*- coding: UTF-8 -*-

import uuid
import secrets
from hashlib import md5

def newRandomId()->str:
    code = uuid.uuid4().hex + secrets.token_hex(16)
    return md5(bytes(code, "ASCII")).hexdigest()
