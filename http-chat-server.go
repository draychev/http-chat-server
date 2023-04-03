package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const (
	activeUntil = 10 * time.Second
)

type UserName string

type User struct {
	Username UserName  `json:"username"`
	LastPing time.Time `json:"lastPing"`
}

type Message struct {
	Username UserName  `json:"username"`
	Message  string    `json:"message"`
	TimeSent time.Time `json:"timeSent"`
}

type Ping struct {
	Username UserName  `json:"username"`
	TimeSent time.Time `json:"timeSent"`
}

type ChatRoom struct {
	messages []Message
	users    map[UserName]*User
}

func (cr *ChatRoom) AddMessage(message Message) {
	cr.messages = append(cr.messages, message)
}

func (cr *ChatRoom) GetMessages() []Message {
	return cr.messages
}

func (cr *ChatRoom) GetActiveUsers() []*User {
	var activeUsers []*User
	for _, user := range cr.users {
		if time.Since(user.LastPing) > activeUntil {
			continue
		}
		activeUsers = append(activeUsers, user)
	}
	return activeUsers
}

func (cr *ChatRoom) RecordPing(ping Ping) {
	if _, ok := cr.users[ping.Username]; !ok {
		cr.users[ping.Username] = &User{Username: ping.Username}
	}
	cr.users[ping.Username].LastPing = time.Now()
}

func main() {
	chatRoom := &ChatRoom{messages: make([]Message, 0)}

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			messages := chatRoom.GetMessages()
			jsonBytes, err := json.Marshal(messages)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(jsonBytes)
		} else if r.Method == http.MethodPost {
			var message Message
			if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			message.TimeSent = time.Now()
			chatRoom.AddMessage(message)
			w.WriteHeader(http.StatusCreated)
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		users := chatRoom.GetActiveUsers()
		jsonBytes, err := json.Marshal(users)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jsonBytes)
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
		var ping Ping
		if err := json.NewDecoder(r.Body).Decode(&ping); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		chatRoom.RecordPing(ping)
		w.WriteHeader(http.StatusCreated)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
