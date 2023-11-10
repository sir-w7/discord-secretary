package main

import (
	"fmt"
	"path/filepath"

	"log"

	"os"
	"os/signal"

	"syscall"

	"math/rand"
	"time"

	"strings"

	"sync"

	"github.com/bwmarrin/discordgo"
)

var exeDir string

var messageList []string
var lastLoaded time.Time

var credentials []string

var messageMode string

var userLastMsg = make(map[string]time.Time)
var channelOwnerAt = make(map[string]time.Time)

var mutex sync.Mutex

const TOKEN = 0
const BOT_NAME = 1
const USER_ID = 2

func main() {
	exeDir = getExeDir()
	fmt.Println(exeDir)

	data, err := os.ReadFile(exeDir + "credentials.txt")
	if err != nil {
		log.Println("Error:", err.Error())
	}
	credentials = strings.Split(string(data), "\n")

	discord, err := discordgo.New(credentials[TOKEN])
	if err != nil {
		fmt.Println("Failed to create a Discord session.")
		fmt.Println("Error:" + err.Error())
		return
	}
	if len(os.Args) > 1 {
		messageMode = os.Args[1]
	} else {
		messageMode = "idle_messages.txt"
	}
	fmt.Println("Selected " + messageMode + " message list.")
	lastLoaded, messageList = loadMessages(messageMode)

	discord.AddHandler(messageCreate)

	err = discord.Open()
	if err != nil {
		fmt.Println("Failed to open a Discord socket.")
		fmt.Println("Error:" + err.Error())
		return
	}
	defer discord.Close()
	log.Println(credentials[BOT_NAME], "is now online.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println(credentials[BOT_NAME], "has left the chat.")
}

func getExeDir() string {
	exePath, _ := os.Executable()
	return filepath.Dir(exePath) + "/"
}

func loadMessages(msgType string) (time.Time, []string) {
	path := exeDir + "messages/" + msgType
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error:", err.Error())
		return time.Time{}, nil
	}

	return fileLastMod(path), strings.Split(string(data), "\n\n")
}

func fileLastMod(path string) time.Time {
	inf, _ := os.Stat(path)
	return inf.ModTime()
}

func isDm(session *discordgo.Session, msg *discordgo.MessageCreate) (bool, error) {
	channel, err := session.State.Channel(msg.ChannelID)
	if err != nil {
		if channel, err = session.Channel(msg.ChannelID); err != nil {
			return false, err
		}
	}

	if channel.Type == discordgo.ChannelTypeDM {
		return true, nil
	} else if channel.Type == discordgo.ChannelTypeGroupDM {
		for _, mentioned_user := range msg.Mentions {
			if mentioned_user.Username+"#"+mentioned_user.Discriminator == credentials[USER_ID] {
				return true, nil
			}
		}
	}

	return false, nil
}

func messageCreate(session *discordgo.Session, msg *discordgo.MessageCreate) {
	dm, err := isDm(session, msg)
	if err != nil {
		fmt.Println("Failed to check channel type.")
		fmt.Println("Error:", err.Error())
		return
	}

	if dm == false {
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	if msg.Author.ID == session.State.User.ID {
		channelOwnerAt[msg.ChannelID] = time.Now()
		return
	}

	if lastLoaded != fileLastMod("messages/"+messageMode) {
		lastLoaded, messageList = loadMessages(messageMode)
		fmt.Println("Reloaded " + messageMode + " message list")
	}

	user := msg.Author.Username

	if time.Since(userLastMsg[user]) > 3*time.Minute || userLastMsg[user].IsZero() {
		log.Println("Message from:", user)

		sendRandMessage(session, msg)
		sendMessage(session, msg)

		userLastMsg[user] = time.Now()
	}
}

func sendRandMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	session.ChannelTyping(msg.ChannelID)
	time.Sleep(2 * time.Second)
	session.ChannelMessageSend(msg.ChannelID, messageList[rand.Intn(len(messageList))])
}

func sendMessage(session *discordgo.Session, msg *discordgo.MessageCreate) {
	session.ChannelTyping(msg.ChannelID)
	time.Sleep(1 * time.Second)
	session.ChannelMessageSend(msg.ChannelID, "\\- Best wishes, "+credentials[BOT_NAME])
}
