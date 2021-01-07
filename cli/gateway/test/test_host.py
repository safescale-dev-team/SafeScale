import requests
import json

version = "v1"
baseurl = "http://localhost:8080"

def RequestException(Exception):
    def __init__(self, *args, **kwargs):
        super(RequestExcetion, self).__init__(*args, **kwargs)

def req(method, request, data):
    response = method(f"{baseurl}/{version}/{request}", data={"name": "TestOVH"})
#     print(response.__dict__, response.content)
    if response.status_code!=200:
        raise Exception(str(response))
    
    res = json.loads(response.content)
    print(res)
    return res

def post(request, data={}):
    return req(requests.post, request, data)

def get(request, data={}):
    return req(requests.post, request, data)
   
response = post("tenant", data={"name": "TestOVH"})

response = get("host/list")

response = post("host")
