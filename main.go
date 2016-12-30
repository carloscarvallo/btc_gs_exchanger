package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/fabioberger/coinbase-go"
	"github.com/joho/godotenv"
)

var envErr = godotenv.Load()
var c = coinbase.ApiKeyClient(os.Getenv("CB_API_KEY"), os.Getenv("CB_SECRET_KEY"))
var config = oauth1.NewConfig(os.Getenv("TW_API_KEY"), os.Getenv("TW_API_SECRET"))
var token = oauth1.NewToken(os.Getenv("TW_ACCESS_TOKEN"), os.Getenv("TW_TOKEN_SECRET"))
var httpClient = config.Client(oauth1.NoContext, token)

func formatCommas(num int) string {
	numString := strconv.Itoa(num)
	re := regexp.MustCompile("(\\d+)(\\d{3})")
	for {
		formatted := re.ReplaceAllString(numString, "$1.$2")
		if formatted == numString {
			return formatted
		}
		numString = formatted
	}
}

func main() {
	tickChan := time.NewTicker(time.Second * 10).C

	client := twitter.NewClient(httpClient)

	if envErr != nil {
		log.Fatal(envErr)
	}

	exchange, exchgErr := c.GetExchangeRate("btc", "pyg")
	if exchgErr != nil {
		log.Fatal(exchgErr)
	}

	exchangePYG := int(exchange)
	sExchangePYG := formatCommas(exchangePYG)

	for {
		select {
		case <-tickChan:
			currentTime := time.Now().Local()
			timeFormatted := currentTime.Format("2006-01-02 15:04:05")

			_, resp, sendErr := client.Statuses.Update(timeFormatted+"\n1 BTC son: "+sExchangePYG+"Gs.", nil)
			fmt.Println("Tweeted at ", timeFormatted)
			fmt.Println("Resp ", resp)

			if sendErr != nil {
				log.Fatal(sendErr)
			}
		}
	}
}
