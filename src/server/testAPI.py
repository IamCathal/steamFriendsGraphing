import requests
import json
 
bodyObj = {
     'level': '2',
     'statMode':'true',
     'testKeys':'false',
     'workers':'2',
     'steamID':'76561197960271945'
    }

req = requests.post("http://localhost:8080/crawl", json = bodyObj)
print(req.json())
 

