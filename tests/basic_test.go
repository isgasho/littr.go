package tests

import (
	"fmt"
	as "github.com/go-ap/activitystreams"
	"testing"
)

type getTest map[string]collectionVal
type postTest map[string]postVal

var defaultCollectionTestPairs = getTest{
	"actors": {
		id:  fmt.Sprintf("%s/actors", apiURL),
		typ: string(as.CollectionType),
		first: &collectionVal{
			id: fmt.Sprintf("%s/actors?page=1", apiURL),
			// TODO(marius): fix actors collection pages
			//typ: string(as.CollectionPageType),
		},
		itemCount: 2,
		items: map[string]objectVal{
			"actors/eacff9dd": {
				id:                fmt.Sprintf("%s/actors/eacff9dd", apiURL),
				typ:               string(as.PersonType),
				name:              "anonymous",
				preferredUsername: "anonymous",
				url:               fmt.Sprintf("http://%s/~anonymous", host),
				inbox: &collectionVal{
					id: fmt.Sprintf("%s/actors/eacff9dd/inbox", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionPageType),
				},
				outbox: &collectionVal{
					id: fmt.Sprintf("%s/actors/eacff9dd/outbox", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionPageType),
				},
				liked: &collectionVal{
					id: fmt.Sprintf("%s/actors/eacff9dd/liked", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionPageType),
				},
				score: 0,
			},
			"actors/dc6f5f5b": {
				id:                fmt.Sprintf("%s/actors/dc6f5f5b", apiURL),
				typ:               string(as.PersonType),
				name:              "system",
				preferredUsername: "system",
				url:               fmt.Sprintf("http://%s/~system", host),
				inbox: &collectionVal{
					id: fmt.Sprintf("%s/actors/dc6f5f5b/inbox", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionType),
				},
				outbox: &collectionVal{
					id: fmt.Sprintf("%s/actors/dc6f5f5b/outbox", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionType),
				},
				liked: &collectionVal{
					id: fmt.Sprintf("%s/actors/dc6f5f5b/liked", apiURL),
					// TODO(marius): Fix different page id when dereferenced vs. in parent collection
					//typ: string(as.OrderedCollectionType),
				},
				score: 0,
			},
		},
	},
	"self/inbox": {
		id:  fmt.Sprintf("%s/self/inbox", apiURL),
		typ: string(as.OrderedCollectionType),
		first: &collectionVal{
			id: fmt.Sprintf("%s/self/inbox?maxItems=50&page=1", apiURL),
		},
		// TODO(marius): We need to fix the criteria for populating the inbox to
		//     verifying if the actor that submitted the activity is local or not
		itemCount: 1, // TODO(marius): :FIX_INBOX: this should be 0
	},
	"self/liked": {
		id:  fmt.Sprintf("%s/self/liked", apiURL),
		typ: string(as.OrderedCollectionType),
		first: &collectionVal{
			id: fmt.Sprintf("%s/self/liked?maxItems=50&page=1", apiURL),
		},
		itemCount: 0,
	},
	"self/outbox": {
		id:  fmt.Sprintf("%s/self/outbox", apiURL),
		typ: string(as.OrderedCollectionType),
		first: &collectionVal{
			id: fmt.Sprintf("%s/self/outbox?maxItems=50&page=1", apiURL),
		},
		itemCount: 1,
		items: map[string]objectVal{
			"actors/dc6f5f5b/outbox/162edb32": {
				id:  fmt.Sprintf("%s/actors/dc6f5f5b/outbox/162edb32", apiURL),
				typ: string(as.CreateType),
				act: &objectVal{
					id: fmt.Sprintf("%s/actors/dc6f5f5b", apiURL),
				},
				obj: &objectVal{
					id:        fmt.Sprintf("%s/actors/dc6f5f5b/outbox/162edb32/object", apiURL),
					typ:       string(as.NoteType),
					name:      "about littr.me",
					url:       "/~system/162edb32",
					content:   "<p>This is a new attempt at the social news aggregator paradigm.<br/>It's based on the ActivityPub web specification and as such tries to leverage federation to prevent some of the pitfalls found in similar existing communities.</p>",
					mediaType: "text/html",
					author:    fmt.Sprintf("%s/actors/dc6f5f5b", apiURL),
					score:     0,
				},
			},
		},
	},
}

var c2sTestPairs = postTest{
	"Like": {
		body: fmt.Sprintf(`{
    "type": "Like",
    "actor": "%s/actors/dc6f5f5b",
    "object": "%s/actors/dc6f5f5b/outbox/162edb32/object"
}`, apiURL, apiURL),
		res: objectVal{
			id: fmt.Sprintf("%s/actors/dc6f5f5b/liked/162edb32", apiURL),
			typ: string(as.LikeType),
			obj: &objectVal{author: fmt.Sprintf("%s/actors/dc6f5f5b", apiURL),
				id: fmt.Sprintf("%s/actors/dc6f5f5b/outbox/162edb32/object", apiURL),
			},
		},
	},
	"Create": {
		body: fmt.Sprintf(`{
  "type": "Create",
  "actor": "%s/actors/dc6f5f5b",
  "to": ["%s/self/outbox"],
  "object": {
    "type": "Note",
    "inReplyTo": "%s/actors/dc6f5f5b/outbox/162edb32",
    "content": "<p>Hello world!</p>"
  }
}`, apiURL, apiURL, apiURL),
		res: objectVal{
			typ: string(as.CreateType),
			obj: &objectVal{
				author: fmt.Sprintf("%s/actors/dc6f5f5b", apiURL),
				typ: string(as.NoteType),
				content: "<p>Hello world!</p>",
			},
		},
	},
}

var s2sTestPairs = postTest{}

func Test_GET(t *testing.T) {
	assertCollection := errOnCollection(t)
	for k, col := range defaultCollectionTestPairs {
		t.Run(k, func(t *testing.T) {
			assertCollection(fmt.Sprintf("%s/%s", apiURL, k), col)
		})
	}

}

func Test_POST_Outbox(t *testing.T) {
	assertPost := errOnPostRequest(t)
	for typ, test := range c2sTestPairs {
		t.Run("Activity_"+typ, func(t *testing.T) {
			assertPost(test)
		})
	}
}