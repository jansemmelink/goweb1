package web

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"os"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jansemmelink/goweb1/app"
	"github.com/michaeljs1990/sqlitestore"
)

var log = logger.New().WithLevel(logger.LevelDebug)

type App interface {
	Run() error
}

func New(app app.App) App {
	// Hash keys should be at least 32 bytes long
	hashKey := []byte(os.Getenv("HASH_KEY"))
	if len(hashKey) == 0 {
		hashKey = securecookie.GenerateRandomKey(32)
		log.Errorf("Using random HASH_KEY")
	}

	// Block keys should be 16 bytes (AES-128) or 32 bytes (AES-256) long.
	// Shorter keys may weaken the encryption used.
	var blockKey = []byte(os.Getenv("BLOCK_KEY"))
	if len(blockKey) == 0 {
		blockKey = securecookie.GenerateRandomKey(32)
		log.Errorf("Using random BLOCK_KEY")
	}

	return webApp{
		app:          app,
		cookieName:   "MyCookieName",
		hashKey:      hashKey,
		blockKey:     blockKey,
		cookieCutter: securecookie.New(hashKey, blockKey),
		sessionStore: nil,
	}
} //New()

type webApp struct {
	app          app.App
	cookieName   string
	hashKey      []byte
	blockKey     []byte
	cookieCutter securecookie.Codec
	sessionStore *sqlitestore.SqliteStore
}

func (w webApp) Run() error {
	//validation
	if len(w.hashKey) != 32 {
		return errors.Errorf("HASH_KEY not 32 long")
	}
	if len(w.blockKey) != 32 {
		return errors.Errorf("BLOCK_KEY not 32 bytes long")
	}

	//todo: option to replace default local session store in prod
	//for prod e.g., replace this e.g. with dynamo db like
	//see https://github.com/gorilla/sessions for list of options
	var err error
	w.sessionStore, err = sqlitestore.NewSqliteStore("./database", "sessions", "/", 3600, []byte("<SecretKey>"))
	if err != nil {
		return errors.Wrapf(err, "failed to create session store")
	}

	//register types stored in session data, else session save will fail
	gob.Register(app.PageData{})
	gob.Register(map[string]interface{}{})

	//setup and start HTTP server
	http.HandleFunc("/", w.hdlr())
	fmt.Println("Starting the server on :3000...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		return errors.Wrapf(err, "http server failed")
	}
	return nil
} //webapp.Run()

type ClientData struct {
	DeviceID string
}

type CtxClientData struct{}

func (w webApp) hdlr() func(httpRes http.ResponseWriter, httpReq *http.Request) {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)

		if httpReq.URL.Path != "/" {
			http.Error(httpRes, fmt.Sprintf("path \"%s\" not found", httpReq.URL.Path), http.StatusNotFound)
			return
		}

		//todo: context logger
		//todo: log current item id in each line of the logger

		ctx := w.userContext(httpReq)
		session := ctx.Value(app.CtxSession{}).(*sessions.Session)

		//load the currect app item to display/process
		currentItemId, ok := session.Values["current_item"].(string)
		if !ok || currentItemId == "" {
			currentItemId = "home"
		}

		//todo: also check app version and redirect in load-balancer to correct app version
		//todo: also check time when item was entered and discard if older than X
		//to protect against users from long ago suddenly waking after that version
		//of app was retired/changed
		currentItem, ok := w.app.GetItem(currentItemId)
		if !ok {
			//unknown item - likely an internal error or old app phased out
			log.Errorf("unknown current_item:\"%s\"", currentItemId)

			//do not just jump home - rather tell user and redirect home
			//so it does not appear like continuity break if there was really a fault
			redirect(httpRes, "Sorry - Session Terminated."+
				"Click to start a new session",
				"Start", "/")
			return
		}

		switch httpReq.Method {
		case http.MethodPost:
			log.Debugf("processing...")
			nextItemId, err := currentItem.Process(ctx, httpReq)
			if err != nil {
				log.Errorf("processing failed: %+v", err)
				redirect(httpRes, "failed to process input", "home", "/") //todo: retries etc...
				return
			}
			log.Debugf("processing done, next=\"%s\"", nextItemId)
			if nextItemId == "" {
				log.Errorf("processing succeeded but did not return nextItemId: %+v", err)
				redirect(httpRes, "failed to process input", "home", "/") //todo: retries etc...
				return
			}
			if currentItemId, currentItem, err = w.navigateTo(ctx, nextItemId); err != nil {
				log.Errorf("failed to nav to %s: %+v", nextItemId, err)
				redirect(httpRes, "failed to navigate", "home", "/")
				return
			}

			//navigate to next
			log.Errorf("NOT YET NAV AFTER POST!!!")

		case http.MethodGet:
			//navigate from menu if GET with ?next=<next item uuid>
			if nextItemUUID := httpReq.URL.Query().Get("next"); nextItemUUID != "" {
				//special case:
				if nextItemUUID == "home" {
					//reset and start over
					var err error
					currentItemId, currentItem, err = w.navigateTo(ctx, "home")
					if err != nil {
						panic(fmt.Sprintf("failed to nav home: %+v", err))
					}
				} else {
					//can only apply if page stored any links
					pageSessionData, ok := session.Values["page_data"].(app.PageData)
					if ok {
						nextSteps, ok := pageSessionData.Links[nextItemUUID]
						if ok {
							logSession(ctx, "before execute next steps")
							nextItemId, err := nextSteps.Execute(ctx)
							if err != nil {
								log.Errorf("failed to execute next steps: %+v", err)
							} else if nextItemId != "" {
								logSession(ctx, "after execute next steps")
								log.Debugf("next:\"%s\"", nextItemId)
								currentItemId, currentItem, err = w.navigateTo(ctx, nextItemId)
								if err != nil {
									log.Errorf("failed to nav to %s: %+v", nextItemId, err)
									redirect(httpRes, "failed to process input", "home", "/") //todo: retries etc...
									return
								}
							} else {
								log.Debugf("next steps dit not defined next item - stay here")
							}
						} else {
							log.Debugf("pageLink %s not found", nextItemUUID)
						}
					} else {
						log.Debugf("pageLinks not defined, ignoring %s", nextItemUUID)
					}
				}
			} else { //if has next=... in URL
				log.Debugf("next=... not defined in URL")
			}
		default:
			log.Errorf("Invalid method")
			http.Error(httpRes, "method not allowed", http.StatusMethodNotAllowed)
			return
		} //switch method

		log.Debugf("Rendering item(%s) ...", currentItemId)
		//render into buffer so that rendering can complete and define page data
		//before we write the cookie and session and then the page content
		//(wrong order does not save correctly)
		pageBuffer := bytes.NewBuffer(nil)
		pageSessionData, err := currentItem.Render(ctx, pageBuffer)
		if err != nil {
			log.Errorf("Rendering failed: %+v", err)
			redirect(httpRes, "Failed to render. Sorry!", "Restart", "/")
			return
		}

		//store optional page session data
		//it may be nil, but will be accessible to app.AppItem.Process()
		//from CtxPageData{}
		if pageSessionData != nil {
			session.Values["page_data"] = pageSessionData
			log.Debugf("UPDATED PAGE DATA ========================")
			log.Debugf("PAGE: (%T)%+v", pageSessionData, pageSessionData)
		}
		//update and save session data
		session.Values["current_item"] = currentItemId
		if err := session.Save(httpReq, httpRes); err != nil {
			log.Errorf("failed to save session: %+v", err)
		} else {
			log.Debugf("Saved Session(%d values, id:%s, name:%s):", len(session.Values), session.ID, session.Name())
			for n, v := range session.Values {
				log.Debugf("  Session[%s] = (%T)%+v", n, v, v)
			}
		}

		//encode updated cookie value into the response
		//(written to httpRes before content)
		clientData := ctx.Value(CtxClientData{}).(ClientData)
		if encoded, err := w.cookieCutter.Encode(w.cookieName, clientData); err == nil {
			cookie := &http.Cookie{
				Name:     w.cookieName,
				Value:    encoded,
				Path:     "/",
				Secure:   true,
				HttpOnly: true,
			}
			http.SetCookie(httpRes, cookie)
			log.Debugf("defined cookie(%s): (%T)%+v", w.cookieName, clientData, clientData)
		} else {
			log.Errorf("failed to encode cookie")
		}

		//write the page to the HTTP server responses
		httpRes.Header().Set("Content-Type", "text/html")
		httpRes.Write(pageBuffer.Bytes())
	} //func()
} //webapp.hdlr()

func (w webApp) userContext(httpReq *http.Request) context.Context {
	//look at client cookie to see if returning device or a new device
	clientData := ClientData{}
	if cookie, err := httpReq.Cookie(w.cookieName); err == nil {
		if err = w.cookieCutter.Decode(w.cookieName, cookie.Value, &clientData); err == nil {
			//log.Debugf("Decoded cookie(%s): (%T)%+v", w.cookieName, clientData, clientData)
		} else {
			log.Errorf("Failed to decode cookie: %+v", err)
		}
	} else {
		log.Errorf("Failed to get cookie: %+v", err)
	}

	//load existing session or create a new session
	if clientData.DeviceID == "" {
		//new session
		clientData.DeviceID = uuid.New().String()
		log.Debugf("New Session: %s", clientData.DeviceID)
	}

	session, err := w.sessionStore.Get(httpReq, clientData.DeviceID)
	if err != nil {
		log.Errorf("failed to get session data: %+v", err)
	} else {
		log.Debugf("Loaded Session(%d values, id:%s, name:%s):", len(session.Values), session.ID, session.Name())
		for n, v := range session.Values {
			log.Debugf("  Session[%s] = (%T)%+v", n, v, v)
		}
	}

	lang, ok := session.Values["lang"].(string)
	if !ok || len(lang) != 2 {
		lang = ""
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxClientData{}, clientData)
	ctx = context.WithValue(ctx, app.CtxSession{}, session)
	ctx = context.WithValue(ctx, app.CtxLang{}, lang)
	return ctx
} //webapp.userContext()

func redirect(httpRes http.ResponseWriter, message, button, link string) {
	//todo: template page
	httpRes.Write([]byte(fmt.Sprintf(
		"<form action=\"%s\" method=\"GET\">"+
			"<p>%s</p>"+
			"<button type=\"submit\">%s</button>"+
			"</form>",
		link, message, button)),
	)
}

func logSession(ctx context.Context, title string) {
	log.Debugf("SESSION %s", title)
	session := ctx.Value(app.CtxSession{}).(*sessions.Session)
	for n, v := range session.Values {
		log.Debugf("  Session[%s] = (%T)%+v", n, v, v)
	}
}

func (w webApp) navigateTo(ctx context.Context, nextItemId string) (string, app.AppItem, error) {
	nextItem, ok := w.app.GetItem(nextItemId)
	if !ok || nextItem == nil {
		return "", nil, errors.Errorf("unknown next:\"%s\"", nextItemId)
	}
	log.Debugf("Nav Item -> %s", nextItemId)
	if err := nextItem.OnEnterActions().Execute(ctx); err != nil {
		return "", nil, errors.Wrapf(err, "failed to execute on_enter_actions")
	}
	return nextItemId, nextItem, nil
}
