package module

import (
	"path"
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"tumblr-spider/Config"
	"os"
	"time"
	"sync/atomic"
)

var (
	gfyRequest = "https://gfycat.com/cajax/get/%s"
)

type File struct {
	User          *User
	URL           string
	UnixTimestamp int64
	Filename      string
}

type Gfy struct {
	GfyItem struct {
		Mp4URL  string `json:"mp4Url"`
		WebmURL string `json:"webmUrl"`
	} `json:"gfyItem"`
}

func NewFile(URL string) File {
	return File{
		URL:      URL,
		Filename: path.Base(URL),
	}
}

func getGfycatURL(slug string) string {
	gfyURL := fmt.Sprintf(gfyRequest, slug)

	var resp *http.Response
	for {
		resp2, err := http.Get(gfyURL)
		if err != nil {
			log.Println("GetGfycatURL:", err)
		} else {
			resp = resp2
			break
		}
	}
	defer resp.Body.Close()

	gfyData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var gfy Gfy

	err = json.Unmarshal(gfyData, &gfy)
	if err != nil {
		log.Fatal("Gfycat Unmarsha1:", err)
	}
	return gfy.GfyItem.Mp4URL
}

func getGfycatFiles(b, slug string) []File {
	var files []File
	regexResult := gfycatSearch.FindStringSubmatch(b)
	if regexResult != nil {
		for i, v := range regexResult[1:] {
			gfyFile := NewFile(getGfycatURL(v))
			if slug != "" {
				gfyFile.Filename = fmt.Sprintf("%s_gfycat_%02d.mp4", slug, i+1)
				files = append(files, gfyFile)
			}
		}
	}
	return files
}

func (f File) Download() {
	filepath := path.Join(Config.Cfg.DownloadDirectory, f.User.String(), path.Base(f.Filename))
	var resp *http.Response
	var err error
	var pic []byte

	for {
		resp, err = http.Get(f.URL)
		if err != nil {
			log.Println(err)
			continue
		}
		defer resp.Body.Close()

		pic, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("READALL:", err)
			continue
		}
		break
	}

	err = ioutil.WriteFile(filepath, pic, 0644)
	if err != nil {
		log.Fatal("writeFile:", err)
	}

	err = os.Chtimes(filepath, time.Now(), time.Unix(f.UnixTimestamp, 0))
	if err != nil {
		log.Println(err)
	}

	FileTracker.Signal(f.Filename)

	Pbar.Increment()

	f.User.DownloadWg.Done()
	atomic.AddUint64(&f.User.FilesProcessed, 1)
	atomic.AddUint64(&Gstats.FilesDownloaded, 1)
	atomic.AddUint64(&Gstats.BytesDownloaded, uint64(len(pic)))

}
