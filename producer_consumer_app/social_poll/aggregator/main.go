package main

import (
	"fmt"
	"github.com/bitly/go-nsq"
	"gopkg.in/mgo.v2"
	"log"
)

type TwitterTrend struct {
	Trend string `bson:"trend"`
}

func main() {
	if err := dialDB(); err != nil {
		log.Fatalln("failed to dial MongoDB:", err)
	}
	c := db.DB("twitter").C("trends")


	consumeFromNSQ(c)
	defer closedb()
}

var db *mgo.Session

func dialDB() error {
	var err error
	log.Println("dialing mongodb host")
	db, err = mgo.Dial("localhost")
	return err
}
func closedb() {
	db.Close()
	log.Println("closed database connection")
}

func consumeFromNSQ(c *mgo.Collection) {
	q, err := nsq.NewConsumer("twitter_trend", "group1", nsq.NewConfig())
	if err != nil {
		panic(err)
	}
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		trend := string(m.Body)
		fmt.Print("trend is", trend)
		addTrends(c, trend)

		return nil
	}))

	if err := q.ConnectToNSQLookupd("localhost:4161");
		err !=nil {
		log.Fatal(err)
		return
	}
}

func addTrends(coll *mgo.Collection, trend string) {
	trended := TwitterTrend{
		Trend: trend,
	}
	err := coll.Insert(trended)
	if err !=nil {
		panic(err)
	}
}
