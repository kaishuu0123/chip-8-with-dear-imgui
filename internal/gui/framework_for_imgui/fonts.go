package framework_for_imgui

import (
	_ "embed"

	"github.com/inkyblackness/imgui-go/v4"
)

var (
	//go:embed fonts/HackGen35Nerd-Regular.ttf
	hackGen []byte
)

func SetupFont(io imgui.IO) []imgui.Font {
	fonts := io.Fonts()

	// fonts.AddFontDefault()
	var fontsData []imgui.Font
	fontsData = append(fontsData, fonts.AddFontFromMemoryTTFV(hackGen, 14.0, imgui.DefaultFontConfig, fonts.GlyphRangesJapanese()))
	fontsData = append(fontsData, fonts.AddFontFromMemoryTTFV(hackGen, 10.0, imgui.DefaultFontConfig, fonts.GlyphRangesJapanese()))

	return fontsData
}
