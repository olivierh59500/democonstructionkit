// Command record-demos exports the local demo collection and creates a portable
// video gallery. Outputs stay outside the individual source repositories.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type production struct{ Repo, File string }

var productions = []production{
	{"3d_doc", "3d-doc"}, {"bilizir-demo", "bilizir"}, {"dma-3d", "dma-3d"}, {"dma-is-back", "dma-is-back"},
	{"go-cocoisthebest", "coco-is-the-best"}, {"go-cuddlymenu", "cuddly-demo"}, {"go-dom-intro", "dom-intro"},
	{"go-fr010", "fr-010"}, {"go-multiscreen", "multiscreen"}, {"go-secondreality", "second-reality"},
	{"go-vectorballs", "vectorballs"}, {"grodan-kvack-kvack-demo", "grodan-kvack-kvack"}, {"megatwist", "megatwist"},
	{"nonameno-demo", "nonameno"}, {"phenomena-dna-scroll-intro", "phenomena-dna"},
	{"tcb-multi-plane-3d-scroller", "tcb-multi-plane-3d-scroller"}, {"tcb-replicants-demo", "tcb-replicants"},
	{"teamg1-demo", "teamg1"}, {"viva_tcb", "viva-tcb"},
}

type entry struct {
	Title    string  `json:"title"`
	File     string  `json:"file"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	FPS      int     `json:"fps"`
	Duration float64 `json:"duration_seconds"`
	Poster   string  `json:"poster"`
	Repo     string  `json:"repository"`
	Size     int64   `json:"bytes"`
}

func (e entry) Time() string { s := int(e.Duration); return fmt.Sprintf("%d:%02d", s/60, s%60) }

func main() {
	root := flag.String("root", "../..", "workspace containing demos and lib")
	out := flag.String("output", "", "output folder; defaults to workspace/videos/portfolio")
	only := flag.String("only", "", "comma-separated repository names; empty selects all")
	jobs := flag.Int("jobs", 1, "number of concurrent exports")
	flag.Parse()
	workspace, err := filepath.Abs(*root)
	if err != nil {
		fail(err)
	}
	if *out == "" {
		*out = filepath.Join(workspace, "videos", "portfolio")
	}
	output, err := filepath.Abs(*out)
	if err != nil {
		fail(err)
	}
	if *jobs < 1 || *jobs > 4 {
		fail(fmt.Errorf("jobs must be between 1 and 4"))
	}
	if err = os.MkdirAll(filepath.Join(output, "logs"), 0755); err != nil {
		fail(err)
	}
	selected := map[string]bool{}
	if *only != "" {
		for _, name := range strings.Split(*only, ",") {
			selected[name] = true
		}
	}
	var pending []production
	for _, p := range productions {
		if *only == "" || selected[p.Repo] {
			pending = append(pending, p)
			delete(selected, p.Repo)
		}
	}
	if len(selected) != 0 {
		fail(fmt.Errorf("unknown repositories: %v", selected))
	}
	queue := make(chan production, len(pending))
	for _, p := range pending {
		queue <- p
	}
	close(queue)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []error
	for range *jobs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range queue {
				if err := record(workspace, output, p); err != nil {
					mu.Lock()
					failures = append(failures, err)
					mu.Unlock()
					fmt.Fprintln(os.Stderr, err)
				}
			}
		}()
	}
	wg.Wait()
	if err := gallery(output); err != nil {
		fail(err)
	}
	if len(failures) > 0 {
		fail(fmt.Errorf("%d recording(s) failed; see %s", len(failures), filepath.Join(output, "logs")))
	}
	fmt.Println("Gallery:", filepath.Join(output, "index.html"))
}

func record(root, output string, p production) error {
	name := filepath.Join(output, p.File+".mp4")
	if _, err := os.Stat(name); err == nil {
		if err := probe(name); err != nil {
			return fmt.Errorf("%s: existing video is invalid: %w", p.Repo, err)
		}
		if _, err := os.Stat(filepath.Join(output, p.File+".json")); err != nil {
			return fmt.Errorf("%s: missing export report: %w", p.Repo, err)
		}
		fmt.Println(p.Repo + ": existing video verified")
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	log, err := os.Create(filepath.Join(output, "logs", p.File+".log"))
	if err != nil {
		return err
	}
	defer log.Close()
	fmt.Println(p.Repo + ": exporting")
	cmd := exec.Command("go", "run", "./dck/cmd/video", "-output", name)
	cmd.Dir = filepath.Join(root, "demos", p.Repo)
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", p.Repo, err)
	}
	if err := probe(name); err != nil {
		return fmt.Errorf("%s: %w", p.Repo, err)
	}
	fmt.Println(p.Repo + ": complete")
	return nil
}

func probe(name string) error {
	data, err := exec.Command("ffprobe", "-v", "error", "-show_streams", "-of", "json", name).Output()
	if err != nil {
		return err
	}
	var result struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			Channels  int    `json:"channels"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	var audio, video bool
	for _, s := range result.Streams {
		if s.CodecType == "video" && s.CodecName == "h264" && s.Width > 0 && s.Height > 0 {
			video = true
		}
		if s.CodecType == "audio" && s.CodecName == "aac" && s.Channels == 2 {
			audio = true
		}
	}
	if !audio || !video {
		return fmt.Errorf("expected H.264 video and stereo AAC audio")
	}
	return nil
}

func gallery(output string) error {
	var entries []entry
	for _, p := range productions {
		data, err := os.ReadFile(filepath.Join(output, p.File+".json"))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		var e entry
		if err = json.Unmarshal(data, &e); err != nil {
			return err
		}
		st, err := os.Stat(filepath.Join(output, p.File+".mp4"))
		if err != nil {
			return err
		}
		e.File, e.Poster, e.Repo, e.Size = p.File+".mp4", p.File+".png", p.Repo, st.Size()
		entries = append(entries, e)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(output, "manifest.json"), append(data, '\n'), 0644); err != nil {
		return err
	}
	t := template.Must(template.New("gallery").Parse(page))
	f, err := os.Create(filepath.Join(output, "index.html"))
	if err != nil {
		return err
	}
	if err = t.Execute(f, entries); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

const page = `<!doctype html>
<html lang="fr"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Malakh Software — Vidéos des démos</title>
<style>body{margin:0;background:#10121b;color:#eee;font:16px system-ui,sans-serif}main{max-width:1300px;margin:48px auto;padding:0 24px}h1{font-size:36px}p{color:#b9bfd2}section{display:grid;grid-template-columns:repeat(auto-fit,minmax(320px,1fr));gap:24px}article{background:#1d2131;border-radius:12px;overflow:hidden;padding-bottom:18px}video{display:block;width:100%;aspect-ratio:4/3;background:#000}h2,article p,article a{margin:16px 20px}h2{font-size:20px}a{color:#aac5ff}</style>
<main><h1>Vidéos des démos</h1><p>{{len .}} vidéos avec leur musique. Cuddly comprend le parcours complet ; FR-010 et Second Reality vont jusqu’à leur fin.</p><section>
{{range .}}<article><video controls preload="none" poster="{{.Poster}}" playsinline><source src="{{.File}}" type="video/mp4"></video><h2>{{.Title}}</h2><p>{{.Time}}</p><a href="{{.File}}" download>Télécharger la vidéo</a></article>{{end}}
</section></main></html>`
