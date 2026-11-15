package adapters

import (
	"context"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/ogadra/Tachigiki/internal/config"
	"github.com/ogadra/Tachigiki/internal/fetch"
)

func init() { fetch.Register("rss", rss) }

func rss(ctx context.Context, src config.Source, day time.Time) ([]fetch.Article, error) {
	b, err := fetch.Get(ctx, src.URL)
	if err != nil {
		return nil, err
	}
	feed, err := gofeed.NewParser().ParseString(string(b))
	if err != nil {
		return nil, err
	}
	var out []fetch.Article
	for _, item := range feed.Items {
		if item.PublishedParsed == nil || item.PublishedParsed.In(day.Location()).Format(time.DateOnly) != day.Format(time.DateOnly) {
			continue
		}
		page, err := fetch.Get(ctx, item.Link)
		if err != nil {
			return nil, err
		}
		body, err := fetch.ExtractBody(page, item.Link)
		if err != nil {
			return nil, err
		}
		out = append(out, fetch.Article{
			SourceID:    src.ID,
			URL:         item.Link,
			Lang:        src.Lang,
			Kind:        fetch.KindArticle,
			Title:       item.Title,
			Body:        body,
			PublishedAt: *item.PublishedParsed,
		})
	}
	return out, nil
}
