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
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// - TYPES

type Reponse interface{}

type TwitchData struct {
	UserId             *string       `json:"user_id"`
	Username           *string       `json:"username"`
	ClientAppId        *string       `json:"client_app_id"`
	ClientAppSecret    *string       `json:"client_app_secret"`
	AccessToken        *string       `json:"access_token"`
	RefreshToken       *string       `json:"refresh_token"`
	ExpiresAtTimestamp *int64        `json:"expires_at_timestamp"`
	Clips              *[]TwitchClip `json:"clips"`

	mutex sync.Mutex
}

type TwitchClip struct {
	Id      string `json:"id"`
	EditUrl string `json:"edit_url"`
}

// - VARIABLES & CONSTANTS

const twitchDataFileName = "twitch-data.json"

var ticker = time.NewTicker(time.Second * 1)
var twitchData TwitchData

// - MAIN

func main() {
	defer ticker.Stop()
	for err := updateTwitchData(); err != nil; {
	}

	go handleTwitchAuth()

	err := updateTwitchUser()

	if err != nil {
		panic(err)
	}

	for {
		listenForTwitchClipCommand()
		twitchData.getLatestTwitchClip()
	}
}

// - TWITCH

func (twitchData *TwitchData) getLatestTwitchClip() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/clips?id=%v", (*twitchData.Clips)[len(*twitchData.Clips)-1].Id)
	headers := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + *twitchData.AccessToken, "Client-Id": *twitchData.ClientAppId}
	resp, err := do[map[string][]*any]("GET", url, nil, headers)
	fmt.Println(err)
	if err != nil {
		return err
	}

	clipUrl := (*(*resp)["data"][0]).(map[string]any)["url"].(string)
	fmt.Println(clipUrl)
	out, err := os.Create("clip.mp4")

	if err != nil {
		return err
	}

	defer out.Close()

	clipResp, err := http.Get(clipUrl)
	defer clipResp.Body.Close()

	_, err = io.Copy(out, clipResp.Body)

	return err
}

func listenForTwitchClipCommand() {
	txt, err := readStdin("create a new clip by writing 'i have small pp' in the terminal :> \n")

	if err != nil {
		fmt.Println("u r the proble, not the solution. try again")
	} else if strings.ToLower(*txt) != "i have small pp" {
		fmt.Println("the commaand u wrote is soooo wrong, i cannot even")
	}

	fmt.Println("creating your new little clip...")

	err = createTwitchClip()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("i super duper awesome cuz i just cliped your stream...u r welcome")
}

func createTwitchClip() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/clips?broadcaster_id=%v", *twitchData.UserId)
	headers := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + *twitchData.AccessToken, "Client-Id": *twitchData.ClientAppId}
	resp, err := do[map[string][]TwitchClip]("POST", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		if twitchData.Clips == nil {
			array := make([]TwitchClip, 0)
			twitchData.Clips = &array
		}
		*twitchData.Clips = append(*twitchData.Clips, (*resp)["data"][0])
		twitchData.save()
		return nil
	}

	return errors.New("Ok, i tried buddy. idk how to tell u this but the clipping faileed...miserably. but dont worry, you can try again #fingercrossed")
}

func updateTwitchUser() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", *twitchData.Username)
	headers := map[string]string{"Authorization": "Bearer " + *twitchData.AccessToken, "Client-Id": *twitchData.ClientAppId}
	resp, err := do[map[string][]map[string]any]("GET", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		userId := (*resp)["data"][0]["id"].(string)
		twitchData.UserId = &userId

		return twitchData.save()
	}

	return errors.New("Something went wrong, but u not the problem...maybe....OK, it s you")
}

func handleTwitchAuth() {
	for {
		<-ticker.C

		twitchData.mutex.Lock()
		defer twitchData.mutex.Unlock()
		for *twitchData.ExpiresAtTimestamp <= time.Now().Unix() {
			err := updateTwitchAccessToken()

			if err != nil {
				updateTwitchAuth()
			}
		}
	}
}

// - TWITCH AUTHORIZATION

func updateTwitchAccessToken() error {
	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%v&client_id=%v&client_secret=%v",
		*twitchData.RefreshToken, *twitchData.ClientAppId, *twitchData.ClientAppSecret,
	)

	resp, err := do[map[string]*any]("POST", url, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if err != nil {
		return nil
	}

	accessToken := (*(*resp)["access_token"]).(string)
	refreshToken := (*(*resp)["refresh_token"]).(string)
	expiresAtTimestamp := time.Now().Unix() + int64((*(*resp)["expires_in"]).(float64)) - 300

	twitchData.AccessToken = &accessToken
	twitchData.RefreshToken = &refreshToken
	twitchData.ExpiresAtTimestamp = &expiresAtTimestamp

	return twitchData.save()
}

func updateTwitchAuth() error {
	url := fmt.Sprintf("https://id.twitch.tv/oauth2/device?client_id=%v&scopes=clips:edit", *twitchData.ClientAppId)
	resp, err := do[map[string]*any]("POST", url, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if err != nil {
		return err
	}

	deviceCode := (*(*resp)["device_code"]).(string)
	userCode := (*(*resp)["user_code"]).(string)
	verificationUrl := (*(*resp)["verification_uri"]).(string)

	if !strings.Contains(verificationUrl, userCode) {
		return errors.New("Verification URL doesnt contain the right user code")
	}

	if err := openBrowser(verificationUrl); err != nil {
		return err
	}

	txt, err := readStdin("Type 'y' when u authorized the device: ")

	if err != nil {
		return err
	}

	if *txt != "y" {
		return errors.New("yoou have to press 'y'...dummyyyyy")
	}

	url = fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?client_id=%v&scope=clips:edit&device_code=%v&grant_type=urn:ietf:params:oauth:grant-type:device_code",
		*twitchData.ClientAppId, deviceCode,
	)
	resp, err = do[map[string]*any]("POST", url, nil, nil)

	if err != nil {
		return err
	}

	accessToken := (*(*resp)["access_token"]).(string)
	refreshToken := (*(*resp)["refresh_token"]).(string)
	expiresAtTimestamp := time.Now().Unix() + int64((*(*resp)["expires_in"]).(float64)) - 300

	twitchData.AccessToken = &accessToken
	twitchData.RefreshToken = &refreshToken
	twitchData.ExpiresAtTimestamp = &expiresAtTimestamp

	return twitchData.save()
}

// - TWITCH STORED DATA

func updateTwitchData() error {
	data, _ := os.ReadFile(twitchDataFileName)

	err := json.Unmarshal(data, &twitchData)

	if err == nil && twitchData.Username != nil && twitchData.ClientAppId != nil && twitchData.ClientAppSecret != nil {
		command, _ := readStdin("I found your twitch username, client app id & secret, do you want to update them? if yes press 'y': ")

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

	twitchClientAppSecret, err := getTwitchClientAppSecret()

	if err != nil {
		return nil
	}

	twitchData.Username = twitchUsername
	twitchData.ClientAppId = twitchClientAppId
	twitchData.ClientAppSecret = twitchClientAppSecret

	return twitchData.save()
}

func getTwitchClientAppSecret() (*string, error) {
	twitchClientAppSecret, err := readStdin("Now...the twitch client app secret...hurrrryyyyy: ")

	if err != nil {
		return nil, err
	}

	return twitchClientAppSecret, nil
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

func (twitchData *TwitchData) save() error {
	return writeToFile(twitchDataFileName, twitchData)
}

func writeToFile(fileName string, data any) error {
	json, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return os.WriteFile(fileName, json, 0644)
}

func openBrowser(url string) error {
	return exec.Command("cmd", []string{"/c", "start", url}...).Start()
}
