package main

import (
	"fmt"
	"image/color"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/gui"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/master_data"
)

var (
	windowFlags imgui.WindowFlags = imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoResize

	millisPerSecond float32 = 1000.0
)

func drawKeyPad(fontSize *imgui.Vec2) {
	{
		drawList := imgui.WindowDrawList()
		cursorPos := imgui.CursorScreenPos()

		pos := cursorPos
		textColor := imgui.CurrentStyle().Color(imgui.StyleColorText)
		buttonColor := imgui.CurrentStyle().Color(imgui.StyleColorButton)
		buttonActiveColor := imgui.CurrentStyle().Color(imgui.StyleColorButtonActive)

		for index, key := range master_data.KeyMapOrder {
			var rectColor color.RGBA
			if imgui.IsKeyPressed(int(key)) {
				rectColor = color.RGBA{
					R: uint8(buttonActiveColor.X * 255),
					G: uint8(buttonActiveColor.Y * 255),
					B: uint8(buttonActiveColor.Z * 255),
					A: uint8(buttonActiveColor.W * 255),
				}
			} else {
				rectColor = color.RGBA{
					R: uint8(buttonColor.X * 255),
					G: uint8(buttonColor.Y * 255),
					B: uint8(buttonColor.Z * 255),
					A: uint8(buttonColor.W * 255),
				}
			}

			if index%4 == 0 {
				pos.X = cursorPos.X
				if index/4 > 0 {
					pos = pos.Plus(imgui.Vec2{X: 0, Y: (fontSize.Y * 2)})
				}
			} else {
				pos = pos.Plus(imgui.Vec2{X: (fontSize.Y * 2), Y: 0})
			}

			drawList.AddRectFilled(
				pos,
				pos.Plus(
					imgui.Vec2{
						X: fontSize.Y * 1.5,
						Y: fontSize.Y * 1.5,
					}),
				imgui.Packed(rectColor),
			)
			drawList.AddText(
				pos.Plus(
					imgui.Vec2{
						X: (fontSize.Y * 1.5 / 2) - (fontSize.X / 2),
						Y: (fontSize.Y * 1.5 / 2) - (fontSize.Y / 2),
					}),
				imgui.Packed(color.RGBA{
					R: uint8(textColor.X * 255),
					G: uint8(textColor.Y * 255),
					B: uint8(textColor.Z * 255),
					A: uint8(textColor.W * 255),
				}),
				string(key),
			)
		}
	}
}

func renderGUI(w *gui.MasterWindow, texture *imgui.TextureID) {
	var masterData = master_data.GetMasterDataInstance()

	w.Platform.NewFrame()
	imgui.NewFrame()

	// DON'T FORGET call PopStyleVar when PushStyleVar called
	imgui.PushStyleVarFloat(imgui.StyleVarWindowRounding, 0.0)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 0, Y: 0})

	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: 0})

	imgui.BeginV("Display", nil, windowFlags|imgui.WindowFlagsAlwaysAutoResize)
	imgui.Image(*texture, imgui.Vec2{X: master_data.DISPLAY_WIDTH, Y: master_data.DISPLAY_HEIGHT})
	displaySize := imgui.WindowSize()
	imgui.End()

	// Pop StyleVarWindowPadding
	imgui.PopStyleVar()

	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: displaySize.Y})
	imgui.SetNextWindowSize(imgui.Vec2{X: displaySize.X, Y: 0})

	imgui.BeginV("Status & Controls", nil, windowFlags)
	if masterData.RunningChip8 {
		imgui.Text("Status: RUNNING")
	} else {
		imgui.Text("Status: STOP")
	}

	if imgui.Button("RESET") {
		masterData.ResetVM()
	}
	imgui.SameLine()
	if masterData.RunningChip8 {
		if imgui.Button("STOP") {
			masterData.StopVM()
		}
	} else {
		if imgui.Button("START") {
			masterData.StartVM()
		}
	}
	statusControlSize := imgui.WindowSize()
	imgui.End()

	imgui.SetNextWindowPos(imgui.Vec2{X: 0, Y: displaySize.Y + statusControlSize.Y})
	imgui.SetNextWindowSize(imgui.Vec2{X: displaySize.X, Y: master_data.MASTER_WINDOW_HEIGHT - displaySize.Y - statusControlSize.Y})
	imgui.BeginV("Message", nil, windowFlags)
	for _, message := range masterData.LogMessages {
		imgui.Text(message)
	}
	imgui.End()

	imgui.SetNextWindowPos(imgui.Vec2{X: displaySize.X, Y: 0})
	imgui.SetNextWindowSize(imgui.Vec2{X: master_data.MASTER_WINDOW_WIDTH - displaySize.X, Y: displaySize.Y})

	imgui.BeginV("Internal of CHIP-8", nil, windowFlags)
	imgui.PushFont(masterData.Window.FontsData[1])
	imgui.BeginTable("V[X] & Stack table", 2)
	imgui.TableSetupColumn("V[x]")
	imgui.TableSetupColumn("Stack")
	imgui.TableHeadersRow()

	for rowIndex := 0; rowIndex < len(masterData.Chip8vm.V); rowIndex++ {
		imgui.TableNextRow()

		imgui.TableSetColumnIndex(0)
		imgui.Text(fmt.Sprintf("V%X: %02X", rowIndex, masterData.Chip8vm.V[rowIndex]))
		imgui.TableSetColumnIndex(1)
		imgui.Text(fmt.Sprintf("S%X: %04X", rowIndex, masterData.Chip8vm.Stack[rowIndex]))
	}
	imgui.EndTable()
	imgui.Separator()
	imgui.BeginTable("MISC registers table", 2)
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(fmt.Sprintf("PC: %04X", masterData.Chip8vm.PC))
	imgui.TableSetColumnIndex(1)
	imgui.Text(fmt.Sprintf("DT: %02X", masterData.Chip8vm.DT))
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(fmt.Sprintf("SP: %02X", masterData.Chip8vm.SP))
	imgui.TableSetColumnIndex(1)
	imgui.Text(fmt.Sprintf("ST: %02X", masterData.Chip8vm.ST))
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(fmt.Sprintf(" I: %04X", masterData.Chip8vm.I))
	imgui.EndTable()
	imgui.PopFont()
	imgui.End()

	fontSize := imgui.CalcTextSize("A", false, 0.0)
	imgui.SetNextWindowPos(imgui.Vec2{X: displaySize.X, Y: displaySize.Y})
	imgui.SetNextWindowSize(imgui.Vec2{X: master_data.MASTER_WINDOW_WIDTH - displaySize.X, Y: fontSize.Y * 10})
	imgui.BeginV("KeyPad", nil, windowFlags)
	// draw KeyPad
	drawKeyPad(&fontSize)
	keyPadWindowSize := imgui.WindowSize()
	imgui.End()

	imgui.SetNextWindowPos(imgui.Vec2{X: displaySize.X, Y: displaySize.Y + keyPadWindowSize.Y})
	imgui.SetNextWindowSize(imgui.Vec2{X: master_data.MASTER_WINDOW_WIDTH - displaySize.X, Y: master_data.MASTER_WINDOW_HEIGHT - displaySize.Y - keyPadWindowSize.Y})
	imgui.BeginV("Debug", nil, windowFlags)
	imgui.Text("[PERF]")
	imgui.Text(fmt.Sprintf("%.3f ms/frame",
		millisPerSecond/imgui.CurrentIO().Framerate()))
	imgui.Text("[FPS]")
	imgui.Text(fmt.Sprintf("%.1f fps",
		imgui.CurrentIO().Framerate()))
	imgui.End()

	// Pop StyleVarWindowRounding
	imgui.PopStyleVar()

	imgui.Render()

	w.Renderer.PreRender(w.ClearColor)
	w.Renderer.Render(w.Platform.DisplaySize(), w.Platform.FramebufferSize(), imgui.RenderedDrawData())
	w.Platform.PostRender()
}
