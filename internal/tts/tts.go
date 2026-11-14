package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/ogadra/Tachigiki/internal/translate"
)

const (
	kokoroURL     = "http://localhost:8880"
	kokoroVoice   = "af_heart"
	voicevoxURL   = "http://localhost:50021"
	voicevoxVoice = "3"
)

func TitlePath(dir string, article int, lang string) string {
	return filepath.Join(dir, fmt.Sprintf("%d-title-%s.wav", article, lang))
}

func SentencePath(dir string, article, sentence int, lang string) string {
	return filepath.Join(dir, fmt.Sprintf("%d-%d-%s.wav", article, sentence, lang))
}

func Run(ctx context.Context, scriptDir, outDir string) error {
	for _, lv := range translate.Levels {
		articles, err := translate.ReadLevel(scriptDir, lv)
		if err != nil {
			return err
		}
		dir := filepath.Join(outDir, lv)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for ai, a := range articles {
			if err := synthesize(ctx, a.Title, TitlePath(dir, ai, "en"), TitlePath(dir, ai, "ja")); err != nil {
				return err
			}
			for si, s := range a.Sentences {
				if err := synthesize(ctx, s, SentencePath(dir, ai, si, "en"), SentencePath(dir, ai, si, "ja")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func synthesize(ctx context.Context, p translate.Pair, enPath, jaPath string) error {
	en, err := kokoro(ctx, p.En)
	if err != nil {
		return err
	}
	ja, err := voicevox(ctx, p.Ja)
	if err != nil {
		return err
	}
	if err := os.WriteFile(enPath, en, 0o644); err != nil {
		return err
	}
	return os.WriteFile(jaPath, ja, 0o644)
}

func kokoro(ctx context.Context, text string) ([]byte, error) {
	body, err := json.Marshal(map[string]string{"model": "kokoro", "input": text, "voice": kokoroVoice, "response_format": "wav"})
	if err != nil {
		return nil, err
	}
	return post(ctx, kokoroURL+"/v1/audio/speech", body)
}

func voicevox(ctx context.Context, text string) ([]byte, error) {
	q := url.Values{"speaker": {voicevoxVoice}, "text": {text}}
	query, err := post(ctx, voicevoxURL+"/audio_query?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return post(ctx, voicevoxURL+"/synthesis?speaker="+voicevoxVoice, query)
}

func post(ctx context.Context, u string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: status %d: %s", u, res.StatusCode, b)
	}
	return b, nil
}
