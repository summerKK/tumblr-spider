package module

import (
	"os"
	"tumblr-spider/Config"
	"io/ioutil"
	"log"
	"sync"
	"runtime/debug"
	"path"
	"fmt"
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
		fmt.Println(files)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			//记录在队列里面的文件是否已经下载过
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

func (t *tracker) Add(name, path string) bool {
	t.Lock()
	defer t.Unlock()

	//文件已经在队列中
	if _, ok := t.m[name]; ok {
		return true
	}

	//添加文件到队列中
	t.m[name] = FileStatus{
		Name:     name,
		Path:     path,
		Priority: 0,
		Exists:   make(chan struct{}),
	}
	return false
}

func (t *tracker) WaitForDownload(name string) {
	<-t.m[name].Exists
}

func (t *tracker) Link(oldfilename, newpath string) {
	t.Lock()
	defer t.Unlock()
	info := t.m[oldfilename]
	newInfo := FileInfo(newpath)
	if !os.SameFile(info.FileInfo(), newInfo) {
		err := os.MkdirAll(path.Dir(newpath), 0755)
		if err != nil {
			log.Fatal(err)
		}
		os.Remove(newpath)
		err = os.Link(info.Path, newpath)
		if err != nil {
			log.Println(info.Path, "-", newpath)
			log.Fatal("t.link", err)
		}
	}
}

func (t *tracker) Signal(file string) {
	close(t.m[file].Exists)
}

func FileInfo(s string) os.FileInfo {
	file, err := os.Stat(s)
	if err != nil {
		log.Fatal("walker.go/FileInfo", err)
	}
	return file
}
