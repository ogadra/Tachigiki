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

// kanteiPress は首相官邸の内閣官房長官記者会見の会見録を、当日の午前 (a) と午後 (p) の分だけ取得する。
func kanteiPress(ctx context.Context, src config.Source, day time.Time) ([]fetch.Article, error) {
	var out []fetch.Article
	for _, half := range []string{"a", "p"} {
		u := fmt.Sprintf("https://www.kantei.go.jp/jp/tyoukanpress/%s/%d_%s.html", day.Format("200601"), day.Day(), half)
		page, err := fetch.Get(ctx, u)
		// 会見のない日や時間帯はページ自体がない
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
			Lang:        "ja",
			Kind:        "press_conference",
			Title:       "内閣官房長官記者会見",
			Body:        body,
			PublishedAt: day,
		})
	}
	return out, nil
}
