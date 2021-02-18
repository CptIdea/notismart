package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/CptIdea/go-vk-api-2"
	"github.com/joho/godotenv"
)

func init() {
	logFile, err := os.OpenFile("notify.log", os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		log.Fatal(err)
	}

	//Clear logs every day
	go func() {
		tick := time.NewTicker(time.Hour * 24)
		for {
			<-tick.C
			logFile.Truncate(0)
		}
	}()

	logFile.Truncate(0)

	log.SetOutput(logFile)

	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	var exist bool

	token, exist = os.LookupEnv("VK_TOKEN")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_TOKEN not exist"))
	}

	version, exist = os.LookupEnv("VK_VERSION")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_VERSION not exist"))
	}

	rawGID, exist := os.LookupEnv("VK_GROUP")
	if !exist {
		log.Fatal(fmt.Errorf(".env VK_GROUP not exist"))
	}

	groupID, err = strconv.Atoi(rawGID)
	if err != nil {
		log.Fatal(fmt.Errorf("failed convert VK_GROUP"))
	}
}

var (
	token   = ""
	groupID = 0
	version = ""
)

type UserNotify struct {
	UserID int            `json:"user_id"`
	Notes  []Notification `json:"notes"`
	sync.Mutex
}

type Notification struct {
	Text    string `json:"text"`
	Trigger string `json:"trigger"`
	When    string `json:"when"`
}

var Users = make(map[int]*UserNotify)

var TimeLayout = "2 Jan 2006 15:04:05 -0700 MST"

var bot = vk.NewSession(token, version)

var NewNoteKB = vk.GenerateKeyBoard("Новая мысль", false, false)

func main() {
	log.Println(time.Now().Format(TimeLayout))

	bot = vk.NewSession(token, version)

	go Checker()
	for {

		updates, err := bot.UpdateCheck(groupID)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, U := range updates.Updates {
			switch U.Object.MessageNew.Text {
			case "Начать":
				_, err := bot.SendKeyboard(U.Object.MessageNew.FromId, NewNoteKB, "Привет! Напиши мне \"Новая мысль\" чтобы записать что-нибудь")
				if err != nil {
					log.Println(err)
				}
			case "Дай время":
				_, err := bot.SendMessage(U.Object.MessageNew.FromId, time.Now().Format(TimeLayout))
				if err != nil {
					log.Println(err)
				}
			case "Новая мысль":
				_, err = bot.KeySet("setNote", "write", U.Object.MessageNew.FromId)
				if err != nil {
					log.Println()
				}

				_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Какую мысль хочешь запомнить?")
				if err != nil {
					log.Println(err)
				}
			default:
				value, err := bot.KeyGet("setNote", U.Object.MessageNew.FromId)
				if err != nil {
					log.Println(err)
				}
				switch value {
				case "write":
					_, err := bot.KeySet("currentNoteText", U.Object.MessageNew.Text, U.Object.MessageNew.FromId)
					if err != nil {
						log.Println(err)
					}

					_, err = bot.KeySet("setNote", "when", U.Object.MessageNew.FromId)
					if err != nil {
						log.Println()
					}

					_, err = bot.SendKeyboard(U.Object.MessageNew.FromId, vk.GenerateKeyBoard("!pКогда зайду;Через час,Завтра;Через минуту, Через пять минут;!pДай время", true, false), fmt.Sprintf("Когда тебе напомнить?\n\nФормат: %s\nСейчас: %s", TimeLayout, time.Now().Format(TimeLayout)))
					if err != nil {
						log.Println()
					}

				case "when":

					if Users[U.Object.MessageNew.FromId] == nil {
						Users[U.Object.MessageNew.FromId] = &UserNotify{
							UserID: U.Object.MessageNew.FromId,
							Notes:  nil,
							Mutex:  sync.Mutex{},
						}
					}
					switch U.Object.MessageNew.Text {
					case "Когда зайду":
						Users[U.Object.MessageNew.FromId].Lock()
						text, _ := bot.KeyGet("currentNoteText", U.Object.MessageNew.FromId)
						Users[U.Object.MessageNew.FromId].Notes = append(Users[U.Object.MessageNew.FromId].Notes, Notification{
							Text:    text,
							Trigger: "online",
							When:    time.Now().Add(time.Minute * 10).Format(TimeLayout),
						})
						log.Println()
						log.Println("==================================")
						log.Printf("Создана мысль пользователем %d\n", U.Object.MessageNew.FromId)
						log.Printf("\tТекст:%s\n", text)
						log.Printf("\tВремя:%s\n", time.Now().Add(time.Minute*10).Format(TimeLayout))
						log.Printf("\tТриггер:online\n")
						Users[U.Object.MessageNew.FromId].Unlock()

						_, err = bot.KeySet("setNote", "wait", U.Object.MessageNew.FromId)
						if err != nil {
							log.Println()
						}

						_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Отлично! Жду твоего возвращения!")
						if err != nil {
							log.Println(err)
						}
					case "Через час":
						U.Object.MessageNew.Text = time.Now().Add(time.Hour).Format(TimeLayout)
						fallthrough

					case "Завтра":
						if U.Object.MessageNew.Text == "Завтра" {
							U.Object.MessageNew.Text = time.Now().Add(time.Hour * 24).Format(TimeLayout)
						}
						fallthrough

					case "Через минуту":
						if U.Object.MessageNew.Text == "Через минуту" {
							U.Object.MessageNew.Text = time.Now().Add(time.Minute).Format(TimeLayout)
						}
						fallthrough

					case "Через пять минут":
						if U.Object.MessageNew.Text == "Через пять минут" {
							U.Object.MessageNew.Text = time.Now().Add(time.Minute * 5).Format(TimeLayout)
						}
						fallthrough

					default:
						_, err := bot.KeySet("currentNoteTime", U.Object.MessageNew.Text, U.Object.MessageNew.FromId)
						if err != nil {
							log.Println(err)
						}
						_, err = bot.KeySet("setNote", "online", U.Object.MessageNew.FromId)
						if err != nil {
							log.Println()
						}

						_, err = bot.SendKeyboard(U.Object.MessageNew.FromId, vk.GenerateKeyBoard("!gДа,!rНет", true, false), "Ждать когда ты зайдешь?")
						if err != nil {
							log.Println()
						}
					}
				case "online":
					trigger := "default"
					if U.Object.MessageNew.Text == "Да" {
						trigger = "online"
					}
					Users[U.Object.MessageNew.FromId].Lock()
					text, _ := bot.KeyGet("currentNoteText", U.Object.MessageNew.FromId)
					time, _ := bot.KeyGet("currentNoteTime", U.Object.MessageNew.FromId)
					Users[U.Object.MessageNew.FromId].Notes = append(Users[U.Object.MessageNew.FromId].Notes, Notification{
						Text:    text,
						Trigger: trigger,
						When:    time,
					})
					log.Println()
					log.Println("==================================")
					log.Printf("Создана мысль пользователем %d\n", U.Object.MessageNew.FromId)
					log.Printf("\tТекст:%s\n", text)
					log.Printf("\tВремя:%s\n", time)
					log.Printf("\tТриггер:%s\n", trigger)
					Users[U.Object.MessageNew.FromId].Unlock()

					_, err = bot.KeySet("setNote", "wait", U.Object.MessageNew.FromId)
					if err != nil {
						log.Println()
					}

					_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Отлично! Скоро напомню тебе!")
					if err != nil {
						log.Println(err)
					}
				}
			}
		}
	}

}

func Checker() {
	i := 0
	for {
		i++
		log.Println()
		log.Println("==================================")
		log.Println("Проверка №", i)
		log.Println("Время:", time.Now().Format(TimeLayout))
		for i, user := range Users {
			user.Lock()
			log.Println("+++++")
			log.Printf("User #%d\n", i)
			log.Printf("Notes:%d\n", len(user.Notes))
			var toDel []string
			for i, note := range user.Notes {
				log.Printf("\tNote #%d\n", i)
				log.Printf("\t\tTrigger:%s\n", note.Trigger)
				log.Printf("\t\tText:%s\n", note.Text)
				log.Printf("\t\tWhen:%s\n", note.When)
				if note.When == "" || note.Trigger == "" {
					continue
				}
				noteTime, err := time.Parse(TimeLayout, note.When)
				if err != nil {
					log.Printf("\t\terror(parse time):%s", err.Error())
					toDel = append(toDel, note.When)
					continue
				}
				var action = "nothing"
				if noteTime.Unix() < time.Now().Unix() {
					switch note.Trigger {
					case "online":
						info, err := bot.GetUsersInfo([]int{user.UserID}, "online")
						if err != nil {
							log.Printf("\t\terror(check online):%s", err.Error())
							continue
						}
						if info[0].Online == 1 {
							toDel = append(toDel, note.When)
							_, err = bot.SendMessage(user.UserID, note.Text)
							action = "send notification"
							if err != nil {
								log.Printf("\t\terror(send):%s", err.Error())
							}
						} else {
							action = "wait online"
						}
					default:
						toDel = append(toDel, note.When)
						_, err = bot.SendMessage(user.UserID, note.Text)
						action = "send notification"
						if err != nil {
							log.Printf("\t\terror(send):%s", err.Error())
						}
					}
				}
				log.Printf("\t\tAction:%s\n", action)
			}

			for _, s := range toDel {
				for i, note := range user.Notes {
					if note.When == s {
						user.Notes = append(user.Notes[:i], user.Notes[i+1:]...)
						break
					}
				}
			}

			user.Unlock()
		}

		time.Sleep(time.Minute / 2)
	}
}

//SUPER POWER CODE COMMENT!!!!
//WOW
//WOW
//WOW
//BOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO
//M
//
//
//
//
//
