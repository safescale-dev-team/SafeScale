import requests
import json

version = "v1"
baseurl = "http://localhost:8080"

def RequestException(Exception):
    def __init__(self, *args, **kwargs):
        super(RequestExcetion, self).__init__(*args, **kwargs)

def req(method, request, params):
    response = method(f"{baseurl}/{version}/{request}", params=params)
#     print(response.__dict__, response.content)
    if response.status_code!=200:
        
        raise Exception(json.loads(response.content))
    
    res = json.loads(response.content)
    print(res)
    return res

def post(request, params={}):
    return req(requests.post, request, params)

def get(request, params={}):
    return req(requests.get, request, params)

def delete(request, params={}):
    return req(requests.delete, request, params)
