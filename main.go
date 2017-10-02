package main

import (
	"tumblr-spider/Config"
	"flag"
	"github.com/blang/semver"
	"log"
	"fmt"
	"tumblr-spider/module"
	"os"
	"os/signal"
	"syscall"
	"bufio"
	"strings"
	"time"
	"sync"
)

var VERSION = "1.4.0"

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

//main,summer
func main() {
	verifyFlags()

	walkblock := make(chan struct{})
	go func() {
		fmt.Println("Scanning directory")
		//获取下载目录的所有文件
		module.GetAllCurrentFiles()
		fmt.Println("Done scanning.")
		//解除阻塞
		close(walkblock)
	}()

	//获取要爬取的目标并初始化
	userBlogs := getUsersToDownload()
	//初始化DB
	module.SetupDatabase(userBlogs)
	defer module.Database.Close()
	//设置用户终端和程序退出后的一些操作
	setupSignalInfo()
	//阻塞操作,等待获取用户已经下载的所有文件
	<-walkblock
	//设置一个放置channel的slice,长度等于要爬取的目标长度
	fileChannels := make([]<-chan module.File, len(userBlogs))

	for {
		//channel的长度等于最大下载数*请求速率
		limiter := make(chan time.Time, Config.Cfg.NumDownloaders*Config.Cfg.RequestRate)
		//每秒请求速率(用ticker控制每秒速率,ticker会按速率通过channel发送信号)
		ticker := time.NewTicker(time.Second / time.Duration(Config.Cfg.RequestRate))

		go func() {
			for t := range ticker.C {
				select {
				case limiter <- t:
				default:
				}
			}
		}()

		for i, user := range userBlogs {
			//遍历目标,爬取数据
			fileChan := module.Scrape(user, limiter)
			fileChannels[i] = fileChan
		}

		done := make(chan struct{})
		defer close(done)

		mergedFiles := module.Merge(done, fileChannels)

		if Config.Cfg.UserProgressBar {
			module.Pbar.Start()
		}

		var downloaderWg sync.WaitGroup
		downloaderWg.Add(Config.Cfg.NumDownloaders)

		for i := 0; i < Config.Cfg.NumDownloaders; i++ {
			go func(i int) {
				module.Downloader(i, limiter, mergedFiles)
				downloaderWg.Done()
			}(i)
		}

		downloaderWg.Wait()

		if Config.Cfg.UserProgressBar {
			module.Pbar.Finish()
		}

		module.UpdateDatabaseVersion()

		fmt.Println("Downloading complete.")

		module.Gstats.PrintStats()

		if !Config.Cfg.ServerMode {
			break
		}

		fmt.Println("Sleeping for", Config.Cfg.ServerSleep)
		time.Sleep(Config.Cfg.ServerSleep)
		Config.Cfg.ForceCheck = false
		ticker.Stop()
	}

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

func setupSignalInfo() {
	sigChan := make(chan os.Signal, 1)
	//监听退出和用户中止信息号
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		for {
			s := <-sigChan
			switch s {
			//中止信号
			case syscall.SIGINT:
				module.Database.Close()
				//打印现在爬取的状态
				module.Gstats.PrintStats()
				os.Exit(1)
			//程序退出
			case syscall.SIGQUIT:
				//打印现在爬取的状态(main函数里面自动关闭了database[defer])
				module.Gstats.PrintStats()
			}
		}
	}()
}

func getUsersToDownload() []*module.User {

	//从download读取用户并验证
	fileResults, err := readUserFile()
	if err != nil {
		log.Fatal(err)
	}

	var userBlogs []*module.User

	userBlogs = append(userBlogs, fileResults...)

	if len(userBlogs) == 0 {
		fmt.Fprintln(os.Stderr, "No users detected.")
		os.Exit(1)
	}
	return userBlogs
}

func readUserFile() ([]*module.User, error) {
	file, err := os.Open(Config.Cfg.UserFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var users []*module.User
	scanner := bufio.NewScanner(file)

	//每次获取文件的一行读取,知道遇到EOF停止
	for scanner.Scan() {
		text := strings.Trim(scanner.Text(), " \n\r\t")
		//分割成两份
		split := strings.SplitN(text, " ", 2)

		b, err := module.NewUser(split[0])
		if err != nil {
			log.Println(err)
			continue
		}
		//user模式为:summer(用户名) photo(标签),如果len(split) > 1代表用户填写了标签,爬取标签的资源
		if len(split) > 1 {
			b.Tag = split[1]
		}
		users = append(users, b)
	}
	return users, scanner.Err()
}
