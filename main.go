package main

import (
	"github.com/cheggaaa/pb"
	"tumblr-spider/Config"
	"flag"
	"github.com/blang/semver"
	"log"
	"fmt"
	"tumblr-spider/module"
)

var VERSION = "1.4.0"

var (
	pBar = pb.New(0)
)

func init() {
	Config.LoadConfig()

	var numDownloaders int
	if Config.Cfg.NumDownloaders == 0 {
		numDownloaders = 10
	} else {
		numDownloaders = Config.Cfg.NumDownloaders
	}

	var requestRate int
	if Config.Cfg.RequestRate == 0 {
		requestRate = 4
	} else {
		requestRate = Config.Cfg.RequestRate
	}

	var downloaderDirectory string
	downloaderDirectory = Config.Cfg.DownloadDirectory

	flag.BoolVar(&Config.Cfg.IgnorePhotos, "ignore-photos", Config.Cfg.IgnorePhotos, "过滤照片")
	flag.BoolVar(&Config.Cfg.IgnoreVideos, "ignore-videos", Config.Cfg.IgnoreVideos, "过滤视频")
	flag.BoolVar(&Config.Cfg.IgnoreAudio, "ignore-audio", Config.Cfg.IgnoreAudio, "过滤音频")
	flag.BoolVar(&Config.Cfg.UserProgressBar, "p", Config.Cfg.UserProgressBar, "下载的时候显示进度条")
	flag.BoolVar(&Config.Cfg.ForceCheck, "force", Config.Cfg.ForceCheck, "强制检查是否是新文件")
	flag.IntVar(&Config.Cfg.NumDownloaders, "d", numDownloaders, "允许下载的最大数")
	flag.IntVar(&Config.Cfg.RequestRate, "r", requestRate, "每秒的请求速率,不要大于15")
	flag.StringVar(&Config.Cfg.DownloadDirectory, "dir", downloaderDirectory, "下载的目录")

	Config.Cfg.Version = semver.MustParse(VERSION)

	flag.Parse()
}

func main() {
	verifyFlags()

	walkblock := make(chan struct{})
	go func() {
		fmt.Println("Scanning directory")
		module.GetAllCurrentFiles()
		fmt.Println("Done scanning.")
		close(walkblock)
	}()

	fmt.Println(1)

}

func verifyFlags() {
	if Config.Cfg.NumDownloaders < 1 {
		log.Println("非法downloader,使用默认值.")
		Config.Cfg.NumDownloaders = 10
	}

	if Config.Cfg.RequestRate < 1 {
		log.Println("非法 request rate,使用默认值.")
		Config.Cfg.RequestRate = 4
	}

	if Config.Cfg.RequestRate > 15 {
		log.Println("WARRING:请求速率大于15/s,tumblr可能会封锁你的IP,继续你将承担风险.")
	}

}
