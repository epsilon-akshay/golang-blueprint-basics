package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-nsq"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/joeshaw/envdecode"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type oauthType struct {
	cred   oauth.Credentials
	client oauth.Client
}

func main() {
	var stoplock sync.Mutex // protects stop
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("Stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	votes := make(chan string) // chan for votes
	publisherStoppedChan := publishToNSQ(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan
}

var connection net.Conn
var reader io.ReadCloser

func dialConn(ctx context.Context, protocol, address string) (net.Conn, error) {
	conn, err := net.DialTimeout(protocol, address, 10*time.Second)
	if err != nil {
		log.Println("could not get conn object with error", err)
		return nil, err
	}
	log.Println("here is the conn object", conn)
	connection = conn
	return conn, nil
}

func closeConn() {
	if reader != nil {
		reader.Close()
	}
	if connection != nil {
		connection.Close()
	}

}

func setupOauth() oauthType {
	var keys struct {
		ConsumerKey    string `env:"SP_TWITTER_KEY,required"`
		ConsumerSecret string `env:"SP_TWITTER_SECRET,required"`
		AccessToken    string `env:"SP_TWITTER_ACCESSTOKEN,required"`
		AccessSecret   string `env:"SP_TWITTER_ACCESSSECRET,required"`
	}
	err := envdecode.Decode(&keys)
	if err != nil {
		log.Fatal("could not fetch env variables")
	}

	return oauthType{
		cred: oauth.Credentials{
			Token:  keys.AccessToken,
			Secret: keys.AccessSecret,
		},
		client: oauth.Client{
			Credentials: oauth.Credentials{
				Token:  keys.ConsumerKey,
				Secret: keys.ConsumerSecret,
			},
		},
	}
}

var (
	client http.Client
	once   sync.Once
	cred   oauthType
)

type Trend struct {
	Name string
}

type tweet struct {
	Trends []Trend
}

func readFromTwitter(vote chan string) {
	u, err := url.Parse("https://api.twitter.com/1.1/trends/place.json")
	if err != nil {
		log.Println("could not parse url")
		return
	}

	query := make(url.Values)
	query.Set("id", "1")
	query.Set("Name", "as")
	qu := query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), strings.NewReader(qu))

	once.Do(func() {
		cred = setupOauth()
		client = http.Client{
			Transport: &http.Transport{
				DialContext: dialConn,
			},
		}
	})
	req.Header.Set("Authorization", cred.client.AuthorizationHeader(&cred.cred, "GET", req.URL, query))
	res, err := client.Do(req)
	if err != nil {
		log.Print("error while making request", err)
	}

	decoder := json.NewDecoder(res.Body)
	for {
		var t []tweet
		err := decoder.Decode(&t)
		if err != nil {
			log.Print("unmarshalling error", err)
			break
		}
		fmt.Println(t)
		vote <- t[0].Trends[0].Name

	}

}

func startTwitterStream(stop chan struct{}, vote chan string) chan struct{} {
	stopChan := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-stop:
				fmt.Print("stopped everything")
				return

			default:
				readFromTwitter(vote)
			}
			stopChan <- struct{}{}
		}
	}()
	return stopChan //to tell outside world that the chan is done
}

func publishToNSQ(trends chan string) chan struct{} {
	q, err := nsq.NewProducer("localhost:4501", nsq.NewConfig())
	if err != nil {
		fmt.Print(err)
	}
	stopChan := make(chan struct{})
	go func() {
		for trend := range trends {
			q.Publish("twitter_trends", []byte(trend))
		}
		stopChan <- struct{}{}
	}()
	return stopChan
}
