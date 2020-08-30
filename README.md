
<p align="center">
  <a href="https://github.com/IamCathal/Req">
    <img
      alt="Req"
      src="https://i.imgur.com/OBMzTA1.png"
      width="760"
    />
  </a>
</p>

![example workflow name](https://github.com/IamCathal/steamFriendsGraphing/workflows/Go/badge.svg) ![coverage badge](coverage_badge.png)

## What's the goal of this project? 
The goal of this project is to determine the degrees of seperation between any two users on [Steam](https://store.steampowered.com/)

## What is a degree of seperation?
*"Six degrees of separation is the idea that all people are six, or fewer, social connections away from each other. Also known as the 6 Handshakes rule. As a result, a chain of "a friend of a friend" statements can be made to connect any two people in a maximum of six steps."* - [Wikipedia](https://en.wikipedia.org/wiki/Six_degrees_of_separation)

## Installation
After cloning the repo you are going to need to get your [Steam Web API key](https://partner.steamgames.com/doc/webapi_overview/auth) and create a file called `APIKEYS.txt` and place it into the main directory.

Now you can run the script either by building and then running it like `./main [arguments] [steam64ID]` or by running it with the command `go run main.go util.go [arguments] [steam64ID]`. To see all available arguments you can use `./main --help`

For the moment the easiest way to find your steam64ID is to use [Steam ID Finder](https://steamidfinder.com/)

The unit tests will fail unless you set your the APIKey environment variable with `os.Setenv("APIKey", "your key")` before line 22 in `main_test.go`

## How do you get that coverage badge?

I'm using Jordan Pole's [gopherbadger](https://github.com/jpoles1/gopherbadger) in a pre-commit hook. The git hook is quite simple and looks like this:
```bash
#!/bin/bash
echo "Testing and generating coverage badge, this might take a few seconds"
badge -png=true
git add coverage_badge.png
```
