package main

import (
	"fmt"
	"github.com/go-chi/chi"
	fc "github.com/insomnimus/fourchan"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, getHome())
}

func handleBoard(w http.ResponseWriter, r *http.Request) {
	board := chi.URLParam(r, "board")
	catalog, err := fc.GetCatalog(board)
	if err != nil {
		fmt.Fprintf(w, "error fetching the board %s: %s", board, err)
		return
	}
	msg := `<html lang="en">
	<body>
	<h1>` + board + `</h1><br>`
	for _, page := range catalog {
		for _, t := range page.Threads {
			temp := `<h2> <a href="/` + board
			temp += fmt.Sprintf("/%s", t.No) + `">`
			if t.Subject == "" {
				temp += "thread"
			}
			temp += fmt.Sprintf("%s</a></h2>\n", html.EscapeString(t.Subject))
			// heading done, add content
			if t.CapCode != "" {
				temp += fmt.Sprintf("<div>%s</div><br>\n", t.CapCode)
			}
			temp += fmt.Sprintf("%s | %s\n", t.Name, t.Comment)
			temp += fmt.Sprintf("omitted %d posts<br>", t.OmittedPosts)
			// add the last replies if any
			if len(t.LastReplies) > 0 {
				temp += "<ul>"
				for _, reply := range t.LastReplies {
					if reply.Comment == "" {
						temp += fmt.Sprintf("<li> %s | no content </li>\n", reply.Name)
						continue
					}
					temp += fmt.Sprintf("<li> %s | %s </li>\n", reply.Name, reply.Comment)
				}
				temp += "</ul>"
			}
			msg += temp
		}
	}
	// close the html
	msg += `</body></div>`
	fmt.Fprint(w, msg)
}

var homeHTML string

func getHome() string {
	if homeHTML != "" {
		return homeHTML
	}
	boards, err := fc.GetBoards()
	if err != nil {
		return fmt.Sprintf("error fetching the list of boards: %s", err)
	}
	msg := `
	<html>
	<body>
	<h1> Boards </h1>
	<br>
	<ul>
	`
	for _, b := range boards {
		msg += `<li> <a href="/` + b.Code + `">`
		msg += fmt.Sprintf("%s </a> </li>\n", b.Title)
	}
	msg += `</ul>
	</body>
	</html>`
	homeHTML = msg
	return msg
}

func handleThread(w http.ResponseWriter, r *http.Request) {
	board := chi.URLParam(r, "board")
	thr := chi.URLParam(r, "thread")
	threadNo, err := strconv.Atoi(thr)
	if err != nil {
		fmt.Fprintf(w, "error converting %s to int: %s", thr, err)
		return
	}
	thread, err := fc.GetThread(board, fc.ThreadNo(threadNo))
	if err != nil {
		fmt.Fprintf(w, "error getting the thread %s: %s", threadNo, err)
		return
	}
	msg := `<html lang="en"><body>`
	if len(thread.Posts) == 0 {
		fmt.Fprintf(w, "no posts, sorry")
		return
	}
	if thread.Posts[0].Subject != "" {
		msg += fmt.Sprintf("<h1> %s </h1>", thread.Posts[0].Subject)
	}
	msg += "<ul>\n"
	// populate the thread
	for _, p := range thread.Posts {
		if p.Comment == "" {
			continue
		}
		msg += fmt.Sprintf("<li> %s </li>\n", resolve(thread.Posts, p))
	}
	// done, close the html
	msg += `</ul></body></html>`
	fmt.Fprint(w, msg)
}

var resolver = regexp.MustCompile(`<a\shref\="#[a-zA-Z0-9]{10}"\sclass\="quotelink">&gt;&gt;([0-9]{9})</a>`)

func resolve(posts []fc.Post, p fc.Post) string {
	if !resolver.MatchString(p.Comment) {
		return p.Comment
	}
	nums := resolver.FindAllStringSubmatch(p.Comment, -1)
	if len(nums) == 0 {
		return p.Comment
	}
	comment := p.Comment
	for _, n := range nums {
		if len(n) < 2 {
			continue
		}
		no, _ := strconv.Atoi(n[1])
		for _, post := range posts {
			if post.ID == no {
				msg := post.Comment
				if resolver.MatchString(msg) {
					msg = resolver.ReplaceAllString(msg, "")
				}
				comment = strings.Replace(comment, n[0], "<blockquote>"+msg+"</blockquote>", -1)
			}
		}
	}
	return comment
}

func main() {
	r := chi.NewRouter()
	r.Get("/", handleHome)
	r.Get("/{board}", handleBoard)
	r.Get("/{board}/{thread}", handleThread)
	fmt.Println("running at port 44444")
	http.ListenAndServe(":44444", r)
}
