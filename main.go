package main

import (
	"fmt"
	"github.com/CptIdea/go-vk-api-2"
	"sync"
	"time"
)

var GroupID = 202576242

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

var bot = vk.NewSession("fdfb0bec49800b941936aaf417d48ae06d129dc539337a4dd13ddfa0a781e19e57b1ffcbfe963c81fc7bd", "5.130")

var NewNoteKB = vk.GenerateKeyBoard("Новая мысль", false, false)

func main() {
	fmt.Println(time.Now().Format(TimeLayout))

	go Checker()
	for {

		updates, err := bot.UpdateCheck(GroupID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, U := range updates.Updates {
			switch U.Object.MessageNew.Text {
			case "Новая мысль":
				_, err = bot.KeySet("setNote", "write", U.Object.MessageNew.FromId)
				if err != nil {
					fmt.Println()
				}

				_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Какую мысль хочешь запомнить?")
				if err != nil {
					fmt.Println(err)
				}
			default:
				value, err := bot.KeyGet("setNote", U.Object.MessageNew.FromId)
				if err != nil {
					fmt.Println(err)
				}
				switch value {
				case "write":
					_, err := bot.KeySet("currentNoteText", U.Object.MessageNew.Text, U.Object.MessageNew.FromId)
					if err != nil {
						fmt.Println(err)
					}

					_, err = bot.KeySet("setNote", "when", U.Object.MessageNew.FromId)
					if err != nil {
						fmt.Println()
					}

					_, err = bot.SendKeyboard(U.Object.MessageNew.FromId, vk.GenerateKeyBoard("!pКогда зайду;Через час,Завтра;Через минуту, Через пять минут", true, false), fmt.Sprintf("Когда тебе напомнить?\n\nФормат: %s\nСейчас: %s", TimeLayout, time.Now().Format(TimeLayout)))
					if err != nil {
						fmt.Println()
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
						fmt.Println()
						fmt.Println("==================================")
						fmt.Printf("Создана мысль пользователем %d\n", U.Object.MessageNew.FromId)
						fmt.Printf("\tТекст:%s\n", text)
						fmt.Printf("\tВремя:%s\n", time.Now().Add(time.Minute*10).Format(TimeLayout))
						fmt.Printf("\tТриггер:online\n")
						Users[U.Object.MessageNew.FromId].Unlock()

						_, err = bot.KeySet("setNote", "wait", U.Object.MessageNew.FromId)
						if err != nil {
							fmt.Println()
						}

						_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Отлично! Жду твоего возвращения!")
						if err != nil {
							fmt.Println(err)
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
							U.Object.MessageNew.Text = time.Now().Add(time.Minute*5).Format(TimeLayout)
						}
						fallthrough

					default:
						_, err := bot.KeySet("currentNoteTime", U.Object.MessageNew.Text, U.Object.MessageNew.FromId)
						if err != nil {
							fmt.Println(err)
						}
						_, err = bot.KeySet("setNote", "online", U.Object.MessageNew.FromId)
						if err != nil {
							fmt.Println()
						}

						_, err = bot.SendKeyboard(U.Object.MessageNew.FromId, vk.GenerateKeyBoard("!gДа,!rНет", true, false), "Ждать когда ты зайдешь?")
						if err != nil {
							fmt.Println()
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
					fmt.Println()
					fmt.Println("==================================")
					fmt.Printf("Создана мысль пользователем %d\n", U.Object.MessageNew.FromId)
					fmt.Printf("\tТекст:%s\n", text)
					fmt.Printf("\tВремя:%s\n", time)
					fmt.Printf("\tТриггер:%s\n", trigger)
					Users[U.Object.MessageNew.FromId].Unlock()

					_, err = bot.KeySet("setNote", "wait", U.Object.MessageNew.FromId)
					if err != nil {
						fmt.Println()
					}

					_, err := bot.SendMessage(U.Object.MessageNew.FromId, "Отлично! Скоро напомню тебе!")
					if err != nil {
						fmt.Println(err)
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
		fmt.Println()
		fmt.Println("==================================")
		fmt.Println("Проверка №", i)
		fmt.Println("Время:", time.Now().Format(TimeLayout))
		for i, user := range Users {
			user.Lock()
			fmt.Println("+++++")
			fmt.Printf("User #%d\n", i)
			fmt.Printf("Notes:%d\n", len(user.Notes))
			var toDel []string
			for i, note := range user.Notes {
				fmt.Printf("\tNote #%d\n", i)
				fmt.Printf("\t\tTrigger:%s\n", note.Trigger)
				fmt.Printf("\t\tText:%s\n", note.Text)
				fmt.Printf("\t\tWhen:%s\n", note.When)
				if note.When == "" || note.Trigger == "" {
					continue
				}
				noteTime, err := time.Parse(TimeLayout, note.When)
				if err != nil {
					fmt.Printf("\t\terror(parse time):%s", err.Error())
					toDel = append(toDel, note.When)
					continue
				}
				var action = "nothing"
				if noteTime.Unix() < time.Now().Unix() {
					switch note.Trigger {
					case "online":
						info, err := bot.GetUsersInfo([]int{user.UserID}, "online")
						if err != nil {
							fmt.Printf("\t\terror(check online):%s", err.Error())
							continue
						}
						if info[0].Online == 1 {
							toDel = append(toDel, note.When)
							_, err = bot.SendMessage(user.UserID, note.Text)
							action = "send notification"
							if err != nil {
								fmt.Printf("\t\terror(send):%s", err.Error())
							}
						}else {
							action = "wait online"
						}
					default:
						toDel = append(toDel, note.When)
						_, err = bot.SendMessage(user.UserID, note.Text)
						action = "send notification"
						if err != nil {
							fmt.Printf("\t\terror(send):%s", err.Error())
						}
					}
				}
				fmt.Printf("\t\tAction:%s\n", action)
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
