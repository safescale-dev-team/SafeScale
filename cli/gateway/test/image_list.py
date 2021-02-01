from helper import post, delete, get

def print_images(r):
    for item in r["images"]:
        print(f"{item['name']}: {item['id']}")

response = get("images")
print_images(response)
