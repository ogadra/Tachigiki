package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ogadra/Tachigiki/internal/config"
	"github.com/ogadra/Tachigiki/internal/fetch"
)

func init() { fetch.Register("kantei_press", kanteiPress) }

func kanteiPress(ctx context.Context, src config.Source, day time.Time) ([]fetch.Article, error) {
	var out []fetch.Article
	for _, session := range []string{"a", "p"} {
		u := fmt.Sprintf("%s%s/%d_%s.html", src.URL, day.Format("200601"), day.Day(), session)
		page, err := fetch.Get(ctx, u)
		if errors.Is(err, fetch.ErrNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		body, err := fetch.ExtractBody(page, u)
		if err != nil {
			return nil, err
		}
		out = append(out, fetch.Article{
			SourceID:    src.ID,
			URL:         u,
			Lang:        src.Lang,
			Kind:        fetch.KindPressConference,
			Title:       "内閣官房長官記者会見",
			Body:        body,
			PublishedAt: day,
		})
	}
	return out, nil
}
