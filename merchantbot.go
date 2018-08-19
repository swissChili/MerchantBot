package main

// add this bot to your server,
// https://discordapp.com/oauth2/authorize?client_id=468216867878600714&scope=bot&permissions=0

// imports all necessary packages
import (
	"./bageldb"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// defines some global variables, such as the MongoDB session, and sever md files
var (
	helpMsg  []byte
	noCmdMsg []byte
	session  *mgo.Session
)

// defines the user type, which will store all user related data

type User struct {
	ID     bson.ObjectId `bson:"_id"`
	UID    int           `bson:"UID"`
	Bagels int           `bson:"Bagels"`
}

func main() {
	fmt.Println("Prg start")
	// reads bot token from `.botkey`
	botkey, tkerr := ioutil.ReadFile(".botkey")
	if tkerr != nil {
		println(tkerr)
		return
	}

	var err error

	session, err = mgo.Dial("localhost")

	helpMsg, err = ioutil.ReadFile("merchant-help.md")
	noCmdMsg, err = ioutil.ReadFile("merchant-nocmd.md")

	// create a new instance of the bot using the bot token
	fmt.Println("Right before discord connection")
	dg, err := discordgo.New("Bot " + string(botkey))
	if err != nil {
		println(err)
		return
	}

	// open the websocket connection
	err = dg.Open()
	if err != nil {
		println(err)
		return
	}

	fmt.Println("Logging in to bot: " + string(botkey))

	dg.AddHandler(messageCreate)

	dg.UpdateStatus(0, "-E")

	// wait until ctrl+c or other quit signal
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// close session without any stuff breaking
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("message!")
	fmt.Println(m.Content)
	// s == the discord session (info about bot)
	// m == message info (contains all information about sent msgs)

	// ignores it's own messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	// some regex to extract the data from commands
	// unescaped version: pay\s+<@!?(\d+)>\s+(\d+)
	sendBagelsRegex := regexp.MustCompile(`(?m)pay\s+<@!?(\d+)>\s+(\d+)`)

	fmt.Println("Message Received!")

	// sets up replies for messages with format `[prompt] [command] [target] [vars]`
	if strings.HasPrefix(m.Content, "-E") {
		fmt.Println("Has prefix")
		fmt.Println("IF THIS GIVES A NIL ERROR THING, RUN `SUDO MONGOD` AND TRY AGAIN")
		// clones the MongoDB instance (this allows async access to db)

		localSession := session.Copy()
		// close the MongoDB session to make sure everythings optimized and such
		defer localSession.Close()

		println("command recieved")

		if m.Content == "-E work" {
			authorID, _ := strconv.Atoi(m.Author.ID)
			workResult := bageldb.Work(localSession, authorID)
			if workResult == 0 {
				balanceEmbed := discordgo.MessageEmbed{
					Color:       0x66f07a,
					Title:       "Success!",
					Description: "You worked successfully! You will not be able to work again for 24 hours. Check your balance with -E balance",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, &balanceEmbed)
			} else if workResult == -1 {
				balanceEmbed := discordgo.MessageEmbed{
					Color:       0xf06666,
					Title:       "HECC!!! Something went wrong!",
					Description: "Looks like something went wrong on our end, silly programmers, messing up the code...",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, &balanceEmbed)
			} else {
				balanceEmbed := discordgo.MessageEmbed{
					Color:       0xf06666,
					Title:       "Looks like you've already worked today!",
					Description: "Try again later!",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, &balanceEmbed)
			}

		} else if m.Content == "-E balance" {
			uid, _ := strconv.Atoi(m.Author.ID)
			userBagels := bageldb.GetUserBagels(localSession, uid)
			balanceEmbed := discordgo.MessageEmbed{
				Color:       0xa37cdc,
				Title:       "Your Balance:",
				Description: "You've got " + strconv.Itoa(userBagels) + "⦿! Get more with -E work",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
				},
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &balanceEmbed)

		} else if len(sendBagelsRegex.FindStringSubmatch(m.Content)) == 3 {
			println(len(sendBagelsRegex.FindStringSubmatch(m.Content)))
			paymentData := sendBagelsRegex.FindStringSubmatch(m.Content)

			paymentTarget, _ := strconv.Atoi(paymentData[1])
			paymentAmount, _ := strconv.Atoi(paymentData[2])
			paymentAmountStr := paymentData[2]
			paymentSender, _ := strconv.Atoi(m.Author.ID)
			//userBagelQuantity := bageldb.GetUserBagels(localSession, paymentSender)

			fmt.Println("In payment quantity if statement")

			bagelTransmission := bageldb.UpdateUserBagels(localSession, paymentTarget, paymentAmount, paymentSender)
			fmt.Println("After bageldb UpdateUserBagels call")
			if bagelTransmission == true {
				fmt.Println("bagelTransmission == true")
				s.ChannelMessageSend(m.ChannelID, "Transaction confirmed by host... processing...")
				// go uses unicode strings so using this character \/ is perfectly fine
				sentEmbed := discordgo.MessageEmbed{
					Color:       0x66f07a,
					Title:       paymentAmountStr + "⦿ have been sent!",
					Description: "<@" + paymentData[1] + ">, check your balance with -E balance!",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, &sentEmbed)
			} else {
				failedEmbed := discordgo.MessageEmbed{
					Color:       0xf06666,
					Title:       "Transaction failed!",
					Description: "Looks like you don't have enough ⦿ to do that! Check your balance with -E balance, and get more bagels with -E work.",
					Thumbnail: &discordgo.MessageEmbedThumbnail{
						URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
					},
				}
				s.ChannelMessageSendEmbed(m.ChannelID, &failedEmbed)
			}
			fmt.Println("After host confirmation")

			// TODO: process DB, confirm DB, updated DB
		} else if m.Content == "-E help" {
			s.ChannelMessageSend(m.ChannelID, string(helpMsg))
			helpEmbed := discordgo.MessageEmbed{
				Color:       0xa37cdc,
				Title:       "Merchant Bot Help:",
				Description: string(helpMsg),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: "https://cdn.discordapp.com/attachments/465229140006666241/476143988718436352/logo.png",
				},
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &helpEmbed)

		} else {
			s.ChannelMessageSend(m.ChannelID, string(noCmdMsg))
			println(m.Content)
		}

	}
}
