import matplotlib.pyplot as plt
import networkx as nx
import sys


G = nx.Graph()
knownFriends = []

filePath = "userData/{}.txt".format(sys.argv[1])
f = open(filePath, "r", encoding="utf-8")
contents = f.read().splitlines()

G.add_node(sys.argv[1])

# add first level friends and original person
for userData in contents:
    steamID = userData.split("\t")[0]
    username = userData.split("\t")[1]
    knownFriends.append(steamID)
    # G.add_edge("76561198051845236", username)
    G.add_edge(sys.argv[1], username)

for userData in contents:
    try:
        f2 = open("userData/{}.txt".format(userData.split("\t")[1]), "r", encoding="utf-8")
        contents2 = f2.read().splitlines()

        for secondaryFriend in contents2:
            secondarySteamID = secondaryFriend.split("\t")[0]
            secondaryUsername = secondaryFriend.split("\t")[1]
            if secondarySteamID in knownFriends:
                G.add_edge(userData.split("\t")[1], secondaryUsername)
                knownFriends.append(secondarySteamID)
                print(userData.split("\t")[1], secondaryUsername)
    except:
        print("file not found")


# G.add_node(elem[:-4])
G.add_node(sys.argv[1])

# for neighbour in contents:
#     G.add_edge("76561198078629620",neighbour[-3:])

print(nx.info(G))

print(knownFriends)

nx.draw(G, node_size=100, with_labels=True, font_size=6)
plt.show()
