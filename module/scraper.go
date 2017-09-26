package module

import (
	"time"
	"sync"
	"net/url"
	"fmt"
	"log"
	"strconv"
	"tumblr-spider/Config"
	"github.com/cheggaaa/pb"
	"net/http"
	"io/ioutil"
	"sync/atomic"
	"encoding/json"
)

var (
	Pbar         = pb.New(0)
	PostParseMap = map[string]func(post Post) []File{
		"photo":   parsePhotoPost,
		"answer":  parseAnswerPost,
		"regular": parseRegularPost,
		"video":   parseVideoPost,
	}
)

const MaxQueueSize = 10000

func Scrape(u *User, limiter <-chan time.Time) <-chan File {
	var once sync.Once
	u.FileChannel = make(chan File, MaxQueueSize)

	go func() {
		done := make(chan struct{})
		closeDone := func() { close(done) }
		var i, numPosts int
		defer func() {
			u.FinishScraping(i)
		}()

		for i = 1; ; i++ {
			if shouldFinishScraping(limiter, done) {
				return
			}

			tumblrURL := makeTumblrURL(u, i)
			ShowProgress(u.Name, "is on page", i, "/", (numPosts/50)+1)

			var resp *http.Response
			var err error
			var contents []byte

			for {
				resp, err = http.Get(tumblrURL.String())
				if err != nil {
					log.Println("http.get:", u, err)
					continue
				}
				contents, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println("ReadAll:", u, err, "(", len(contents), "/", resp.ContentLength, ")")
					continue
				}
				err = resp.Body.Close()
				if err != nil {
					fmt.Println(err)
				}
				break
			}
			atomic.AddUint64(&Gstats.BytesOverhead, uint64(len(contents)))

			//过滤掉JS里面的不必要的字符串
			contents = TrimJS(contents)

			var blog TumblrLog
			err = json.Unmarshal(contents, &blog)
			if err != nil {
				ioutil.WriteFile("json_error.txt", contents, 0644)
				log.Println("Unmarshal:", err)
			}
			numPosts = blog.TotalPosts
			u.ScrapeWg.Add(1)
			defer u.ScrapeWg.Done()

			for _, post := range blog.Posts {
				id, err := post.ID.Int64()
				if err != nil {
					log.Println(err)
				}
				u.UpdateHighestPost(id)
				//如果没有开启强制检查,获取的id<上一次的postID
				if !Config.Cfg.ForceCheck && id <= u.LastPostID {
					once.Do(closeDone)
					return
				}
				u.Queue(post)
			}
		}

	}()
}

func shouldFinishScraping(limiter <-chan time.Time, done <-chan struct{}) bool {
	select {
	case <-done:
		return true
	default:
		select {
		case <-done:
			return true
			//收到limiter的请求,爬取网站
		case <-limiter:
			return false
		}
	}
}

func makeTumblrURL(user *User, i int) *url.URL {
	base := fmt.Sprintf("https://%s.tumblr.com/api/read/json", user.Name)
	tumblrURL, err := url.Parse(base)
	if err != nil {
		log.Fatal("tumblrURL:", err)
	}

	vals := url.Values{}
	vals.Set("num", "50")
	vals.Add("start", strconv.Itoa((i-1)*50))

	if user.Tag != "" {
		vals.Add("tagged", user.Tag)
	}

	tumblrURL.RawQuery = vals.Encode()
	return tumblrURL
}

func ShowProgress(s ...interface{}) {
	if Config.Cfg.UserProgressBar {
		Pbar.Update()
	} else if len(s) > 0 {
		fmt.Println(s...)
	}
}

func TrimJS(c []byte) []byte {
	// The length of "var tumblr_api_read = " is 22.
	return c[22:len(c)-2]
}

func ParseDataForFiles(p Post) (files []File) {
	fn, ok := PostParseMap[p.Type]
	if ok {
		files = fn(p)
	}
	return
}
