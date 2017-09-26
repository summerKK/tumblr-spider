package module

import (
	"sync"
	"fmt"
)

var Gstats = NewGlobalStats()

type Tracker struct {
	sync.RWMutex
	Blog map[*User]bool
}

type GlobalStats struct {
	//总共下载文件个数
	FilesDownloaded uint64
	//找到的文件
	FilesFound uint64
	//已经存在本地的文件
	AlreadyExists uint64
	//创建的硬链接
	Hardlinked uint64
	//下载的总大小
	BytesDownloaded uint64
	//请求接口(JSON)大小
	BytesOverhead uint64
	//节省的带宽
	ByresSaved uint64
	//正在爬取的目标
	NowScraping Tracker
}

func NewGlobalStats() *GlobalStats {
	return &GlobalStats{
		NowScraping: Tracker{
			Blog: make(map[*User]bool),
		},
	}
}

//打印现在每个正在爬取的目标的状态
//打印每个正在爬取的目标的scraping和downloading
func (g *GlobalStats) PrintStats() {
	g.NowScraping.RLock()
	defer g.NowScraping.RUnlock()

	fmt.Println()
	for k, v := range g.NowScraping.Blog {
		if v {
			//打印每个目标的下载状态
			fmt.Println(k.GetStatus())
		}
	}
	fmt.Println()

	//打印所有目标的下载状况
	fmt.Println(g.FilesDownloaded, "/", g.FilesFound-g.AlreadyExists, "files downloaded.")
	if g.AlreadyExists != 0 {
		//已经下载过的
		fmt.Println(g.AlreadyExists, "previously downloaded.")
	}
	if g.Hardlinked != 0 {
		//创建的硬链接
		fmt.Println(g.Hardlinked, "new hardlinks.")
	}
	fmt.Println("程序一共下载了:", ByteSize(g.BytesDownloaded))
	fmt.Println("API请求下载了:", ByteSize(g.BytesOverhead))
	fmt.Println("节省了带宽:", ByteSize(g.ByresSaved))
}
