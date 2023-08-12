package app

import "github.com/go-msvc/errors"

type fileItemPrompt struct {
	Caption string       `json:"caption"`
	Next    fileItemNext `json:"next"`
}

func (prompt fileItemPrompt) Validate() error {
	if prompt.Caption == "" {
		return errors.Errorf("missing caption")
	}
	if err := prompt.Next.Validate(); err != nil {
		return errors.Wrapf(err, "invalid next")
	}
	return nil
}

//		// 	html += fmt.Sprintf("<form method=\"POST\">%s<input name=\"input\"/><button type=\"submit\">Enter</button></form>", p.Caption)

// if p := currentItem.Prompt; p != nil {
// 	log.Debugf("form: %+v", httpReq.Form)
// 	input := httpReq.Form["input"]
// 	log.Debugf("input:\"%s\"", input)

// 	//save input
// 	session.Values[p.Name] = input
// 	//todo validate

// 	//navigate to next after valid input
// 	nextItem, ok := app.app[p.Next]
// 	if ok {
// 		log.Debugf("Navigate to prompt.next=%s", p.Next)
// 		currentItemId = p.Next
// 		currentItem = nextItem
// 	} else {
// 		log.Errorf("unknown next:\"%s\"", p.Next)
// 	}
// } //case POST
