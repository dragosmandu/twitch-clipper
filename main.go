package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// - TYPES

type Reponse interface{}

type TwitchUserProvidedData struct {
	username        string
	clientAppId     string
	clientAppSecret string
}

type TwitchAuth struct {
	AccessToken string `json:"access_token"`
}

type TwitchAuth2 struct {
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

var twitchUserProvidedData TwitchUserProvidedData = TwitchUserProvidedData{username: "cptazazel", clientAppId: "", clientAppSecret: ""}
var twitchAuth TwitchAuth
var twitchUser TwitchUser = TwitchUser{Id: "770869829"}
var twitchClip TwitchClip

// - MAIN

func main() {
	// err := updateTwitchUserProvidedData()

	// if err != nil {
	// 	panic(err)
	// }

	err := updateTwitchAuth2()

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
	headers := map[string]string{"Authorization": "Bearer " + twitchAuth.AccessToken, "Client-Id": twitchUserProvidedData.clientAppId}
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

func updateTwitchAuth2() error { // this should work to get the right token
	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/device?client_id=%v&scopes='clips:edit'",
		twitchUserProvidedData.clientAppId,
	)
	resp, err := do[TwitchAuth2]("POST", url, nil, nil)

	if err != nil {
		return err
	}

	fmt.Println(resp)

	return nil
}

func updateTwitchUser() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", twitchUserProvidedData.username)
	headers := map[string]string{"Authorization": "Bearer " + twitchAuth.AccessToken, "Client-Id": twitchUserProvidedData.clientAppId}
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
	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?client_id=%v&client_secret=%v&grant_type=client_credentials",
		twitchUserProvidedData.clientAppId,
		twitchUserProvidedData.clientAppSecret,
	)
	resp, err := do[TwitchAuth]("POST", url, nil, nil)

	if err != nil {
		return err
	}

	twitchAuth = *resp

	return nil
}

// - TWITCH USER PROVIDED DATA

func updateTwitchUserProvidedData() error {
	errA, errB, errC := updateTwitchUsername(), updateTwitchClientAppId(), updateTwitchClientAppSecret()

	if errA != nil {
		return errA
	}

	if errB != nil {
		return errB
	}

	return errC
}

func updateTwitchClientAppSecret() error {
	twitchClientAppSecret, err := readStdin("Pleeeeeease...kindly tell me your Twitch client app secret: ")

	if err != nil {
		return err
	}

	twitchUserProvidedData.clientAppSecret = *twitchClientAppSecret

	return nil
}

func updateTwitchClientAppId() error {
	twitchClientAppId, err := readStdin("Please tell me your Twitch client app id, or else...: ")

	if err != nil {
		return err
	}

	twitchUserProvidedData.clientAppId = *twitchClientAppId

	return nil
}

func updateTwitchUsername() error {
	twitchUsername, err := readStdin("Please tell me your Twitch username: ")

	if err != nil {
		return err
	}

	isTwitchUsernameValid := regexp.MustCompile(`^[a-zA-Z0-9\_]+$`).MatchString(*twitchUsername)

	if !isTwitchUsernameValid {
		return errors.New("Dummy, your twitch username was invalid :) \n")
	}

	twitchUserProvidedData.username = *twitchUsername

	return nil
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
	} else if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New(fmt.Sprintf("Request failed with status: %v \n", resp.Status))
	}

	defer resp.Body.Close()
	var respBodyTarget R

	err = json.NewDecoder(resp.Body).Decode(&respBodyTarget)

	if err != nil {
		return nil, err
	}

	return &respBodyTarget, nil
}
