package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

// [{"navigate":{"url":"http://url.com"}},
// {"click":{"element":".element"}}]

func NewAction(actionType string, params json.RawMessage) (Action, error) {
	var action Action
	var err error
	switch actionType {
	case "click":
		action = &ClickAction{}
		err = json.Unmarshal(params, &action)
	case "paginate":
		action = &PaginateAction{}
		err = json.Unmarshal(params, &action)
	default:
		err = fmt.Errorf("Failed to create new action. Unknown or undefined action type")
	}
	return action, err
}

type Action interface {
	Execute(ctx context.Context, f *ChromeFetcher) error
}

type ClickAction struct {
	Element string `json:"element"`
}

func (a *ClickAction) Execute(ctx context.Context, f *ChromeFetcher) error {
	path := filepath.Join(viper.GetString("CHROME_SCRIPTS"), "scroll2bottom.js")
	return f.RunJSFromFile(ctx, path, fmt.Sprintf(`clickElement("%s");`, a.Element))
}

type PaginateAction struct {
	MaxPage int    `json:"maxpage"`
	Element string `json:"element"`
}

func (pa *PaginateAction) Execute(ctx context.Context, f *ChromeFetcher) error {
	path := filepath.Join(viper.GetString("CHROME_SCRIPTS"), "scroll2bottom.js")
	return f.RunJSFromFile(ctx, path, fmt.Sprintf(`ScrollDown(%d, "%s");`, pa.MaxPage, pa.Element))
}
