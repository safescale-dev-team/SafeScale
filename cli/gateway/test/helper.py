import requests
import json

version = "v1"
baseurl = "http://localhost:8080"

def RequestException(Exception):
    def __init__(self, *args, **kwargs):
        super(RequestExcetion, self).__init__(*args, **kwargs)

def req(method, request, params={}, body=None):
    if body is None:
        response = method(f"{baseurl}/{version}/{request}", data=params)
    else:
        response = method(f"{baseurl}/{version}/{request}", data=params, json=body)
        
#     print(response.__dict__, response.content)
    if response.status_code!=200:
        
        raise Exception(json.loads(response.content))
    
    print("----", response.content)
    res = json.loads(response.content)
    return res

def post(request, params={}, body=None):
    return req(requests.post, request, params, body)

def get(request, params={}, body=None):
    return req(requests.get, request, params, body)

def delete(request, params={}, body=None):
    return req(requests.delete, request, params, body)
