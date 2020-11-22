
# This is a MODIFIED version of the original repo
Some functionality is missing and this repo only serves to act as a dijkstra implementation 
for [IamCathal/SteamFriendsGraphing](https://github.com/IamCathal/steamfriendsgraphing)

I'm including it in the source code due to some annoying errors with package imports when using a forked version

## Documentation
[godoc](https://godoc.org/github.com/RyanCarrier/dijkstra)

#### Creating a graph

```go
package main

func main(){
  graph:=dijkstra.NewGraph()
  //Add the 3 verticies
  graph.AddVertex(0)
  graph.AddVertex(1)
  graph.AddVertex(2)
  //Add the Arcs
  graph.AddArc(0,1,1)
  graph.AddArc(0,2,1)
  graph.AddArc(1,0,1)
  graph.AddArc(1,2,2)
}

```

### Finding paths

Once the graph is created, shortest or longest paths between two points can be generated.
```go

best, err := graph.Shortest(0,2)
if err!=nil{
  log.Fatal(err)
}
fmt.Println("Shortest distance ", best.Distance, " following path ", best.Path)

best, err := graph.Longest(0,2)
if err!=nil{
  log.Fatal(err)
}
fmt.Println("Longest distance ", best.Distance, " following path ", best.Path)

```
