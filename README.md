
<p align="center">
  <a href="https://github.com/IamCathal/Req">
    <img
      alt="Req"
      src="https://i.imgur.com/OBMzTA1.png"
      width="760"
    />
  </a>
</p>

![example workflow name](https://github.com/IamCathal/steamFriendsGraphing/workflows/Go/badge.svg) ![go report card badge](https://goreportcard.com/badge/github.com/iamcathal/steamfriendsgraphing)

## What's the goal of this project? 
Determine the degrees of seperation between any two users on [Steam](https://store.steampowered.com/). This is done through crawling a users friend list and creating a graph representation of it.


## What is a degree of seperation?
*"Six degrees of separation is the idea that all people are six, or fewer, social connections away from each other. Also known as the 6 Handshakes rule. As a result, a chain of "a friend of a friend" statements can be made to connect any two people in a maximum of six steps."* - [Wikipedia](https://en.wikipedia.org/wiki/Six_degrees_of_separation)

## How does the program work?

The application can be split into two clear halves. One half does the crawling to gather the information needed and the other compiles this data into a graph format. A single user can be crawled to just map the friend network for one target user or two users can be chosen to attempt to find a degree of seperation between them.

### Crawling
Without utilizing some form of concurrency crawling would take forever. Since the application has a lot of downtime in waiting for API calls to return and it's not too CPU intensive overall using concurrency to process multiple users at once is key. The current implementation uses a [worker pool](https://gobyexample.com/worker-pools). 

A worker pool allows the application to place jobs into a pool where workers can then asynchronously pull them down and process them. This is the best way of going about this problem and the amount of workers can be set by the user to increase overall throughput.


<p align="center">
    <img
      alt="worker pool diagram with gophers"
      src="https://miro.medium.com/max/800/1*ugshDOhXfC287WWhG4IfSA.jpeg"
      width="550"
    />
  </a>
</p>

<p align="center">
 Heres a nice illustration of a worker pool courtesy of <a href="https://medium.com/@j.d.livni">Joseph Livni</a>
</p>


### Graphing
The graphing functionality can be split into two sections:
* Create the graph output seen by the user using [go-echarts](https://github.com/go-echarts/go-echarts)
* Find the degree of seperation between two users if possible using [dijkstra](https://github.com/IamCathal/dijkstra2)

## Installation
After cloning the repo you are going to need to get your [Steam Web API key](https://partner.steamgames.com/doc/webapi_overview/auth) and create a file called `APIKEYS.txt` and place it into the root directory.


Now you can run the script. The easiest way is to build and then run the executable like so:
``cd src && go build`` and then `` ./steamFriendsGraphing [flags] [steamID]``. Don't forget to use `--help` to see all options.

*Keep in mind that for now the executable can only be invoked from the `src` directory*

For the moment the easiest way to find your steam64ID is to use [Steam ID Finder](https://steamidfinder.com/)

## Testing

Tests are split into two groups; service and integration. Heres how to run each set of tests:

|             |                                                         |
| ----------- |:-------------------------------------------------------:| 
| Service     | `cd src && go test -v ./... --tags=service`             |
| Integration | `cd src && go test -v ./... --tags=integration`         |
| All         | `cd src && go test -v ./... --tags=service,integration -p 1` |

My personal githook runs all service tests before committing. Heres an example
```bash
#!/bin/bash
cd prjectDirectory
go test -v ./... -tags=service 
```

For personal testing I use this command. It runs the tests for all the packages and automatically opens a chrome tab with the coverage report.
```
go test -v -p 1 ./... -cover -coverprofile=coverage.out --tags=service,integration && \ 
  go tool cover -html=coverage.out -o coverage.html && \
    google-chrome coverage.html
```
