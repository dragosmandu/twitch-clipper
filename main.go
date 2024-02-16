package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// - TYPES

type Reponse interface{}

type TwitchData struct {
	Id          *string `json:"id"`
	Username    *string `json:"username"`
	ClientAppId *string `json:"client_app_id"`
}

type TwitchAuth struct {
	DeviceCode      string `json:"device_code"`
	ExpIn           int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationUrl string `json:"verification_uri"`
}

type TwitchUser struct {
	Id string `json:"id"`
}

type TwitchClip struct {
	Id      string `json:"id"`
	EditUrl string `json:"edit_url"`
}

// - VARIABLES

var twitchData TwitchData
var twitchAuth TwitchAuth
var twitchUser TwitchUser = TwitchUser{Id: "770869829"}
var twitchClip TwitchClip

// - MAIN

func main() {
	for err := updateTwitchData(); err != nil; {
	}

	err := updateTwitchAuth()

	if err != nil {
		panic(err)
	}

	// err = updateTwitchUser()

	// if err != nil {
	// 	panic(err)
	// }

	// listenForTwitchClipCommand()
}

// - TWITCH

func listenForTwitchClipCommand() {
	txt, err := readStdin("create a new clip by writing 'i have small pp' in the terminal :> \n")

	if err != nil {
		fmt.Println("u r the proble, not the solution. try again")
		listenForTwitchClipCommand()
	} else if strings.ToLower(*txt) != "i have small pp" {
		fmt.Println("the commaand u wrote is soooo wrong, i cannot even")
		listenForTwitchClipCommand()
	}

	fmt.Println("creating your new little clip...")

	err = createTwitchClip()

	if err != nil {
		fmt.Println(err)
		listenForTwitchClipCommand()
		return
	}

	fmt.Printf("clip created... %v\n", twitchClip)

	listenForTwitchClipCommand()
}

func createTwitchClip() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/clips?broadcaster_id=%v", twitchUser.Id)
	headers := map[string]string{"Authorization": "Bearer " + "access_token", "Client-Id": "twitchUserProvidedData.clientAppId"}
	resp, err := do[map[string][]TwitchClip]("POST", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		twitchClip = (*resp)["data"][0]
		return nil
	}

	return errors.New("Ok, i tried buddy. idk how to tell u this but the clipping faileed...miserably. but dont worry, you can try again #fingercrossed")
}

func updateTwitchUser() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", "twitchUserProvidedData.username")
	headers := map[string]string{"Authorization": "Bearer " + "access-token", "Client-Id": "twitchUserProvidedData.clientAppId"}
	resp, err := do[map[string][]TwitchUser]("GET", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		twitchUser = (*resp)["data"][0]
		return nil
	}

	return errors.New("Something went wrong, but u not the problem...maybe....OK, it s you")
}

func updateTwitchAuth() error {
	url := fmt.Sprintf("https://id.twitch.tv/oauth2/device?client_id=%v&scopes=clips:edit", *twitchData.ClientAppId)
	resp, err := do[TwitchAuth]("POST", url, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if err != nil {
		return err
	}
	// curl --location 'https://id.twitch.tv/oauth2/token' \
	// --form'client_id="0mmkby2n450y6ho3s2b4xth9fjggz1"' \
	// --form'scope="channel:manage:broadcast"' \
	// --form'device_code="ike3GM8QIdYZs43KdrWPIO36LofILoCyFEzjlQ91"' \
	// --form'grant_type="urn:ietf:params:oauth:grant-type:device_code"'
	return nil
}

// - TWITCH DATA

func updateTwitchData() error {
	const twitchDataFileName = "twitch-data.json"
	data, _ := os.ReadFile(twitchDataFileName)

	err := json.Unmarshal(data, &twitchData)

	if err == nil && twitchData.Username != nil && twitchData.ClientAppId != nil {
		command, _ := readStdin("I found your twitch username and client app id, do you want to update them? if yes press 'y': ")

		if command == nil || *command != "y" {
			return nil
		}
	}

	twitchUsername, err := getTwitchUsername()

	if err != nil {
		return err
	}

	twitchClientAppId, err := getTwitchClientAppId()

	if err != nil {
		return err
	}

	twitchData = TwitchData{Id: nil, Username: twitchUsername, ClientAppId: twitchClientAppId}
	data, err = json.MarshalIndent(twitchData, "", " ")

	if err != nil {
		return nil
	}

	return os.WriteFile(twitchDataFileName, data, 0644)
}

func getTwitchClientAppId() (*string, error) {
	twitchClientAppId, err := readStdin("Please tell me your Twitch client app id, or else...: ")

	if err != nil {
		return nil, err
	}

	return twitchClientAppId, nil
}

func getTwitchUsername() (*string, error) {
	twitchUsername, err := readStdin("Please tell me your Twitch username: ")

	if err != nil {
		return nil, err
	}

	isTwitchUsernameValid := regexp.MustCompile(`^[a-zA-Z0-9\_]+$`).MatchString(*twitchUsername)

	if !isTwitchUsernameValid {
		return nil, errors.New("Dummy, your twitch username was invalid :) \n")
	}

	return twitchUsername, nil
}

// - HELPERS

func readStdin(msg string) (*string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(msg)

	txt, err := reader.ReadString('\n')

	if err != nil {
		return nil, err
	}

	txt = strings.TrimSpace(txt)

	if len(txt) == 0 {
		return nil, errors.New("The text must not be empty \n")
	}

	return &txt, nil
}

func do[R Reponse](method, url string, body, headers map[string]string) (*R, error) {
	jsonVal, err := json.Marshal(body)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewReader(jsonVal))

	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)

		return nil, errors.New(fmt.Sprintf("Request failed with status: %v \n", string(respBody)))
	}

	var respBodyTarget R

	err = json.NewDecoder(resp.Body).Decode(&respBodyTarget)

	if err != nil {
		return nil, err
	}

	return &respBodyTarget, nil
}
