from helper import post, delete, get
from data import network_name

print(f"create network {network_name}")
response = post("network", params={"name": network_name})
