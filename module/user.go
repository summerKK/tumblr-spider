package module

import (
	"sync"
	"fmt"
	"regexp"
	"errors"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

var userVerificationRegex = regexp.MustCompile(`^[A-za-z0-9\-]+$`)

type UserAction int

const (
	_ UserAction = iota
	//用户爬取状态,用户默认状态
	Scraping
	//用户已经完成了scraping,但是文件还在队列中(正在下载或者等待下载)
	Downloading
)

// 保存要爬取的用户的一些详细信息
type User struct {
	Name, Tag     string
	LastPostID    int64
	HighestPostID int64
	Status        UserAction

	sync.RWMutex
	FilesFoud      uint64
	FilesProcessed uint64

	Done                 chan struct{}
	FileChannel          chan File
	IdProccessChan       chan int64
	FileProcessChan      chan int
	ScrapeWg, DownloadWg sync.WaitGroup
}

func NewUser(name string) (*User, error) {
	if !userVerificationRegex.MatchString(name) {
		return nil, errors.New("newUser:用户格式不正确:" + name)
	}
	query := fmt.Sprintf("https://api.tumblr.com/v2/blog/%s.tumblr.com/avatar/16", name)
	resp, err := http.Get(query)
	if err != nil {
		return nil, errors.New("newUser:无法连接到tumblr去验证用户:" + name)
	}
	defer resp.Body.Close()

	var js map[string]interface{}
	contents, _ := ioutil.ReadAll(resp.Body)

	if json.Unmarshal(contents, &js) == nil {
		return nil, errors.New("newUser:用户无法找到:" + name)
	}

	//用户验证通过
	u := &User{
		Name:          name,
		LastPostID:    0,
		HighestPostID: 0,
		//Scraping用户默认爬取状态
		Status:          Scraping,
		Done:            make(chan struct{}),
		IdProccessChan:  make(chan int64, 10),
		FileProcessChan: make(chan int, 10),
	}
	Gstats.NowScraping.Blog[u] = true
	return u, nil
}

//打印爬取目标的状态
func (u *User) GetStatus() string {
	isLimited := ""
	if u.FilesFoud-u.FilesProcessed > MaxQueueSize {
		isLimited = "[ LIMITED ]"
	}

	return fmt.Sprint(u.Name, " - ", u.Status, " ( ", u.FilesProcessed, "/", u.FilesFoud, " ) ", isLimited)
}

func (u *User) DoneScrap() {
	u.DownloadWg.Wait()
	fmt.Println("Done downloading for", u.Name)
	//停止helper function
	close(u.Done)
	Gstats.NowScraping.Blog[u] = false
	updateDatabase(u.Name, u.HighestPostID)
}

func (u *User) FinishScraping(i int) {
	fmt.Println("Done scraping for", u.Name, " (", i, "pages )")
	u.ScrapeWg.Wait()
	u.Status = Downloading

	close(u.FileChannel)
	go u.DoneScrap()
}

func (u *User) UpdateHighestPost(i int64) {
	u.RLock()
	//如果大于最大的postID就更新
	if i > u.HighestPostID {
		u.RUnlock()
		u.Lock()
		if i > u.HighestPostID {
			u.HighestPostID = i
		}
		u.Unlock()
		u.RLock()
	}
	u.RUnlock()
}

func (u *User) Queue(p Post) {
	files :=
}