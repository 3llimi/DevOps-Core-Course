#!/usr/bin/env python3
"""
Dynamic inventory script for local Vagrant VM.
Discovers host details dynamically at runtime.
"""
import json
import socket
import subprocess

def get_vagrant_info():
    hostname = socket.gethostname()
    ip = socket.gethostbyname(hostname)
    return hostname, ip

def main():
    hostname, ip = get_vagrant_info()
    
    inventory = {
        "webservers": {
            "hosts": ["localhost"],
            "vars": {
                "ansible_connection": "local",
                "ansible_user": "vagrant",
                "ansible_python_interpreter": "/usr/bin/python3",
                "discovered_hostname": hostname,
                "discovered_ip": ip
            }
        },
        "_meta": {
            "hostvars": {
                "localhost": {
                    "ansible_connection": "local",
                    "ansible_user": "vagrant",
                    "ansible_python_interpreter": "/usr/bin/python3",
                    "discovered_hostname": hostname,
                    "discovered_ip": ip
                }
            }
        }
    }
    print(json.dumps(inventory, indent=2))

if __name__ == "__main__":
    main()
