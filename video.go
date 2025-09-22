package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"slices"
)

func getVideoAspectRatio(filePath string) (string, error) {
	type cmdOut struct {
		Streams []struct {
			Width              int     `json:"width"`
			Height             int     `json:"height"`
			DisplayAspectRatio *string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffprobe error: %v", err)
	}
	params := cmdOut{}
	err = json.Unmarshal(out.Bytes(), &params)
	if err != nil {
		return "", fmt.Errorf("could not parse ffprobe output: %v", err)
	}
	stdOut := params.Streams[0]
	if stdOut.DisplayAspectRatio == nil {
		return getAspectRatio(stdOut.Width, stdOut.Height), nil
	}
	return *stdOut.DisplayAspectRatio, nil
}

func getAspectRatio(width, height int) string {
	widthFactor := getFactor(width)
	heightFactor := getFactor(height)

	var gcf int
	if len(widthFactor) < len(heightFactor) {
		gcf = getGCF(widthFactor, heightFactor)
	} else {
		gcf = getGCF(heightFactor, widthFactor)
	}

	aspectWith := width / gcf
	aspectHeight := height / gcf
	return fmt.Sprintf("%v:%v", aspectWith, aspectHeight)
}

func getGCF(i, j []int) int {
	biggestNum := math.Inf(-1)
	for _, num := range i {
		if ok := slices.Contains(j, num); ok {
			if float64(num) > biggestNum {
				biggestNum = float64(num)
			}
		}
	}
	return int(biggestNum)
}

func getFactor(num int) []int {
	sqrt := math.Sqrt(float64(num))
	roundSqrt := math.Round(sqrt)
	var factors []int
	for i := 1; i <= int(roundSqrt); i++ {
		if factor := num % i; factor == 0 {
			factors = append(factors, i)
			factors = append(factors, num/i)
		}
	}
	return factors
}

func processVideoForFastStart(filePath string) (string, error) {
	outputFilePath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputFilePath)
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ffmpeg error: %w", err)
	}
	return outputFilePath, nil
}
