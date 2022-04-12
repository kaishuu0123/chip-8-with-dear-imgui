package main

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/audio"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/chip8"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/gui"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/master_data"
)

const (
	MASTER_WINDOW_WIDTH  = 800
	MASTER_WINDOW_HEIGHT = 600
	DISPLAY_WIDTH        = 640
	DISPLAY_HEIGHT       = 320
)

func runChip8() {
	var foreGroundColor = color.RGBA{R: 156, G: 220, B: 254, A: 255}
	var backGroundColor = color.RGBA{R: 50, G: 50, B: 54, A: 255}

	var masterData = master_data.GetMasterDataInstance()
	masterData.InitAllTickers()

	// alias
	var display = masterData.DisplayRGBA

	masterData.RunningChip8 = true
	for masterData.RunningChip8 {
		select {
		case <-masterData.VideoTicker.C:
			for y := 0; y < int(display.Bounds().Dy()); y++ {
				for x := 0; x < int(display.Bounds().Dx()); x++ {
					videoX := x / 10
					videoY := y / 10

					if masterData.Chip8vm.Video[videoY][videoX] == 0 {
						display.Set(x, y, backGroundColor)
					} else {
						display.Set(x, y, foreGroundColor)
					}
				}
			}
		case <-masterData.ClockTicker.C:
			masterData.Chip8vm.Step()
		case <-masterData.DelayTicker.C:
			if masterData.Chip8vm.DT > 0 {
				masterData.Chip8vm.DT--
			}
		case <-masterData.SoundTicker.C:
			if masterData.Chip8vm.ST > 0 {
				masterData.Audio.OutSineWave()
				masterData.Chip8vm.ST--
			}
		}
	}
}

func resetWhenOnDrop(file_name string) {
	var md = master_data.GetMasterDataInstance()
	md.AddLogMessage(fmt.Sprintf("Loading ROM ... (PATH: %s)", file_name))

	md.StopAllTickers()
	md.RunningChip8 = false
	md.Chip8vm, _ = chip8.LoadFromFile(file_name)
	md.AddLogMessage("Loading ROM completed.")
	md.ResetAllTickers()
	md.RunningChip8 = true
}

func onDrop(names []string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s", names[0]))
	dropInFiles := sb.String()
	resetWhenOnDrop(dropInFiles)
}

func processKeyboardEvents() {
	masterData := master_data.GetMasterDataInstance()
	for k, v := range master_data.KeyMap {
		if imgui.IsKeyPressed(int(k)) {
			masterData.Chip8vm.PressKey(v)
		}
		if imgui.IsKeyReleased(int(k)) {
			masterData.Chip8vm.ReleasedKey(v)
		}
	}
}

func main() {
	masterData := master_data.GetMasterDataInstance()

	masterData.DisplayRGBA = image.NewRGBA(image.Rect(0, 0, DISPLAY_WIDTH, DISPLAY_HEIGHT))

	masterData.Window = gui.NewMasterWindow("CHIP-8 with Dear ImGUI", MASTER_WINDOW_WIDTH, MASTER_WINDOW_HEIGHT, 0)
	var window *gui.MasterWindow = masterData.Window
	window.SetDropCallback(onDrop)

	vm, _ := chip8.LoadROM(chip8.Boot, false)
	masterData.Chip8vm = vm

	masterData.Audio, _ = audio.NewAudio()
	err := masterData.Audio.Start()
	if err != nil {
		panic(err)
	}
	defer masterData.Audio.Close()

	masterData.AddLogMessage("CHIP-8 with Dear ImGUI initialized!")
	masterData.AddLogMessage("Please drag and drop CHIP-8's ROM (Binary data ONLY)")

	go runChip8()

	for !window.Platform.ShouldStop() {
		window.Platform.ProcessEvents()
		processKeyboardEvents()

		texture, _ := window.Renderer.CreateImageTexture(masterData.DisplayRGBA)

		renderGUI(window, &texture)
	}

	// ToDO: Signal Handling (use NotifyContext?)
}
