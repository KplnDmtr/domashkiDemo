package main

// сюда писать код

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

var (
	BotToken string
	// @BotFather в телеграме даст вам это
	// BotToken = "_golangcourse_test"
	// урл выдаст вам игрок или хероку
	WebhookURL = "https://eea3-188-123-230-19.ngrok.io"
	// WebhookURL = "http://127.0.0.1:8081"
)

type Task struct {
	creator     *tgbotapi.User
	assignee    *tgbotapi.User
	taskID      int
	description string
}

type Bot struct {
	bot            *tgbotapi.BotAPI
	wg             *sync.WaitGroup
	log            *log.Logger
	currentTaskNum int
	tasks          []Task
}

func (b *Bot) SendMessage(message string, where int64) {
	defer b.wg.Done()
	_, err := b.bot.Send(tgbotapi.NewMessage(where, message))
	if err != nil {
		log.Printf("Ошибка произошла при передаче сообщения: %s пользователю: %d ", message, where)
	}
}

func (b *Bot) FindIndex(taskID int) (index int, exist bool) {
	for index, task := range b.tasks {
		if task.taskID == taskID {
			return index, true
		}
	}
	return index, false
}

func (b *Bot) DeleteTask(taskID int) (ok bool) {
	if index, isExist := b.FindIndex(taskID); isExist {
		b.tasks = slices.Delete(b.tasks, index, index+1)
		return true
	}
	return false
}

func (b *Bot) FindTask(taskID int) (task *Task, exist bool) {
	if index, isExist := b.FindIndex(taskID); isExist {
		return &b.tasks[index], isExist
	}
	return nil, false
}

func (b *Bot) ProcessTaskCommand(message *tgbotapi.Message) {
	if len(b.tasks) == 0 {
		respToChat := "Нет задач"
		b.wg.Add(1)
		go b.SendMessage(respToChat, message.Chat.ID)
		return
	}
	var respToChat string
	for _, task := range b.tasks {
		if task.assignee == nil {
			respToChat += fmt.Sprintf("%d. %s by @%s\n/assign_%d\n\n", task.taskID, task.description, task.creator.UserName, task.taskID)
			continue
		}
		if task.assignee.ID == message.From.ID {
			respToChat += fmt.Sprintf("%d. %s by @%s\nassignee: я\n/unassign_%d /resolve_%d\n\n", task.taskID, task.description, task.creator.UserName, task.taskID, task.taskID)
			continue
		}
		respToChat += fmt.Sprintf("%d. %s by @%s\nassignee: @%s\n\n", task.taskID, task.description, task.creator.UserName, task.assignee.UserName)
	}
	respToChat = strings.TrimSuffix(respToChat, "\n\n")
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func (b *Bot) ProcessNewCommand(message *tgbotapi.Message) {
	fmt.Println(message.CommandArguments())
	newText := message.CommandArguments()
	b.currentTaskNum++
	newTask := Task{
		creator:     message.From,
		description: newText,
		taskID:      b.currentTaskNum,
	}
	b.tasks = append(b.tasks, newTask)
	respToChat := fmt.Sprintf("Задача \"%s\" создана, id=%d", newText, b.currentTaskNum)
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func (b *Bot) ProcessAssignCommand(message *tgbotapi.Message) {
	args := strings.Split(message.Command(), "_")
	if len(args) <= 1 {
		b.wg.Add(1)
		go b.SendMessage("Недостаточно аргументов для команды /assign", message.Chat.ID)
		return
	}
	arg, err := strconv.Atoi(args[1])
	if err != nil {
		b.wg.Add(1)
		go b.SendMessage("Вы не подали номер задачи", message.Chat.ID)
		return
	}

	var oldAssignee *tgbotapi.User
	task, exist := b.FindTask(arg)
	if !exist {
		b.wg.Add(1)
		go b.SendMessage("Не найдено такой задачи", message.Chat.ID)
		return
	}

	oldAssignee = task.assignee
	task.assignee = message.From
	respToNewAssignee := fmt.Sprintf("Задача \"%s\" назначена на вас", task.description)
	b.wg.Add(1)
	go b.SendMessage(respToNewAssignee, message.Chat.ID)

	respToOldAssignee := "Задача \"" + task.description + "\" назначена на @" + task.assignee.UserName
	whereTosend := task.creator.ID
	if oldAssignee != nil {
		whereTosend = oldAssignee.ID
	}
	if whereTosend != message.Chat.ID {
		b.wg.Add(1)
		go b.SendMessage(respToOldAssignee, whereTosend)
	}
}

func (b *Bot) ProccessUnassignCommand(message *tgbotapi.Message) {
	args := strings.Split(message.Command(), "_")
	if len(args) <= 1 {
		b.wg.Add(1)
		go b.SendMessage("Недостаточно аргументов для команды /unassign", message.Chat.ID)
		return
	}
	arg, err := strconv.Atoi(args[1])
	if err != nil {
		b.wg.Add(1)
		go b.SendMessage("Вы не подали номер задачи", message.Chat.ID)
		return
	}

	respToChat := "не найдено такой задачи"

	task, exist := b.FindTask(arg)
	if !exist {
		b.wg.Add(1)
		go b.SendMessage(respToChat, message.Chat.ID)
		return
	}
	if task.assignee == nil || task.assignee.ID != message.From.ID {
		b.wg.Add(1)
		go b.SendMessage("Задача не на вас", message.Chat.ID)
		return
	}

	task.assignee = nil
	respToChat = "Принято"
	respToCreator := fmt.Sprintf("Задача \"%s\" осталась без исполнителя", task.description)
	b.wg.Add(1)
	go b.SendMessage(respToCreator, task.creator.ID)
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func (b *Bot) ProcessResolveCommand(message *tgbotapi.Message) {
	args := strings.Split(message.Command(), "_")
	if len(args) <= 1 {
		b.wg.Add(1)
		go b.SendMessage("Недостаточно аргументов для команды /resolve", message.Chat.ID)
		return
	}
	arg, err := strconv.Atoi(args[1])
	if err != nil {
		b.wg.Add(1)
		go b.SendMessage("Вы не подали номер задачи", message.Chat.ID)
		return
	}

	respToChat := "не найдено такой задачи"
	task, exist := b.FindTask(arg)
	if !exist {
		b.wg.Add(1)
		go b.SendMessage(respToChat, message.Chat.ID)
		return
	}
	if task.assignee == nil || task.assignee.ID != message.From.ID {
		b.wg.Add(1)
		go b.SendMessage("Задача не на вас", message.Chat.ID)
		return
	}

	respToChat = fmt.Sprintf("Задача \"%s\" выполнена", task.description)
	if task.assignee != nil && task.assignee.ID != task.creator.ID {
		respToCreator := fmt.Sprintf("Задача \"%s\" выполнена @%s", task.description, task.assignee.UserName)
		whereToSend := task.creator.ID
		b.wg.Add(1)
		go b.SendMessage(respToCreator, whereToSend)
	}
	task = nil
	if ok := b.DeleteTask(arg); !ok {
		b.wg.Add(1)
		go b.SendMessage("не найдено такой задачи", message.Chat.ID)
		return
	}
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func (b *Bot) ProcessMyCommand(message *tgbotapi.Message) {
	var respToChat string
	for _, task := range b.tasks {
		if task.assignee != nil && task.assignee.ID == message.From.ID {
			respToChat += fmt.Sprintf("%d. %s by @%s\n/unassign_%d /resolve_%d\n\n", task.taskID, task.description, task.creator.UserName, task.taskID, task.taskID)
		}
	}
	respToChat = strings.TrimSuffix(respToChat, "\n\n")

	if respToChat == "" {
		respToChat = "За мной не закреплено команд"
	}
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func (b *Bot) ProcessOwnerCommand(message *tgbotapi.Message) {
	var respToChat string
	for _, task := range b.tasks {
		if task.creator.ID == message.From.ID {
			if task.assignee != nil && task.creator.ID == task.assignee.ID {
				respToChat += fmt.Sprintf("%d. %s by @%s\n/unassign_%d /resolve_%d\n\n", task.taskID, task.description, task.creator.UserName, task.taskID, task.taskID)
				continue
			}
			respToChat += fmt.Sprintf("%d. %s by @%s\n/assign_%d\n\n", task.taskID, task.description, task.creator.UserName, task.taskID)
		}
	}
	respToChat = strings.TrimSuffix(respToChat, "\n\n")

	if respToChat == "" {
		respToChat = "Вы не создавали задач"
	}
	b.wg.Add(1)
	go b.SendMessage(respToChat, message.Chat.ID)
}

func startTaskBot(_ context.Context) error {
	// сюда пишите ваш код

	myBot := Bot{
		log:   log.New(os.Stdout, "MyBot: ", log.Ldate|log.Ltime),
		wg:    &sync.WaitGroup{},
		tasks: make([]Task, 0),
	}
	var err error
	myBot.bot, err = tgbotapi.NewBotAPI(BotToken)
	if err != nil {
		log.Fatalf("NewBotAPI failed: %s", err)
		return err
	}

	myBot.bot.Debug = true
	fmt.Printf("Authorized on account %s\n", myBot.bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(WebhookURL)
	if err != nil {
		log.Fatalf("NewWebhook failed: %s", err)
		return err
	}

	_, err = myBot.bot.Request(wh)
	if err != nil {
		log.Fatalf("SetWebhook failed: %s", err)
		return err
	}

	updates := myBot.bot.ListenForWebhook("/")

	port := "8081"

	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen :" + port)

	// получаем все обновления из канала updates
	for update := range updates {
		if update.Message == nil {
			continue
		}
		commands := strings.Split(update.Message.Command(), "_")
		var command string
		if len(commands) > 0 {
			command = commands[0]
		}
		switch command {
		case "tasks":
			myBot.ProcessTaskCommand(update.Message)
		case "new":
			myBot.ProcessNewCommand(update.Message)
		case "assign":
			myBot.ProcessAssignCommand(update.Message)
		case "unassign":
			myBot.ProccessUnassignCommand(update.Message)
		case "resolve":
			myBot.ProcessResolveCommand(update.Message)
		case "my":
			myBot.ProcessMyCommand(update.Message)
		case "owner":
			myBot.ProcessOwnerCommand(update.Message)
		case "start":
			respToChat := `/tasks показывает список задач

			/new XXX YYY ZZZ - создаёт новую задачу
			
			/assign_$ID - делаеть пользователя исполнителем задачи
			
			/unassign_$ID - снимает задачу с текущего исполнителя
			
			/resolve_$ID - выполняет задачу, удаляет её из списка
			
			/my - показывает задачи, которые назначены на меня
			
			/owner - показывает задачи, которые были созданы мной`
			myBot.wg.Add(1)
			go myBot.SendMessage(respToChat, update.Message.Chat.ID)
		default:
			respToChat := "Не понимаю команду"
			myBot.wg.Add(1)
			go myBot.SendMessage(respToChat, update.Message.Chat.ID)
		}

	}
	myBot.wg.Wait()
	return nil
}

func main() {
	BotToken = os.Getenv("BotToken")
	if BotToken == "" {
		panic("Не найден токен")
	} else {
		fmt.Print(BotToken)
	}
	err := startTaskBot(context.Background())
	if err != nil {
		panic(err)
	}
}
