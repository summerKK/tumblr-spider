package module

import (
	"github.com/boltdb/bolt"
	"log"
	"strconv"
	"tumblr-spider/Config"
	"fmt"
	"github.com/blang/semver"
)

var Database *bolt.DB

func SetupDatabase(userBlogs []*User) {
	db, err := bolt.Open("tumblr-update.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	Database = db

	err = db.Update(func(tx *bolt.Tx) error {
		b, boltErr := tx.CreateBucketIfNotExists([]byte("tumblr"))
		if boltErr != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		for _, blog := range userBlogs {
			v := b.Get([]byte(blog.Name))
			if len(v) != 0 {
				blog.LastPostID, _ = strconv.ParseInt(string(v), 10, 64)
				blog.UpdateHighestPost(blog.LastPostID)
			}
		}

		storedVersion := string(b.Get([]byte("__VERSION__")))

		v, err := semver.Parse(storedVersion)
		if err != nil {
			log.Println(err)
		}

		checkVersion(v)

		return nil
	})
}
func checkVersion(version semver.Version) {
	fmt.Println("current version is", Config.Cfg.Version)
	if version.LT(Config.Cfg.Version) {
		Config.Cfg.ForceCheck = true
		log.Println("Checking entire tumblrblog due to new version.")
	}
}

func updateDatabase(name string, id int64) {
	err := Database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tumblr"))
		if b == nil {
			log.Println(`Bucket "tumblr" 在数据库不存在,程序可能出问题了!`)
		}

		if err := b.Put([]byte(name), []byte(strconv.FormatInt(id, 10))); err != nil {
			return err
		}
		return nil

	})

	if err != nil {
		log.Fatal("database:", err)
	}
}

func UpdateDatabaseVersion() {
	err := Database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("tumblr"))
		if b == nil {
			log.Println(`Bucket "tumblr" 在数据库不存在,程序可能出问题了!`)
		}
		if err := b.Put([]byte(`_VERSION_`), []byte(Config.Cfg.Version.String())); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal("database:", err)
	}
}
