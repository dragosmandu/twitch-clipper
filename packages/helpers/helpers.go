package helpers

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
	"strings"
)

type Reponse interface{}

func ReadStdin(msg string) (*string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(msg)

	txt, err := reader.ReadString('\n')

	if err != nil {
		return nil, err
	}

	txt = strings.TrimSpace(txt)

	if len(txt) == 0 {
		return nil, errors.New("the text must not be empty")
	}

	return &txt, nil
}

func Do[R Reponse](method, url string, body, headers map[string]string) (*R, error) {
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

		return nil, fmt.Errorf("request failed with status: %v", string(respBody))
	}

	var respBodyTarget R

	err = json.NewDecoder(resp.Body).Decode(&respBodyTarget)

	if err != nil {
		return nil, err
	}

	return &respBodyTarget, nil
}

func WriteToFile(fileName string, data any) error {
	json, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return os.WriteFile(fileName, json, 0644)
}

func ReadFromFile(filename string, to interface{}) error {
	b, err := os.ReadFile(filename)

	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &to)

	return err
}

func AppendToFile(fileName string, data string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return nil
	}

	_, err = f.WriteString(data)

	return err
}

func OpenBrowser(url string) error {
	return exec.Command("cmd", []string{"/c", "start", url}...).Start()
}

// ffmpeg -i twitch-clip.mp4 -vf "crop=515:340:1700:340,scale=1080:1920" output.mp4 (crop camera for mid right)
// ffmpeg -i twitch-clip.mp4 -lavfi "[0:v]scale=1920*2:1080*2,boxblur=luma_radius=min(h\,w)/20:luma_power=1:chroma_radius=min(cw\,ch)/20:chroma_power=1[bg];[0:v]scale=-1:1080[ov];[bg][ov]overlay=(W-w)/2:(H-h)/2,crop=w=1080:h=1920" output.mp4
// ffmpeg -i twitch-clip.mp4 -vf "crop=1080:366:0:0" output1.mp4
// ffmpeg -i output.mp4 -i output1.mp4 -filter_complex vstack=inputs=2 output2.mp4 (combine 2 videos)
func InstallPrereq() error {
	fmt.Println("i need FFMPEG tool in order to edit the video. write 'winget install ffmpeg' in your terminal")

	txt, err := ReadStdin("press 'y' when u installed it")

	if err != nil {
		return err
	}

	if *txt != "y" {
		return errors.New("ffmpeg is required")
	}

	return nil
}
