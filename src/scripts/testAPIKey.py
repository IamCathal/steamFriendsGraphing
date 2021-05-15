import requests
import sys

BASEURL = "http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key="

def main(APIKey):
    res = requests.get(f"{BASEURL}{APIKey}&steamid=76561198078629620&relationship=friend")
    try:
        resObj = res.json()
    except:
        if (res.status_code == 403):
            print("Invalid")
        else:
            print(f"Unexpected failure. Error code is {res.status_code}")
    else:
        print("Valid")

if __name__ == "__main__":
    main(sys.argv[1])