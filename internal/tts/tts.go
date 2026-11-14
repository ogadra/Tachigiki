// Package tts は英文をKokoro、和訳をVOICEVOXで合成し、文ごとのWAVを書き出す。
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

// SegmentPath は記事 a の文 s (タイトルは s = -1) の lang 音声の置き場所を返す。
func SegmentPath(dir string, a, s int, lang string) string {
	if s < 0 {
		return filepath.Join(dir, fmt.Sprintf("%d-title-%s.wav", a, lang))
	}
	return filepath.Join(dir, fmt.Sprintf("%d-%d-%s.wav", a, s, lang))
}

// Run は scriptDir の {level}.json をすべて読み、outDir/{level}/ に文ごとのWAVを書き出す。
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
			pairs := append([]translate.Pair{a.Title}, a.Sentences...)
			for i, p := range pairs {
				en, err := kokoro(ctx, p.En)
				if err != nil {
					return err
				}
				ja, err := voicevox(ctx, p.Ja)
				if err != nil {
					return err
				}
				if err := os.WriteFile(SegmentPath(dir, ai, i-1, "en"), en, 0o644); err != nil {
					return err
				}
				if err := os.WriteFile(SegmentPath(dir, ai, i-1, "ja"), ja, 0o644); err != nil {
					return err
				}
			}
		}
	}
	return nil
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
