package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/zpnk/go-bitly"
)

const (
	threshold			= 10 // sec
	discordCdnBaseURI	= "https://cdn.discordapp.com/attachments/"
)

type EmoticonDictionary struct {
	Mutex		sync.Mutex
	UpdatedAt	time.Time
	Data		map[string]string
}

func updateImageDictionary(e *EmoticonDictionary) error {
	e.Mutex.Lock()
	defer e.Mutex.Unlock()

	now			:= time.Now()
	duration	:= e.UpdatedAt.Sub(now)

	if math.Abs(duration.Seconds()) < threshold {
		fmt.Println("[UPDATE DICT] THRESHOLD LIMIT")
		return nil
	}

	resp, err := http.Get(os.Getenv("DICT_URI"))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true

	e.Data = make(map[string]string)

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("[UPDATE DICT] Unknown error in csv parser")
			continue
		}

		e.Data[record[0]] = record[1]
	}

	e.UpdatedAt = now

	fmt.Println("[UPDATE DICT] Update succesfly.")

	return nil
}

func getLinkByDictionary(s *EmoticonDictionary, key string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if val, exists := s.Data[key]; exists {
		return val, nil
	}

	return "", errors.New("Key not found in dictionary")
}


func main() {
	b := bitly.New(os.Getenv("BITLY_TOKEN"))

	dg, err := discordgo.New("Bot " + os.Getenv("APIKEY"))

	if err != nil {
		log.Fatal("error creating Discord session: ", err)
	}

	emoticonDictionary	:= EmoticonDictionary{}
	emoticonRegex		:= regexp.MustCompile(`;[\w_\?\!]+;`)

	updateImageDictionary(&emoticonDictionary)


	dg.AddHandler(func (s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}

		if m.Content == "ping" {
			s.ChannelMessageSend(m.ChannelID, "pong")
			return
		}

		groups := emoticonRegex.FindSubmatch([]byte(m.Content))
		if len(groups) == 1 {
			emoticonName := string(groups[0][1:len(groups[0])-1])

			updateImageDictionary(&emoticonDictionary)
			emoticonCdnName, err := getLinkByDictionary(
				&emoticonDictionary,
				emoticonName,
			)

			if err == nil {
				emoticonURI := discordCdnBaseURI + emoticonCdnName

				shortURL := emoticonURI

				if emoticonURI[len(emoticonURI) - 3:] != "gif" {
					linkObj, err := b.Links.Shorten(emoticonURI)
					if err != nil {
						log.Print(err)
					} else {
						shortURL = linkObj.URL
					}
				}

				s.ChannelMessageSend(m.ChannelID, shortURL)

				fmt.Println("[Emoticon] 200 custom-emoticon://" + emoticonName)
			} else {
				fmt.Println("[Emoticon] 404 custom-emoticon://" + emoticonName)
			}
		}

	})

	if err := dg.Open(); err != nil {
		log.Fatal("error open websocket: ", err)
	}

	fmt.Println("[SYSTEM] Bot is now running.  Press ^C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

