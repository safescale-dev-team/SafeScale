from helper import post, delete, get

print(f"delete network {network_name}")
response = delete(f"network/{network_name}")
