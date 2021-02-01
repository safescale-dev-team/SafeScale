from helper import post, delete, get
from data import network_name, host_name

print(f"create host {host_name}")
response = post("host", body={"name": host_name, "network": network_name, "sizing_as_string": "cpu ~ 1, ram ~ 2", "image_id": "Ubuntu 18.04"})
