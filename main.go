package main

import (
	"github.com/dragosmandu/twitch-clipper/packages/twitch"
)

// - MAIN

func main() {
	twitch.ConfigureTwitch()
}

/*
 	ffmpeg -i twitch-clip.mp4 -vf "crop=515:340:1700:340,scale=1080:-1" output.mp4
	ffmpeg -i twitch-clip.mp4 -vf "crop=1080:1080:0:0,scale=1080:1206" output1.mp4
	ffmpeg -i output.mp4 -i output1.mp4 -filter_complex vstack=inputs=2 output2.mp4
	ffmpeg -i twitch-clip.mp4 -lavfi "[0:v]scale=1920*2:1080*2,boxblur=luma_radius=min(h\,w)/20:luma_power=1:chroma_radius=min(cw\,ch)/20:chroma_power=1[bg];[0:v]scale=-1:1080[ov];[bg][ov]overlay=(W-w)/2:(H-h)/2,crop=w=1080:h=1920" output.mp4

	_____________________
	|	    vid1        |
	|					|    crop camera +/- whatever u want -> scale it to 1080:-1
	|					|
	|	__________		|
	|					|
	|					|
	|					|
	|					|
	|					| 	crop area of interest x:y -> scale it to 1080:1920-vid1.height
	|		vid2		|
	|					|
	|					|
	|					|
	|					|
	_____________________

	-> vstack  vid1 vid2
*/
