package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Data struct {
	Channel string 		`json:"channel"`
	Members []string 	`json:"members"`
}

func main() {
	loginId := getNonemptyEnv("LOGIN_ID")
	loginPassword := getNonemptyEnv("LOGIN_PASSWORD")
	mattermostServer := getNonemptyEnv("MATTERMOST_SERVER")

	sessionToken, userId := Login(loginId, loginPassword, mattermostServer)

	fmt.Printf("Session token %s\n", sessionToken)
	fmt.Printf("User ID %s\n", userId)

	var data []Data

	channels := GetAllPublicChannels(mattermostServer, sessionToken)

	for _, channel := range channels {
		fmt.Printf("Channel: %s %s\n", channel.DisplayName, channel.Id)

		JoinChannel(mattermostServer, sessionToken, userId, channel.Id)
		users := GetChannelMembers(mattermostServer, sessionToken, channel.Id)

		var userNames []string
		for _, user := range users {
			userNames = append(userNames, user.Username)
		}

		data = append(data, Data{channel.Name, userNames})

		fmt.Printf("Users:    %d\n", len(users))
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println()
	fmt.Println(string(bytes))
}

func getNonemptyEnv(key string) string {
	value := os.Getenv(key)

	if value == "" {
		fmt.Printf("Needs environment variable '%s' to be set\n", key)
		os.Exit(1)
	}

	return value
}

func Login(id string, password string, server string) (string, string) {
	url := fmt.Sprintf("%s/api/v4/users/login", server)

	requestBody, err := json.Marshal(map[string]string{
		"login_id": id,
		"password": password,
	})

	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	token := resp.Header.Get("Token")
	if token == "" {
		log.Fatal("Expected a Token header")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data map[string]interface{}

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err, string(body))
	}

	return token, data["id"].(string)
}

func JoinChannel(server string, token string, userId string, channelId string) {
	client := http.DefaultClient

	url := fmt.Sprintf("%s/api/v4/channels/%s/members", server, channelId)

	requestBody, err := json.Marshal(map[string]string{
		"user_id": userId,
		"post_root_id": "",
	})

	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

type Channel struct {
	Id string			`json:"id"`
	DisplayName string 	`json:"display_name"`
	Name string 		`json:"name"`
}

func GetAllPublicChannels(server string, token string) []Channel {
	client := http.DefaultClient

	url := fmt.Sprintf("%s/api/v4/teams/q7gsredccbfjjre91dr58zfnoc/channels?pageSize=10000", server)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []Channel

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err, string(body))
	}

	return data
}

type User struct {
	Id string 			`json:"user_id"`
	Username string 	`json:"username"`
}

func GetChannelMembers(server string, token string, channelId string) []User {
	return getChannelMembers(server, token, channelId, 0)
}

func getChannelMembers(server string, token string, channelId string, page int) []User {
	client := http.DefaultClient

	url := fmt.Sprintf("%s/api/v4/users?in_channel=%s&page=%d&pageSize=200&active=true", server, channelId, page)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data []User

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err, string(body))
	}

	if len(data) == 200 {
		return append(data, getChannelMembers(server, token, channelId, page + 1)...)
	}

	return data
}
