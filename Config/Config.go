package Config

import (
	"time"
	"github.com/blang/semver"
	"github.com/burntsushi/toml"
	"log"
)

var Cfg Config

type Config struct {
	NumDownloaders    int           `toml:"num_downloaders"`
	RequestRate       int           `toml:"rate"`
	ForceCheck        bool          `toml:"force"`
	ServerMode        bool          `toml:"server_mode"`
	ServerSleep       time.Duration `toml:"sleep_time"`
	DownloadDirectory string        `toml:directory`

	IgnorePhotos    bool `toml:"ignore_photos"`
	IgnoreVideos    bool `toml:"ignore_videos"`
	IgnoreAudio     bool `toml:ignore_audio`
	UserProgressBar bool `toml:"user_progress_bar"`

	Version semver.Version
}

func LoadConfig() {
	var err error
	if _, err = toml.DecodeFile("./Config/config.toml", &Cfg); err != nil {
		log.Fatal(err)
	}
	if Cfg.DownloadDirectory == "" {
		Cfg.DownloadDirectory = "L:\\tumblr"
	}
}
