package main

import (
	"fmt"
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

var exe_dir string

var message_list []string
var last_loaded time.Time

var credentials []string

var message_mode string

var user_last_msg = make(map[string]time.Time)
var channel_owner_at = make(map[string]time.Time)

var mutex sync.Mutex

const TOKEN = 0
const BOT_NAME = 1
const USER_ID = 2

func main() {
	data, err := os.ReadFile("credentials.txt")
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
		message_mode = os.Args[1]
	} else {
		message_mode = "idle_messages.txt"
	}
	fmt.Println("Selected " + message_mode + " message list.")
	last_loaded, message_list = load_messages(message_mode)

	discord.AddHandler(message_create)

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

func get_exe_dir() string {

}

func load_messages(msg_type string) (time.Time, []string) {
	path := "messages/" + msg_type
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Error:", err.Error())
		return time.Time{}, nil
	}

	return file_last_mod(path), strings.Split(string(data), "\n\n")
}

func file_last_mod(path string) time.Time {
	inf, _ := os.Stat(path)
	return inf.ModTime()
}

func is_dm(session *discordgo.Session, msg *discordgo.MessageCreate) (bool, error) {
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

func message_create(session *discordgo.Session, msg *discordgo.MessageCreate) {
	dm, err := is_dm(session, msg)
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
		channel_owner_at[msg.ChannelID] = time.Now()
		return
	}

	if last_loaded != file_last_mod("messages/"+message_mode) {
		last_loaded, message_list = load_messages(message_mode)
		fmt.Println("Reloaded " + message_mode + " message list")
	}

	user := msg.Author.Username + "#" + msg.Author.Discriminator
	if (user_last_msg[user].IsZero() || time.Since(user_last_msg[user]) > 5*time.Minute) &&
		(channel_owner_at[msg.ChannelID].IsZero() || time.Since(channel_owner_at[msg.ChannelID]) > 3*time.Minute) {
		log.Println("Message from:", user)

		session.ChannelTyping(msg.ChannelID)
		time.Sleep(5 * time.Second)
		session.ChannelMessageSend(msg.ChannelID, message_list[rand.Intn(len(message_list))])

		time.Sleep(2 * time.Second)
		session.ChannelMessageSend(msg.ChannelID, "- Best wishes, "+credentials[BOT_NAME])

		user_last_msg[user] = time.Now()
	}
}
