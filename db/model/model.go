package model

type PostTagLink struct {
	PostID int
	TagID  int
}

type PostTagLinkFetcher interface {
	Next(*PostTagLink) (bool, error)
}
