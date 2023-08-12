package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-msvc/logger"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/michaeljs1990/sqlitestore"
)

var cookieCutter securecookie.Codec

const cookieName = "MyCookieName"

var store *sqlitestore.SqliteStore

func init() {
	var err error
	store, err = sqlitestore.NewSqliteStore("./database", "sessions", "/", 3600, []byte("<SecretKey>"))
	if err != nil {
		panic(err)
	}
	//todo: for prod, replace this e.g. with dynamo db like
	//see https://github.com/gorilla/sessions for list of options
}

func main() {
	app, err := Load("./app.json")
	if err != nil {
		panic(fmt.Sprintf("%+s", err))
	}

	// Hash keys should be at least 32 bytes long
	hashKey := []byte(os.Getenv("HASH_KEY"))
	if len(hashKey) == 0 {
		hashKey = securecookie.GenerateRandomKey(32)
		log.Errorf("Using random HASH_KEY")
	}
	if len(hashKey) != 32 {
		panic("HASH_KEY not 32 long")
	}

	// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
	// Shorter keys may weaken the encryption used.
	var blockKey = []byte(os.Getenv("BLOCK_KEY"))
	if len(blockKey) == 0 {
		blockKey = securecookie.GenerateRandomKey(32)
		log.Errorf("Using random BLOCK_KEY")
	}
	if len(blockKey) != 32 {
		panic("BLOCK_KEY not 32 bytes long")
	}
	cookieCutter = securecookie.New(hashKey, blockKey)

	http.HandleFunc("/", hdlr(app))
	fmt.Println("Starting the server on :3000...")
	http.ListenAndServe(":3000", nil)
}

var log = logger.New().WithLevel(logger.LevelDebug)

type ClientData struct {
	DeviceID string
}

func hdlr(app App) func(httpRes http.ResponseWriter, httpReq *http.Request) {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)

		//look at client cookie to see if returning device or a new device
		clientData := ClientData{}
		if cookie, err := httpReq.Cookie(cookieName); err == nil {
			if err = cookieCutter.Decode(cookieName, cookie.Value, &clientData); err == nil {
				log.Debugf("Received Cookie Value: %+v", clientData)
			} else {
				log.Errorf("failed to decode cookie: %+v", err)
			}
		} else {
			log.Errorf("failed to get cookie: %+v", err)
		}

		//load existing session or create a new session
		if clientData.DeviceID == "" {
			//new session
			clientData.DeviceID = uuid.New().String()
			log.Debugf("New Session: %s", clientData.DeviceID)
		}

		session, err := store.Get(httpReq, clientData.DeviceID)
		if err != nil {
			log.Errorf("failed to get session data: %+v", err)
		} else {
			log.Debugf("Loaded Session: %#v\n", session)
		}

		//current id from session data
		currentItemId, ok := session.Values["current_item"].(string)
		if !ok || currentItemId == "" {
			currentItemId = "home"
		}
		currentItem, ok := app[currentItemId]
		if !ok {
			log.Errorf("unknown current_item:\"%s\", reset to \"home\"", currentItemId)
			currentItemId = "home"
			currentItem, ok = app[currentItemId]
			if !ok {
				http.Error(httpRes, "cannot go home", http.StatusInternalServerError)
				return
			}
		}

		switch httpReq.Method {
		case http.MethodPost:
			if p := currentItem.Prompt; p != nil {
				httpReq.ParseForm()
				log.Debugf("form: %+v", httpReq.Form)
				input := httpReq.Form["input"]
				log.Debugf("input:\"%s\"", input)

				//save input
				session.Values[p.Name] = input
				//todo validate

				//navigate to next after valid input
				nextItem, ok := app[p.Next]
				if ok {
					log.Debugf("Navigate to prompt.next=%s", p.Next)
					currentItemId = p.Next
					currentItem = nextItem
				} else {
					log.Errorf("unknown next:\"%s\"", p.Next)
				}
			} //case POST
		case http.MethodGet:
			//navigate from menu is GET with ?next=<next item>
			if nextItemId := httpReq.URL.Query().Get("next"); nextItemId != "" {
				nextItem, ok := app[nextItemId]
				if ok {
					log.Debugf("Navigate to next=%s", nextItemId)
					currentItemId = nextItemId
					currentItem = nextItem
				} else {
					log.Errorf("unknown next:\"%s\"", nextItemId)
				}
			} //case GET
		} //switch method

		//ctx := context.Background()

		//render
		html := ""

		//nav links
		html += fmt.Sprintf("<p><a href=\"/?next=home\">Home</a></p>")
		//html += fmt.Sprintf("<p><a href=\"/?next=back\">Back</a></p>")

		//crumbs: todo... include only when can jump back, or show incremental back crumbs

		log.Debugf("Rendering current_item:\"%s\"", currentItemId)
		if m := currentItem.Menu; m != nil {
			html += fmt.Sprintf("<h1>%s</h1>\n", m.Title)
			for _, item := range m.Items {
				html += fmt.Sprintf("<p><a href=\"/?next=%s\">%s</a></p>", item.Next, item.Caption)
			}
		} else if p := currentItem.Prompt; p != nil {
			html += fmt.Sprintf("<form method=\"POST\">%s<input name=\"input\"/><button type=\"submit\">Enter</button></form>", p.Caption)
		} else if f := currentItem.Final; f != nil {
			html += fmt.Sprintf("<p>%s</p>", f.Caption)
		} else {
			http.Error(httpRes, "unknown item type", http.StatusInternalServerError)
			return
		}

		//update and save session data
		session.Values["current_item"] = currentItemId
		err = session.Save(httpReq, httpRes)
		if err != nil {
			log.Errorf("failed to save session: %+v", err)
		}

		//encode updated cookie value into the response
		if encoded, err := cookieCutter.Encode(cookieName, clientData); err == nil {
			cookie := &http.Cookie{
				Name:     cookieName,
				Value:    encoded,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			}
			http.SetCookie(httpRes, cookie)
			log.Debugf("defined cookie")
		} else {
			log.Errorf("failed to encode cookie")
		}

		httpRes.Header().Set("Content-Type", "text/html")
		httpRes.Write([]byte(html))
	}
}
