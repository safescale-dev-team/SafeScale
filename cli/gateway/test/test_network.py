from helper import post, delete, get

name = "test"

response = post("network", params={"name": name})

response = get(f"network/{name}")

response = delete(f"network/{name}")
