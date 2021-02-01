from helper import post, delete, get

name = "test2"

response = post("cluster", params={"name": name, "flavor": 2, })
