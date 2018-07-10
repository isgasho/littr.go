package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mariusor/littr.go/models"

	"github.com/gorilla/mux"
)

type comment struct {
	Item
	Parent   *comment
	Path     []byte
	FullPath []byte
	Children []*comment
}

type contentModel struct {
	Title         string
	InvertedTheme bool
	Content       comment
}

func sluggify(s string) string {
	if s == "" {
		return s
	}
	return strings.Replace(s, "/", "-", -1)
}

func ReparentComments(allComments []*comment) {
	for _, cur := range allComments {
		par := func(t []*comment, path []byte) *comment {
			// findParent
			if path == nil {
				return nil
			}
			for _, n := range t {
				if bytes.Equal(path, n.FullPath) {
					return n
				}
			}
			return nil
		}(allComments, cur.Path)

		if par != nil {
			cur.Parent = par
			par.Children = append(par.Children, cur)
		}
	}
}

// handleMain serves /{year}/{month}/{day}/{hash} request
func (l *Littr) HandleContent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date, err := time.Parse(time.RFC3339, fmt.Sprintf("%s-%s-%sT00:00:00+00:00", vars["year"], vars["month"], vars["day"]))
	if err != nil {
		l.HandleError(w, r, StatusUnknown, err)
		return
	}
	hash := vars["hash"]
	items := make([]Item, 0)

	db := l.Db

	sel := `select "content_items"."id", "content_items"."key", "mime_type", "data", "title", "content_items"."score",
			"submitted_at", "submitted_by", "handle", "path", "content_items"."flags" from "content_items"
			left join "accounts" on "accounts"."id" = "content_items"."submitted_by"
			where "submitted_at" > $1::date and "content_items"."key" ~* $2`
	rows, err := db.Query(sel, date, hash)
	if err != nil {
		l.HandleError(w, r, StatusUnknown, err)
		return
	}
	m := contentModel{InvertedTheme: l.InvertedTheme}
	p := models.Content{}
	var i Item
	for rows.Next() {
		var handle string
		err = rows.Scan(&p.Id, &p.Key, &p.MimeType, &p.Data, &p.Title, &p.Score, &p.SubmittedAt, &p.SubmittedBy, &handle, &p.Path, &p.Flags)
		if err != nil {
			l.HandleError(w, r, StatusUnknown, err)
			return
		}
		m.Title = string(p.Title)
		i = LoadItem(p, handle)
		m.Content = comment{Item: i, Path: p.Path, FullPath: p.FullPath()}
	}
	if p.Data == nil {
		l.HandleError(w, r, http.StatusNotFound, fmt.Errorf("not found"))
		return
	}
	items = append(items, i)

	if r.Method == http.MethodGet {
		q := r.URL.Query()
		yay := len(q["yay"]) > 0
		nay := len(q["nay"]) > 0
		multiplier := 0

		if yay || nay {
			if nay {
				multiplier = -1
			}
			if yay {
				multiplier = 1
			}
			_, err := l.Vote(p, multiplier, CurrentAccount.id)
			if err != nil {
				log.Print(err)
			}
			http.Redirect(w, r, p.PermaLink(), http.StatusFound)
		}
	}

	if r.Method == http.MethodPost {
		e, err := l.ContentFromRequest(r, p.FullPath())
		if err != nil {
			l.HandleError(w, r, http.StatusInternalServerError, err)
			return
		}
		l.Vote(*e, 1, CurrentAccount.id)
		http.Redirect(w, r, p.PermaLink(), http.StatusFound)
	}

	allComments := make([]*comment, 0)
	allComments = append(allComments, &m.Content)
	// comments
	selCom := `select "content_items"."id", "content_items"."key", "mime_type", "data", "title", "content_items"."score", 
			"submitted_at", "submitted_by", "handle", "path", "content_items"."flags" from "content_items" 
			left join "accounts" on "accounts"."id" = "content_items"."submitted_by" 
			where "path" <@ $1 order by "path" asc, "score" desc`
	{
		rows, err := db.Query(selCom, m.Content.Path)

		if err != nil {
			l.HandleError(w, r, StatusUnknown, err)
			return
		}
		for rows.Next() {
			c := models.Content{}
			var handle string
			err = rows.Scan(&c.Id, &c.Key, &c.MimeType, &c.Data, &c.Title, &c.Score, &c.SubmittedAt, &c.SubmittedBy, &handle, &c.Path, &c.Flags)
			if err != nil {
				l.HandleError(w, r, StatusUnknown, err)
				return
			}

			i := LoadItem(c, handle)
			com := comment{Item: i, Path: c.Path, FullPath: c.FullPath()}
			items = append(items, i)
			allComments = append(allComments, &com)
		}
	}

	ReparentComments(allComments)
	_, err = LoadVotes(CurrentAccount, items)
	if err != nil {
		log.Print(err)
	}
	err = l.SessionStore.Save(r, w, l.GetSession(r))
	if err != nil {
		log.Print(err)
	}

	RenderTemplate(w, "content.html", m)
}
