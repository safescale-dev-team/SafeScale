from helper import post, delete, get
from data import network_name, host_name

print(f"delete network {host_name}")
response = delete(f"host/{host_name}")
