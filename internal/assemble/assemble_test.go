package assemble

import (
	"testing"

	"github.com/ogadra/Tachigiki/internal/translate"
)

func TestBuildPlacesSentencesEnJaEn(t *testing.T) {
	oneSec := make([]byte, bytesPerSecond)
	articles := []translate.Article{{Sentences: []translate.Pair{{En: "a", Ja: "あ"}, {En: "b", Ja: "い"}}}}
	pcm, script, err := build(articles, func(int, int, string) ([]byte, error) { return oneSec, nil })
	if err != nil {
		t.Fatal(err)
	}
	// 効果音 + タイトル英 + 間 + タイトル日 + 間 + タイトル後の残り
	first := chimeSec + 1 + gapWithin + 1 + gapTitle
	want := []Segment{
		{"en", first, first + 1},
		{"ja", first + 1 + gapWithin, first + 2 + gapWithin},
		{"en", first + 2 + 2*gapWithin, first + 3 + 2*gapWithin},
	}
	got := script.Articles[0].Sentences[0].Segments
	for i := range want {
		if d := got[i].Start - want[i].Start; d > 1e-9 || d < -1e-9 || got[i].Lang != want[i].Lang {
			t.Fatalf("segment %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
	second := script.Articles[0].Sentences[1].Segments[0].Start
	if d := second - (want[2].End + gapBetween); d > 1e-9 || d < -1e-9 {
		t.Fatalf("second sentence starts at %v", second)
	}
	if end := float64(len(pcm)) / bytesPerSecond; end < second+3 {
		t.Fatalf("pcm too short: %v", end)
	}
}
