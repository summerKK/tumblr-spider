package module

import (
	"time"
	"os"
	"path"
	"tumblr-spider/Config"
	"log"
)

func Downloader(id int, limiter <-chan time.Time, filechan <-chan File) {
	for f := range filechan {
		err := os.MkdirAll(path.Join(Config.Cfg.DownloadDirectory, f.User.String()), 0755)
		if err != nil{
			log.Fatal(err)
		}

		<-limiter
		ShowProgress(f)
		f.Download()
	}
}
