# -*- coding: UTF-8 -*-

class Header(object):
    """
    Http request header
    """

    def __init__(self):
        self._values = dict()

    def _unifyKey_(self, key):
        return (key or "").lower()

    def get(self, key:str)->str:
        """ get header """
        key = self._unifyKey_(key)
        for v in (self._values.get(key) or list()):
            return v
        return ""

    def add(self, key:str, val:str):
        """ add header """
        key = self._unifyKey_(key)
        self._values[key] = self._values.get(key, list())
        if val not in self._values[key]:
            self._values[key].append(val)

    def set(self, key:str, val:str):
        """ set header """
        key = self._unifyKey_(key)
        self._values[key] = [val]

    def delete(self, key:str):
        """ delete header """
        key = self._unifyKey_(key)
        if self._values.get(key):
            del self._values[key]
        
    @property
    def values(self)->list:
        return self._values.items()
