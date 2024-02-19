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
