package util

// FriendsStruct is exactly whats saved on file for any given user
type FriendsStruct struct {
	Username    string      `json:"username"`
	FriendsList Friendslist `json:"friendslist"`
}

// Friend holds details of a friend for a given user
type Friend struct {
	Username     string `json:"username"`
	Steamid      string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendSince  int    `json:"friend_since"`
}

// FriensdList holds all friends for a given user
type Friendslist struct {
	Friends []Friend `json:"friends"`
}

// UserStatsStruct is the response from the steam web API
// for /getPlayerSummary calls
type UserStatsStruct struct {
	Response Response `json:"response"`
}

// Player holds all details returned by the steam web API for
// the /getPlayerSummary endpoint
type Player struct {
	Steamid                  string `json:"steamid"`
	Communityvisibilitystate int    `json:"communityvisibilitystate"`
	Profilestate             int    `json:"profilestate"`
	Personaname              string `json:"personaname"`
	Commentpermission        int    `json:"commentpermission"`
	Profileurl               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	Avatarmedium             string `json:"avatarmedium"`
	Avatarfull               string `json:"avatarfull"`
	Avatarhash               string `json:"avatarhash"`
	Personastate             int    `json:"personastate"`
	Realname                 string `json:"realname"`
	Primaryclanid            string `json:"primaryclanid"`
	Timecreated              int    `json:"timecreated"`
	Personastateflags        int    `json:"personastateflags"`
	Loccountrycode           string `json:"loccountrycode"`
}

// Response is filler
type Response struct {
	Players []Player `json:"players"`
}
