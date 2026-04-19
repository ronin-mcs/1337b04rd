package models

import "time"

type Post struct {
	PostID        int
	Title         string
	TextContent   string
	AnonID        int
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

type Anon struct {
	AnonID int
	PostID int
	Avatar string // store the name of character from Rick and Morty api
	// it can be searched as "https://rickandmortyapi.com/api/character/?name=rick"
	AnonName string
}

type Comment struct {
	CommentID   int
	PostID      int
	AddressedTo int // comment_id or orifinal post
	TextContent string
	AnonID      int
	CreatedAt   time.Time
}

type Session struct {
	SessionId int         `json:"session_id"`
	Sessions  map[int]int `json:"sessions"`
	// list of post_id mapping to its anon_id in this post
}

type Attachment struct {
	AttachmentID int
	PostID       int
	CommentID    *int
	// если CommentID == nil, attachment относится к посту
	// если CommentID != nil, attachment относится к комменту
	FileKey      string
	OriginalName string
	ContentType  string
}
