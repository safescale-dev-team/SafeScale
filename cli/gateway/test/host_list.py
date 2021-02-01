from helper import post, delete, get

def print_hosts(r):
    for item in r["hosts"]:
        print(item['name'])
       
response = post("tenant", params={"name": "ovh-snapearth"})

response = get("hosts")
print_hosts(response)
