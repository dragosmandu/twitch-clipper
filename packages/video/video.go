package video

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type cropArgs struct {
	width  int
	height int
	x      int
	y      int
}

type scaleArgs struct {
	width  int
	height int
}

var upperVideoPartCropArgs = cropArgs{
	width:  515,
	height: 340,
	x:      1700,
	y:      340,
}
var upperVideoPartScaleArgs = scaleArgs{
	width:  1080,
	height: -1,
}

var lowerVideoPartCropArgs = cropArgs{
	width:  1080,
	height: 1080,
	x:      0,
	y:      0,
}

var lowerVideoPartScaleArgs = scaleArgs{
	width:  1080,
	height: 1206,
}

const upperVideoPartFileName = "./resources/upper-video-part.mp4"
const lowerVideoPartFileName = "./resources/lower-video-part.mp4"
const portraitVideoFileName = "./resources/portrait-video.mp4"
const twitchClipFileName = "./resources/twitch-clip.mp4"

func CreatePortraitVideo() error {
	defer cleanOurStuff()

	if _, err := os.ReadFile(twitchClipFileName); err != nil {
		return errors.New("i couldnt find the downloaded clip...u duuuuummmyyyyyy")
	}

	fmt.Println("creating the portrait video...")

	err := createUpperVideoPart()

	if err != nil {
		return err
	}

	err = createLowerVideoPart()

	if err != nil {
		return err
	}

	return combineParts()
}

func createUpperVideoPart() error {
	cmdArgs := []string{
		"-y",
		"-i",
		twitchClipFileName,
		"-vf",
		fmt.Sprintf("crop=%v:%v:%v:%v,scale=%v:%v", upperVideoPartCropArgs.width, upperVideoPartCropArgs.height, upperVideoPartCropArgs.x, upperVideoPartCropArgs.y, upperVideoPartScaleArgs.width, upperVideoPartScaleArgs.height),
		upperVideoPartFileName,
	}

	return exec.Command("ffmpeg", cmdArgs...).Run()
}

func createLowerVideoPart() error {
	cmdArgs := []string{
		"-y",
		"-i",
		twitchClipFileName,
		"-vf",
		fmt.Sprintf("crop=%v:%v:%v:%v,scale=%v:%v", lowerVideoPartCropArgs.width, lowerVideoPartCropArgs.height, lowerVideoPartCropArgs.x, lowerVideoPartCropArgs.y, lowerVideoPartScaleArgs.width, lowerVideoPartScaleArgs.height),
		lowerVideoPartFileName,
	}

	return exec.Command("ffmpeg", cmdArgs...).Run()
}

func combineParts() error {
	cmdArgs := []string{
		"-y",
		"-i",
		upperVideoPartFileName,
		"-i",
		lowerVideoPartFileName,
		"-filter_complex",
		"vstack=inputs=2",
		portraitVideoFileName,
	}

	return exec.Command("ffmpeg", cmdArgs...).Run()
}

func cleanOurStuff() {
	os.Remove(upperVideoPartFileName)
	os.Remove(lowerVideoPartFileName)
	os.Remove(twitchClipFileName)
}
