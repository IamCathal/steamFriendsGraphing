import requests
import json
import sys
 
bodyObj = {
     'level': '2',
     'statMode':'true',
     'testKeys':'false',
     'workers':'2',
     'steamID':'76561197960271945'
    }

if len(sys.argv) < 2:
    print("Invalid arguments\nUsage: python3 testAPI.py [mode]")
else:
    req = requests.post(f"http://localhost:8080/{sys.argv[1]}", json = bodyObj)
    print(req.json())

