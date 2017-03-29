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
var cb = coinbase.ApiKeyClient(os.Getenv("CB_API_KEY"), os.Getenv("CB_SECRET_KEY"))
var config = oauth1.NewConfig(os.Getenv("TW_API_KEY"), os.Getenv("TW_API_SECRET"))
var token = oauth1.NewToken(os.Getenv("TW_ACCESS_TOKEN"), os.Getenv("TW_TOKEN_SECRET"))
var httpClient = config.Client(oauth1.NoContext, token)

const (
	fTicketHour  = "00:00"
	sTicketHour  = "06:00"
	tTicketHour  = "12:00"
	frTicketHour = "18:00"

	baseCurrency  = "USD"
	quoteCurrency = "PYG"

	tz = "America/Asuncion"
)

func formatCommas(num int) string {
	numString := strconv.Itoa(num)
	re := regexp.MustCompile(`(\d+)(\d{3})`)
	for {
		formatted := re.ReplaceAllString(numString, "$1.$2")
		if formatted == numString {
			return formatted
		}
		numString = formatted
	}
}

func getExchange(c chan bool, x chan string) {
	for {
		<-c
		exchange, exchgErr := cb.GetExchangeRate(baseCurrency, quoteCurrency)
		if exchgErr != nil {
			log.Fatal(exchgErr)
		}
		exchangePYG := int(exchange)
		s := formatCommas(exchangePYG)
		x <- s
	}
}

func tweetCurrency(xMsg chan string) {
	for {
		exchangePYG := <-xMsg
		client := twitter.NewClient(httpClient)
		currentTime, locationErr := time.LoadLocation(tz)

		if locationErr != nil {
			log.Fatal(locationErr)
		}

		timeFormatted := time.Now().In(currentTime).Format("2006-01-02 15:04")

		statusMessage := fmt.Sprintf("%s\n1 BTC son: %sGs. #btc #gs #pyg #bitcoin #paraguay #guaranies", timeFormatted, exchangePYG)

		_, resp, sendErr := client.Statuses.Update(statusMessage, nil)
		fmt.Println("Tweeted at ", timeFormatted)
		fmt.Println("Resp ", resp)

		if sendErr != nil {
			log.Fatal(sendErr)
		}
	}
}

func getDate(c chan bool) {
	utcLoc, locationErr := time.LoadLocation(tz)

	if locationErr != nil {
		log.Fatal(locationErr)
	}

	utcNow := time.Now().In(utcLoc).Format("15:04")
	if utcNow == fTicketHour || utcNow == sTicketHour || utcNow == tTicketHour || utcNow == frTicketHour {
		c <- true
	}
}

func main() {
	ticker := time.NewTicker(time.Minute)
	dChan := make(chan bool)
	xChan := make(chan string)

	if envErr != nil {
		log.Fatal(envErr)
	}

	for {
		go getDate(dChan)
		go getExchange(dChan, xChan)
		go tweetCurrency(xChan)
		<-ticker.C
	}
}
