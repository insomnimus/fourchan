// Package fourchan is a small wrapper for the 4chan web api.
package fourchan

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// endpoints.
const (
	BoardsEndpoint = "https://a.4cdn.org/boards.json"
	BASE           = "https://a.4cdn.org"
)

// ThreadNo is the id of a thread.
type ThreadNo int

func (t ThreadNo) String() string {
	return fmt.Sprintf("%d", t)
}

// Page represents a page in a board.
type Page struct {
	// order of the page.
	PageNo  int             `json:"page"`
	Threads []ThreadPreview `json:"threads"`
}

// ThreadPreview is a preview for a thread.
type ThreadPreview struct {
	// Caption code. OP only, can be empty.
	CapCode string `json:"capcode,omitempty"`
	// 1= closed, 0 = open.
	Closed int `json:"closed,omitempty"`
	// OP's comment, if any.
	Comment      string `json:"com"`
	Ext          string `json:"ext"`
	FileName     string `json:"filename"`
	LastModified int    `json:"last_modified"`
	LastReplies  []Post `json:"last_replies,omitempty"`
	Name         string `json:"name"`
	// no is like an id, pass no to GetThread().
	No           ThreadNo `json:"no"`
	Now          string   `json:"now"`
	OmittedPosts int      `json:"omitted_posts,omitempty"`
	ReplyCount   int      `json:"replies"`
	// resto is 0 for the op.
	Resto       int    `json:"resto"`
	SemanticURL string `json:"semantic_url"`
	// sticky 1= pinned, 0= not pinned.
	Sticky  int    `json:"sticky,omitempty"`
	Subject string `json:"sub,omitempty"`
	Time    int    `json:"time"`
}

// FullThread represents an entire thread with all of its contents.
type FullThread struct {
	Posts []Post `json:"posts"`
}

// Post represents a single comment/post in a thread.
type Post struct {
	// posts id/no.
	ID int `json:"no"`
	// 1= pinned, 0= not pinned.
	Sticky int `json:"sticky,omitempty"`
	// 1= closed, 0= open.
	Closed int    `json:"closed,omitempty"`
	Now    string `json:"now"`
	// username of the poster, defaults to anonymous.
	Name string `json:"name"`
	// OP only, subject of the thread.
	Subject string `json:"sub,omitempty"`
	// the text content of the post.
	Comment string `json:"com"`
	// posts attachments file name, if any.
	FileName string `json:"filename"`
	// file extension.
	Ext  string `json:"ext"`
	Time int    `json:"time"`
	// 0 for op.
	Resto int `json:"resto"`
	// caption code, OP only. can be empty.
	CapCode     string `json:"capcode"`
	SemanticURL string `json:"semantic_url,omitempty"`
	ReplyCount  int    `json:"replies,omitempty"`
	// number of unique posters.
	UniquePosters int `json:"unique_ips,omitempty"`
}

// Board represents a 4chan board.
type Board struct {
	// boards full title, i.e "music" or "pokemon".
	Title string `json:"title"`
	// boards short name, i.e "mu" for music.
	Code string `json:"board"`
	// a short text describing the board.
	Desc string `json:"meta_description"`
	// safe for work= 1, NSFW= 0
	SFW int `json:"ws_board"`
}

// Catalog returns the catalog pages for the board.
func (b *Board) Catalog() ([]Page, error) {
	return GetCatalog(b.Code)
}

// GetThread returns a complete thread.
//
// board: short code of the board, i.e "mu" or "g".
// id: threads thread no/id.
func GetThread(board string, id ThreadNo) (*FullThread, error) {
	uri := fmt.Sprintf("%s/%s/thread/%s.json", BASE, board, id)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, wrap(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, wrap(err)
	}
	var thread FullThread
	err = json.Unmarshal(data, &thread)
	return &thread, wrap(err)
}

// GetCatalog returns the catalog pages for a fourchan board.
//
// board: the boards short code, for example "mu" (for the music board).
func GetCatalog(board string) ([]Page, error) {
	uri := fmt.Sprintf("%s/%s/catalog.json", BASE, board)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, wrap(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, wrap(err)
	}
	var catalog []Page
	err = json.Unmarshal(data, &catalog)
	return catalog, wrap(err)
}

// GetBoards returns all available boards.
func GetBoards() ([]Board, error) {
	resp, err := http.Get(BoardsEndpoint)
	if err != nil {
		return nil, wrap(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, wrap(err)
	}
	response := struct {
		Boards []Board `json:"boards"`
	}{}
	err = json.Unmarshal(data, &response)
	return response.Boards, wrap(err)
}

func wrap(err error) error {
	if err == nil {
		return err
	}
	return fmt.Errorf("fourchan: %w", err)
}
