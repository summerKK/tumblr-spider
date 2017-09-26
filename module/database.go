package module

import (
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

var Database *bolt.DB

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
