package master_data

import (
	"image"
	"time"

	"github.com/kaishuu0123/chip-8-dear-imgui/internal/audio"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/chip8"
	"github.com/kaishuu0123/chip-8-dear-imgui/internal/gui"
)

type MasterData struct {
	Chip8vm      *chip8.VirtualMachine
	RunningChip8 bool

	Window *gui.MasterWindow

	DisplayRGBA *image.RGBA
	Audio       *audio.Audio

	LogMessages []string

	ClockTicker *time.Ticker
	VideoTicker *time.Ticker
	DelayTicker *time.Ticker
	SoundTicker *time.Ticker
}

var (
	CLOCK_HZ = time.Second / 500
	VIDEO_HZ = time.Second / 60
	DELAY_HZ = time.Second / 60
	SOUND_HZ = time.Second / 60
)

const (
	MASTER_WINDOW_WIDTH  = 800
	MASTER_WINDOW_HEIGHT = 600

	DISPLAY_WIDTH  = 640
	DISPLAY_HEIGHT = 320
)

var masterDataInstance *MasterData = newMasterData()

func newMasterData() *MasterData {
	return &MasterData{}
}

func GetMasterDataInstance() *MasterData {
	return masterDataInstance
}

func (m *MasterData) AddLogMessage(message string) {
	m.LogMessages = append(m.LogMessages, message)
}

func (m *MasterData) InitAllTickers() {
	m.ClockTicker = time.NewTicker(CLOCK_HZ)
	m.VideoTicker = time.NewTicker(VIDEO_HZ)
	m.DelayTicker = time.NewTicker(DELAY_HZ)
	m.SoundTicker = time.NewTicker(SOUND_HZ)
}

func (m *MasterData) StopAllTickers() {
	m.ClockTicker.Stop()
	m.VideoTicker.Stop()
	m.DelayTicker.Stop()
	m.SoundTicker.Stop()
}

func (m *MasterData) ResetAllTickers() {
	m.ClockTicker.Reset(CLOCK_HZ)
	m.VideoTicker.Reset(VIDEO_HZ)
	m.DelayTicker.Reset(DELAY_HZ)
	m.SoundTicker.Reset(SOUND_HZ)
}

func (m *MasterData) ResetVM() {
	m.StopAllTickers()
	m.RunningChip8 = false
	var currentROM []byte
	currentROM = m.Chip8vm.ROM[chip8.BASE:]
	m.Chip8vm, _ = chip8.LoadROM(currentROM, false)
	m.ResetAllTickers()
	m.RunningChip8 = true
	m.AddLogMessage("Reset VM completed.")
}

func (m *MasterData) StopVM() {
	m.StopAllTickers()
	m.RunningChip8 = false
	m.AddLogMessage("VM stopped.")
}

func (m *MasterData) StartVM() {
	m.ResetAllTickers()
	m.RunningChip8 = true
	m.AddLogMessage("VM started.")
}

var KeyMap = map[rune]uint{
	'1': 0x1,
	'2': 0x2,
	'3': 0x3,
	'4': 0xC,
	'Q': 0x4,
	'W': 0x5,
	'E': 0x6,
	'R': 0xD,
	'A': 0x7,
	'S': 0x8,
	'D': 0x9,
	'F': 0xE,
	'Z': 0xA,
	'X': 0x0,
	'C': 0xB,
	'V': 0xF,
}

var KeyMapOrder = []rune{
	'1', '2', '3', '4',
	'Q', 'W', 'E', 'R',
	'A', 'S', 'D', 'F',
	'Z', 'X', 'C', 'V',
}
