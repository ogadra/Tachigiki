// Package assemble は文ごとのWAVを英→日→英の順に連結し、レベルごとのm4aと再生位置付きのscript.jsonを作る。
package assemble

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ogadra/Tachigiki/internal/translate"
	"github.com/ogadra/Tachigiki/internal/tts"
)

const (
	sampleRate     = 24000
	bytesPerSecond = sampleRate * 2 // 16bit mono

	gapWithin  = 0.6 // 英→日→英の間
	gapBetween = 1.2 // 文と文の間
	gapTitle   = 1.5 // 記事タイトルの後
	chimeHz    = 880
	chimeSec   = 0.3
)

type Segment struct {
	Lang  string  `json:"lang"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

type Sentence struct {
	En       string    `json:"en"`
	Ja       string    `json:"ja"`
	Segments []Segment `json:"segments"`
}

type Article struct {
	Title     translate.Pair `json:"title"`
	SourceURL string         `json:"source_url"`
	Start     float64        `json:"start"`
	Sentences []Sentence     `json:"sentences"`
}

type Script struct {
	Date     string    `json:"date"`
	Level    string    `json:"level"`
	Articles []Article `json:"articles"`
}

// Run は scriptDir と audioDir を読み、outDir/{level}/ に audio.m4a と script.json を書き出す。
func Run(ctx context.Context, day, scriptDir, audioDir, outDir string) error {
	for _, lv := range translate.Levels {
		articles, err := translate.ReadLevel(scriptDir, lv)
		if err != nil {
			return err
		}
		segDir := filepath.Join(audioDir, lv)
		pcm, script, err := build(articles, func(a, s int, lang string) ([]byte, error) {
			return normalize(ctx, tts.SegmentPath(segDir, a, s, lang))
		})
		if err != nil {
			return err
		}
		script.Date, script.Level = day, lv
		dir := filepath.Join(outDir, lv)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		if err := encode(ctx, pcm, filepath.Join(dir, "audio.m4a")); err != nil {
			return err
		}
		b, err := json.MarshalIndent(script, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "script.json"), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// build は load で読んだ16bit monoのPCMを並べ、各文の再生位置を記録する。load の s はタイトルのとき -1。
func build(articles []translate.Article, load func(a, s int, lang string) ([]byte, error)) ([]byte, Script, error) {
	var pcm bytes.Buffer
	now := func() float64 { return float64(pcm.Len()) / bytesPerSecond }
	put := func(a, s int, lang string) (Segment, error) {
		b, err := load(a, s, lang)
		if err != nil {
			return Segment{}, err
		}
		seg := Segment{Lang: lang, Start: now()}
		pcm.Write(b)
		seg.End = now()
		return seg, nil
	}
	var script Script
	for ai, a := range articles {
		art := Article{Title: a.Title, SourceURL: a.SourceURL, Start: now()}
		pcm.Write(chime())
		for _, lang := range []string{"en", "ja"} {
			if _, err := put(ai, -1, lang); err != nil {
				return nil, Script{}, err
			}
			pcm.Write(silence(gapWithin))
		}
		pcm.Write(silence(gapTitle - gapWithin))
		for si, s := range a.Sentences {
			sen := Sentence{En: s.En, Ja: s.Ja}
			for i, lang := range []string{"en", "ja", "en"} {
				if i > 0 {
					pcm.Write(silence(gapWithin))
				}
				seg, err := put(ai, si, lang)
				if err != nil {
					return nil, Script{}, err
				}
				sen.Segments = append(sen.Segments, seg)
			}
			pcm.Write(silence(gapBetween))
			art.Sentences = append(art.Sentences, sen)
		}
		script.Articles = append(script.Articles, art)
	}
	return pcm.Bytes(), script, nil
}

func silence(sec float64) []byte {
	return make([]byte, int(sec*sampleRate)*2)
}

func chime() []byte {
	n := int(chimeSec * sampleRate)
	b := make([]byte, n*2)
	for i := range n {
		fade := 1 - float64(i)/float64(n)
		v := int16(0.3 * math.MaxInt16 * fade * math.Sin(2*math.Pi*chimeHz*float64(i)/sampleRate))
		b[2*i], b[2*i+1] = byte(v), byte(v>>8)
	}
	return b
}

// normalize はWAVの音量をそろえ、24kHz 16bit monoのPCMにして返す。
func normalize(ctx context.Context, wav string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-i", wav, "-af", "loudnorm=I=-16:TP=-1.5:LRA=11", "-ar", "24000", "-ac", "1", "-f", "s16le", "-")
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

func encode(ctx context.Context, pcm []byte, out string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-y", "-f", "s16le", "-ar", "24000", "-ac", "1", "-i", "-", "-c:a", "aac", "-b:a", "64k", out)
	cmd.Stdin = bytes.NewReader(pcm)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
