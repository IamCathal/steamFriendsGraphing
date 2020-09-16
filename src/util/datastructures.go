package util

// IndivFriend holds profile information for
// a single user
type IndivFriend struct {
	Steamid string `json:"steamid"`
	// Steamid is the default steamid64 value for a user
	Relationship string `json:"relationship"`
	// Relationship is always friend in this case
	FriendSince int64 `json:"friend_since,omitempty"`
	// FriendSince is the unix timestamp of when the friend request was accepted
	Username string `json:"username"`
	// Username is steam public username
}

// FriendsStruct is messy but it holds the array of all friends for a given user
type FriendsStruct struct {
	Username    string `json:"username"`
	FriendsList struct {
		Friends []IndivFriend `json:"friends"`
	} `json:"friendslist"`
}

// Player holds all account information for a given user. This is the
// response from the getPlayerSummaries endpoint
type Player struct {
	Steamid                  string `json:"steamid"`
	Communityvisibilitystate int    `json:"communityvisibilitystate"`
	Profilestate             int    `json:"profilestate"`
	Personaname              string `json:"personaname"`
	Profileurl               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	Avatarmedium             string `json:"avatarmedium"`
	Avatarfull               string `json:"avatarfull"`
	Personastate             int    `json:"personastate"`
	Realname                 string `json:"realname,omitempty"`
	Primaryclanid            string `json:"primaryclanid,omitempty"`
	Timecreated              int    `json:"timecreated,omitempty"`
	Personastateflags        int    `json:"personastateflags,omitempty"`
	Loccountrycode           string `json:"loccountrycode,omitempty"`
	Commentpermission        int    `json:"commentpermission,omitempty"`
}

// UserStatsStruct is the JSON response of the
// userstats lookup to get usernames from steamIDs
type UserStatsStruct struct {
	Response struct {
		Players []struct {
			Player
		} `json:"players"`
	} `json:"response"`
}
