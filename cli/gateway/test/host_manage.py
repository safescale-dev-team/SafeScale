from helper import post, delete, get
from data import network_name, host_name

def print_hosts(r):
    for item in r["hosts"]:
        print(item['name'])

print("host list")
response = get("hosts")
print_hosts(response)

print(f"inspect {host_name}")
response = get(f"host/{host_name}")

print(f"ssh identifiers for {host_name}")
response = get(f"host/{host_name}/ssh")

print(f"reboot {host_name}")
response = post(f"host/{host_name}/reboot")

print(f"stop {host_name}")
response = post(f"host/{host_name}/stop")

print(f"start {host_name}")
response = post(f"host/{host_name}/start")

print(f"security groups for {host_name}")
response = get(f"host/{host_name}/security_groups")
