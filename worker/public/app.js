const $ = (id) => document.getElementById(id);
const audio = $("audio");
let script = { articles: [] };
let spans = []; // [{start, end, el}] 再生位置の昇順

async function load() {
  const res = await fetch(`/data/${$("day").value}/${$("level").value}/script.json`);
  if (!res.ok) {
    $("articles").textContent = "この日のこのレベルの音声はありません。";
    return;
  }
  script = await res.json();
  audio.src = `/data/${script.date}/${script.level}/audio.m4a`;
  audio.playbackRate = Number($("rate").value);
  render();
  captions();
}

// ページに出している英文と和訳を、audio の字幕としても渡す
function captions() {
  const time = (t) => new Date(t * 1000).toISOString().slice(11, 23);
  const cues = spans.map((s) => `${time(s.start)} --> ${time(s.end)}\n${s.el.textContent}`);
  const vtt = `WEBVTT\n\n${cues.join("\n\n")}\n`;
  $("captions").src = URL.createObjectURL(new Blob([vtt], { type: "text/vtt" }));
}

function render() {
  spans = [];
  $("articles").replaceChildren(
    ...script.articles.map((a) => {
      const section = document.createElement("section");
      const h2 = document.createElement("h2");
      h2.textContent = `${a.title.en} / ${a.title.ja}`;
      section.append(h2);
      for (const s of a.sentences) {
        const button = document.createElement("button");
        button.className = "sentence";
        const en = document.createElement("span");
        en.textContent = s.en;
        const ja = document.createElement("span");
        ja.className = "ja";
        ja.textContent = s.ja;
        button.append(en, ja);
        button.addEventListener("click", () => seek(s.segments[0].start));
        for (const seg of s.segments) {
          spans.push({ start: seg.start, end: seg.end, el: seg.lang === "en" ? en : ja });
        }
        section.append(button);
      }
      return section;
    }),
  );
}

function seek(t) {
  audio.currentTime = t;
  audio.play();
}

let active = null;
audio.addEventListener("timeupdate", () => {
  const t = audio.currentTime;
  const cur = spans.find((s) => s.start <= t && t < s.end)?.el ?? null;
  if (cur === active) return;
  active?.classList.remove("active");
  cur?.classList.add("active");
  cur?.scrollIntoView({ block: "center", behavior: "smooth" });
  active = cur;
});

function skipArticle(dir) {
  const starts = script.articles.map((a) => a.start);
  const t = audio.currentTime;
  // 前へは記事の頭から2秒以内なら1つ前の記事へ、それ以降なら今の記事の頭へ戻す
  const target = dir > 0 ? starts.find((s) => s > t) : (starts.findLast((s) => s < t - 2) ?? 0);
  if (target !== undefined) seek(target);
}

navigator.mediaSession?.setActionHandler("play", () => audio.play());
navigator.mediaSession?.setActionHandler("pause", () => audio.pause());
navigator.mediaSession?.setActionHandler("previoustrack", () => skipArticle(-1));
navigator.mediaSession?.setActionHandler("nexttrack", () => skipArticle(1));
audio.addEventListener("play", () => {
  navigator.mediaSession.metadata = new MediaMetadata({ title: `Tachigiki ${script.date}`, artist: script.level });
});

$("rate").addEventListener("change", () => {
  audio.playbackRate = Number($("rate").value);
});
$("day").addEventListener("change", load);
$("level").addEventListener("change", load);

const days = await (await fetch("/data/days")).json();
$("day").replaceChildren(...days.map((d) => new Option(d, d)));
if (days.length > 0) load();
