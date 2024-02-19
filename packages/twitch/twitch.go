package twitch

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dragosmandu/twitch-clipper/packages/helpers"
)

type TwitchData struct {
	UserId             *string `json:"user_id"`
	Username           *string `json:"username"`
	ClientAppId        *string `json:"client_app_id"`
	ClientAppSecret    *string `json:"client_app_secret"`
	AccessToken        *string `json:"access_token"`
	RefreshToken       *string `json:"refresh_token"`
	ExpiresAtTimestamp *int64  `json:"expires_at_timestamp"`

	mutex sync.Mutex
}

type TwitchClip struct {
	Id      string `json:"id"`
	EditUrl string `json:"edit_url"`
}

const twitchDataFileName = "twitch-data.json"
const twitchClipsFileName = "twitch-clips.txt"
const twitchClipFileName = "twitch-clip.mp4"

var twitchData TwitchData

// - Twitch

func ConfigureTwitch() {
	for err := updateTwitchInputData(); err != nil; {
	}

	go func() {
		for err := handleTwitchAuth(); err != nil; {
			if txt, _ := helpers.ReadStdin("i failed to get twitch authorization, do you want to try again? press 'y' for yes: "); *txt != "y" {
				break
			}
		}
	}()

	for waitForTwitchClipCommand(); ; {
	}
}

// - Twitch API

func getTwitchClipDownloadUrl(clipId string) (*string, error) {
	url := fmt.Sprintf("https://api.twitch.tv/helix/clips?id=%v", clipId)
	headers := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + *twitchData.AccessToken, "Client-Id": *twitchData.ClientAppId}
	resp, err := helpers.Do[map[string]any]("GET", url, nil, headers)

	if err != nil {
		return nil, err
	}

	downloadUrl := strings.Split((*resp)["data"].([]interface{})[0].(map[string]any)["thumbnail_url"].(string), "-preview")[0] + ".mp4"

	return &downloadUrl, nil
}

func downloadTwitchClip(url *string) error {
	if url == nil {
		return errors.New("invalid empty url string")
	}

	out, err := os.Create(twitchClipFileName)

	if err != nil {
		return err
	}

	defer out.Close()

	clipResp, err := http.Get(*url)

	if err != nil {
		return err
	}

	defer clipResp.Body.Close()

	_, err = io.Copy(out, clipResp.Body)

	return err
}

func waitForTwitchClipCommand() {
	txt, err := helpers.ReadStdin("create a new clip by writing 'i have small pp' in the terminal :> \n")

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
	resp, err := helpers.Do[map[string][]TwitchClip]("POST", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		time.Sleep(time.Second * 15)
		downloadUrl, err := getTwitchClipDownloadUrl((*resp)["data"][0].Id)

		if err1 := downloadTwitchClip(downloadUrl); err1 != nil || err != nil {
			fmt.Printf("I failed to download and share the latest clip, but you may see it in the 'twitch-clips.txt' file... hehe %v %v\n", err, err1)
		}

		return helpers.AppendToFile(twitchClipsFileName, fmt.Sprintf("Captured on %v: %v \n", time.Now().Format(time.RFC3339), (*resp)["data"][0].EditUrl))
	}

	return errors.New("ok, i tried buddy. idk how to tell u this but the clipping faileed...miserably. but dont worry, you can try again #fingercrossed")
}

func updateTwitchUser() error {
	url := fmt.Sprintf("https://api.twitch.tv/helix/users?login=%v", *twitchData.Username)
	headers := map[string]string{"Authorization": "Bearer " + *twitchData.AccessToken, "Client-Id": *twitchData.ClientAppId}
	resp, err := helpers.Do[map[string][]map[string]any]("GET", url, nil, headers)

	if err != nil {
		return err
	}

	if (*resp)["data"] != nil && len((*resp)["data"]) > 0 {
		userId := (*resp)["data"][0]["id"].(string)
		twitchData.UserId = &userId

		return twitchData.Save()
	}

	return errors.New("something went wrong, but u not the problem...maybe....OK, it s you")
}

// - Twitch Auth

func handleTwitchAuth() error {
	var ticker = time.NewTicker(time.Second * 1)
	const delay time.Duration = 20
	defer ticker.Stop()

	for {
		<-ticker.C

		twitchData.mutex.Lock()
		for i := 0; *twitchData.ExpiresAtTimestamp <= time.Now().Unix() && i < 10; i++ {
			err := updateTwitchAccessToken()

			if err != nil {
				updateTwitchAuth()
			}

			time.Sleep(time.Second * delay)
		}

		for i := 0; twitchData.UserId == nil && i < 10; i++ {
			updateTwitchUser()
			time.Sleep(time.Second * delay)
		}

		if twitchData.UserId == nil || twitchData.AccessToken == nil || *twitchData.ExpiresAtTimestamp <= time.Now().Unix() {
			break
		}
		twitchData.mutex.Unlock()
	}

	return errors.New("failed to update twitch authorization")
}

func updateTwitchAccessToken() error {
	url := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%v&client_id=%v&client_secret=%v",
		*twitchData.RefreshToken, *twitchData.ClientAppId, *twitchData.ClientAppSecret,
	)

	resp, err := helpers.Do[map[string]*any]("POST", url, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if err != nil {
		return nil
	}

	accessToken := (*(*resp)["access_token"]).(string)
	refreshToken := (*(*resp)["refresh_token"]).(string)
	expiresAtTimestamp := time.Now().Unix() + int64((*(*resp)["expires_in"]).(float64)) - 300

	twitchData.AccessToken = &accessToken
	twitchData.RefreshToken = &refreshToken
	twitchData.ExpiresAtTimestamp = &expiresAtTimestamp

	return twitchData.Save()
}

func updateTwitchAuth() error {
	url := fmt.Sprintf("https://id.twitch.tv/oauth2/device?client_id=%v&scopes=clips:edit", *twitchData.ClientAppId)
	resp, err := helpers.Do[map[string]*any]("POST", url, nil, map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if err != nil {
		return err
	}

	deviceCode := (*(*resp)["device_code"]).(string)
	userCode := (*(*resp)["user_code"]).(string)
	verificationUrl := (*(*resp)["verification_uri"]).(string)

	if !strings.Contains(verificationUrl, userCode) {
		return errors.New("verification URL doesnt contain the right user code")
	}

	if err := helpers.OpenBrowser(verificationUrl); err != nil {
		return err
	}

	txt, err := helpers.ReadStdin("Type 'y' when u authorized the device: ")

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
	resp, err = helpers.Do[map[string]*any]("POST", url, nil, nil)

	if err != nil {
		return err
	}

	accessToken := (*(*resp)["access_token"]).(string)
	refreshToken := (*(*resp)["refresh_token"]).(string)
	expiresAtTimestamp := time.Now().Unix() + int64((*(*resp)["expires_in"]).(float64)) - 300

	twitchData.AccessToken = &accessToken
	twitchData.RefreshToken = &refreshToken
	twitchData.ExpiresAtTimestamp = &expiresAtTimestamp

	return twitchData.Save()
}

// - Twitch input data

func (twitchData *TwitchData) Save() error {
	return helpers.WriteToFile(twitchDataFileName, twitchData)
}

func updateTwitchInputData() error {
	err := helpers.ReadFromFile(twitchDataFileName, &twitchData)

	if err == nil && twitchData.Username != nil && twitchData.ClientAppId != nil && twitchData.ClientAppSecret != nil {
		command, _ := helpers.ReadStdin("I found your twitch username, client app id & secret, do you want to update them? if yes press 'y': ")

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

	return twitchData.Save()
}

func getTwitchClientAppSecret() (*string, error) {
	twitchClientAppSecret, err := helpers.ReadStdin("Now...the twitch client app secret...hurrrryyyyy: ")

	if err != nil {
		return nil, err
	}

	return twitchClientAppSecret, nil
}

func getTwitchClientAppId() (*string, error) {
	twitchClientAppId, err := helpers.ReadStdin("Please tell me your Twitch client app id, or else...: ")

	if err != nil {
		return nil, err
	}

	return twitchClientAppId, nil
}

func getTwitchUsername() (*string, error) {
	twitchUsername, err := helpers.ReadStdin("Please tell me your Twitch username: ")

	if err != nil {
		return nil, err
	}

	isTwitchUsernameValid := regexp.MustCompile(`^[a-zA-Z0-9\_]+$`).MatchString(*twitchUsername)

	if !isTwitchUsernameValid {
		return nil, errors.New("dummy, your twitch username was invalid :)")
	}

	return twitchUsername, nil
}
