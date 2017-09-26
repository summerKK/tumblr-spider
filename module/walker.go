package module

import (
	"os"
	"tumblr-spider/Config"
	"io/ioutil"
	"log"
	"sync"
	"runtime/debug"
)

type FileStatus struct {
	Name     string
	Path     string
	Priority int

	Exists chan struct{}
}

type tracker struct {
	sync.Mutex
	m map[string]FileStatus
}

//保存已经在本地存在的文件
var FileTracker = tracker{m: make(map[string]FileStatus)}

func (f FileStatus) FileInfo() os.FileInfo {
	file, err := os.Stat(f.Path)
	if err != nil {
		debug.PrintStack()
		log.Fatal(err)
	}
	return file
}

//获取已经下载过的文件
func GetAllCurrentFiles() {
	os.MkdirAll(Config.Cfg.DownloadDirectory, 0755)
	//获取下载目录的所有dir
	dirs, err := ioutil.ReadDir(Config.Cfg.DownloadDirectory)
	if err != nil {
		panic(err)
	}
	for _, d := range dirs {
		//忽略文件类型
		if !d.IsDir() {
			continue
		}
		//打开文件夹,遍历文件夹里面内容
		dir, err := os.Open(Config.Cfg.DownloadDirectory + string(os.PathSeparator) + d.Name())
		if err != nil {
			log.Fatal(err)
		}
		//文件夹所有文件,0代表返回所有
		files, err := dir.Readdirnames(0)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			//文件存在
			if info, ok := FileTracker.m[f]; ok {
				p := dir.Name() + string(os.PathSeparator) + f
				//检查文件状态
				checkFile, err := os.Stat(p)
				if err != nil {
					log.Fatal(err)
				}
				//不是同类型文件
				if !os.SameFile(info.FileInfo(), checkFile) {
					os.Remove(p)
					//创建硬链接
					err := os.Link(info.Path, p)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				//新文件
				closedChannel := make(chan struct{})
				close(closedChannel)

				FileTracker.m[f] = FileStatus{
					Name:     f,
					Path:     dir.Name() + string(os.PathSeparator) + f,
					Priority: 0,
					Exists:   closedChannel,
				}
			}
		}

	}
}
