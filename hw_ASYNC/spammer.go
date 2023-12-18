package main

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
)

func RunPipeline(cmds ...cmd) {
	wg := &sync.WaitGroup{}
	var chanIn chan interface{}
	for i := 0; i < len(cmds); i++ {
		wg.Add(1)
		chanOut := make(chan interface{})
		go worker(wg, chanIn, chanOut, cmds[i])
		chanIn = chanOut
	}
	wg.Wait()
}

func worker(wg *sync.WaitGroup, in chan interface{}, out chan interface{}, f func(inv, outv chan interface{})) {
	defer wg.Done()
	defer close(out)
	f(in, out)
}

func SelectUsers(in, out chan interface{}) {
	wait := &sync.WaitGroup{}
	idCache := make(map[uint64]bool, 0)
	mutex := &sync.Mutex{}
	getUser := func(email string) {
		defer wait.Done()
		user := GetUser(email)
		mutex.Lock()
		_, exist := idCache[user.ID]
		idCache[user.ID] = true
		mutex.Unlock()
		if !exist {
			out <- user
		}
	}
	for email := range in {
		wait.Add(1)
		go getUser(email.(string))
	}
	wait.Wait()
}

func SelectMessages(in, out chan interface{}) {
	// 	in - User
	// 	out - MsgID
	users := make([]User, GetMessagesMaxUsersBatch)
	var i int
	wait := &sync.WaitGroup{}
	getMessages := func(users []User) {
		defer wait.Done()
		mailsID, err := GetMessages(users...)
		if err == nil {
			for _, mail := range mailsID {
				out <- mail
			}
		} else {
			fmt.Printf("Error occured in function GetMessages called from SelectMessages: %s\n", err)
		}
	}
	for user := range in {
		users[i] = user.(User)
		i++
		if i == GetMessagesMaxUsersBatch {
			wait.Add(1)
			usersLocal := make([]User, len(users))
			copy(usersLocal, users)
			go getMessages(usersLocal)
			i = 0
		}
	}
	if i != 0 {
		wait.Add(1)
		go getMessages(users[:i])
	}
	wait.Wait()
}

func CheckSpam(in, out chan interface{}) {
	// in - MsgID
	// out - MsgData
	wait := &sync.WaitGroup{}
	utilityChannel := make(chan interface{}, HasSpamMaxAsyncRequests)
	defer close(utilityChannel)
	checkingSpam := func(msgID MsgID) {
		defer wait.Done()
		res, err := HasSpam(msgID)
		<-utilityChannel
		if err == nil {
			out <- MsgData{
				msgID,
				res,
			}
		} else {
			fmt.Printf("Error occured in function HasSpam called from CheckSpam: %s\n", err)
		}
	}
	for msgID := range in {
		utilityChannel <- 1
		wait.Add(1)
		go checkingSpam(msgID.(MsgID))
	}
	wait.Wait()
}

func CombineResults(in, out chan interface{}) {
	// in - MsgData
	// out - string
	all := make([]MsgData, 0)
	for msg := range in {
		all = append(all, msg.(MsgData))
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].HasSpam != all[j].HasSpam {
			return all[i].HasSpam
		}
		return all[i].ID < all[j].ID
	})
	for _, one := range all {
		result := strconv.FormatBool(one.HasSpam) + " " + strconv.FormatUint(uint64(one.ID), 10)
		out <- result
	}
}
