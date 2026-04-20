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
	AnonID   int
	PostID   int
	Avatar   string // strore characterID as string, so we can retrieve from api and get avatar image
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
		ExpiresAt time.Time   `json:"expires_at"`
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
