package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var rdb *redis.Client

func initRedis() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rdb = redis.NewClient(&redis.Options{
		Addr: "192.168.1.5:6379",
		//Password: "",
		PoolSize: 200,
	})
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		fmt.Println("ping redis failed err:", err)
		return err
	}
	return nil
}

func main() {
	err := initRedis()
	if err != nil {
		fmt.Println("init redis failed err :", err)
		return
	}
	ctx := context.Background()

	var cursor uint64
	keys, cursor, err := rdb.Scan(ctx, cursor, "*", 100).Result()
	if err != nil {
		fmt.Println("scan keys failed err:", err)
		return
	}
	for _, key := range keys {
		//fmt.Println("key:",key)
		sType, err := rdb.Type(ctx, key).Result()
		if err != nil {
			fmt.Println("get type failed :", err)
			return
		}
		fmt.Printf("key :%v ,type is %v\n", key, sType)
		if sType == "string" {
			val, err := rdb.Get(ctx, key).Result()
			if err != nil {
				fmt.Println("get key values failed err:", err)
				return
			}
			fmt.Printf("key :%v ,value :%v\n", key, val)
		} else if sType == "list" {
			val, err := rdb.LPop(ctx, key).Result()
			if err != nil {
				fmt.Println("get list value failed :", err)
				return
			}
			fmt.Printf("key:%v value:%v\n", key, val)
		}
	}
	var imgs []Img
	if result, err := rdb.HGetAll(ctx, "eqq:578779391:msg-476302554:messages").Result(); err != nil {
		fmt.Println("err", err)
	} else {
		fmt.Println(len(result))
		for _, v := range result {
			//fmt.Printf("value is%v\n\n", v)
			var img Img
			if strings.Contains(v, "image") {
				//fmt.Println("content")
				//fmt.Printf("is   %v\n\n", v)
				b := []byte(v)
				if err := json.Unmarshal(b, &img); err != nil {
					fmt.Println(err)
				} else {
					imgs = append(imgs, img)
				}
			}
		}
	}
	for i, v := range imgs {

		if v.File.Url == "" {
			//fmt.Printf("%+v\n", v)
			continue
		} else {
			if strings.Contains(v.File.Url, "?") {
				v.File.Url = strings.Split(v.File.Url, "?")[0]
			}
			fmt.Printf("NO.%d is %v\n", i, v.File.Url)
			fname := strings.Join([]string{strconv.Itoa(i), "jpg"}, ".")
			line := strings.Join([]string{"wget", v.File.Url, "-O", fname}, " ")
			writeline("476302554.sh", line)
		}
	}
	final := "find . -name \"*.jpg\" -size -100k -exec rm {} \\;"
	writeline("476302554.sh", final)

}

type Img struct {
	SenderId  int64  `json:"senderId"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Date      string `json:"date"`
	Id        string `json:"_id"`
	Time      int64  `json:"time"`
	Role      string `json:"role"`
	Title     string `json:"title"`
	Files     []struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"files"`
	AnonymousId   interface{} `json:"anonymousId"`
	Anonymousflag interface{} `json:"anonymousflag"`
	File          struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"file"`
}

func writeline(fname, content string) {
	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0776)
	if err != nil {
		log.Println(err)
	}
	//defer f.Close()
	_, err = f.WriteString(content)
	_, _ = f.WriteString("\n")
	if err != nil {
		log.Println("写文件出错")
	} else {
		//log.Printf("写入%d个字节", n)
	}
}
