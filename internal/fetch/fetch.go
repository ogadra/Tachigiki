package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/markusmobius/go-trafilatura"
	"github.com/temoto/robotstxt"

	"github.com/ogadra/Tachigiki/internal/config"
)

const (
	KindArticle         = "article"
	KindPressConference = "press_conference"
)

const (
	userAgent    = "Tachigiki"
	hostInterval = time.Second
	maxArticles  = 3
)

type Article struct {
	SourceID    string    `json:"source_id"`
	URL         string    `json:"url"`
	Lang        string    `json:"lang"`
	Kind        string    `json:"kind"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

type Adapter func(ctx context.Context, src config.Source, day time.Time) ([]Article, error)

var adapters = map[string]Adapter{}

func Register(kind string, a Adapter) {
	if _, ok := adapters[kind]; ok {
		panic("fetch: adapter registered twice: " + kind)
	}
	adapters[kind] = a
}

func Run(ctx context.Context, day time.Time, sources []config.Source, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, src := range sources {
		a, ok := adapters[src.Kind]
		if !ok {
			return fmt.Errorf("source %s: unknown kind %q", src.ID, src.Kind)
		}
		articles, err := a(ctx, src, day)
		if err != nil {
			return fmt.Errorf("source %s: %w", src.ID, err)
		}
		if len(articles) > maxArticles {
			articles = articles[:maxArticles]
		}
		for i, art := range articles {
			b, err := json.MarshalIndent(art, "", "  ")
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s-%d.json", src.ID, i)), b, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

var ErrNotFound = errors.New("not found")

var (
	mu       sync.Mutex
	robots   = map[string]*robotstxt.RobotsData{}
	lastSeen = map[string]time.Time{}
)

func Get(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	r, ok := robots[u.Host]
	if !ok {
		status, body, err := request(ctx, u.Scheme+"://"+u.Host+"/robots.txt")
		if err != nil {
			return nil, err
		}
		if r, err = robotstxt.FromStatusAndBytes(status, body); err != nil {
			return nil, err
		}
		robots[u.Host] = r
	}
	if !r.TestAgent(u.RequestURI(), userAgent) {
		return nil, fmt.Errorf("%s: disallowed by robots.txt", rawURL)
	}
	status, body, err := request(ctx, rawURL)
	if err != nil {
		return nil, err
	}
	switch {
	case status == http.StatusNotFound:
		return nil, fmt.Errorf("%s: %w", rawURL, ErrNotFound)
	case status != http.StatusOK:
		return nil, fmt.Errorf("%s: status %d", rawURL, status)
	}
	return body, nil
}

func request(ctx context.Context, rawURL string) (int, []byte, error) {
	u, _ := url.Parse(rawURL)
	if wait := hostInterval - time.Since(lastSeen[u.Host]); wait > 0 {
		time.Sleep(wait)
	}
	defer func() { lastSeen[u.Host] = time.Now() }()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	return res.StatusCode, body, err
}

func ExtractBody(page []byte, rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	res, err := trafilatura.Extract(bytes.NewReader(page), trafilatura.Options{OriginalURL: u})
	if err != nil {
		return "", fmt.Errorf("%s: %w", rawURL, err)
	}
	return res.ContentText, nil
}
