package tokens

import (
	"fmt"
	"strings"
)

// CSS custom property names (design token identifiers).
const (
	CSSColorCanvas    = "--color-canvas"
	CSSColorSurface   = "--color-surface"
	CSSColorBorder    = "--color-border"
	CSSColorText      = "--color-text"
	CSSColorTextMuted = "--color-text-muted"
	CSSColorAccent    = "--color-accent"
	CSSColorDanger    = "--color-danger"
	CSSColorSuccess   = "--color-success"
	CSSColorOnAccent  = "--color-on-accent"
	CSSColorCanvasHi  = "--color-canvas-hi"
	CSSColorCanvasLo  = "--color-canvas-lo"

	CSSSpaceXS = "--space-xs"
	CSSSpaceSM = "--space-sm"
	CSSSpaceMD = "--space-md"
	CSSSpaceLG = "--space-lg"
	CSSSpaceXL = "--space-xl"

	CSSFontSans    = "--font-sans"
	CSSFontMono    = "--font-mono"
	CSSFontDisplay = "--font-display"

	CSSRadiusSM = "--radius-sm"
	CSSRadiusMD = "--radius-md"
)

// DesignPairs is the ordered name→value list for CSS generation and contracts.
func DesignPairs() [][2]string {
	return [][2]string{
		{CSSColorCanvas, ColorCanvas},
		{CSSColorSurface, ColorSurface},
		{CSSColorBorder, ColorBorder},
		{CSSColorText, ColorText},
		{CSSColorTextMuted, ColorTextMuted},
		{CSSColorAccent, ColorAccent},
		{CSSColorDanger, ColorDanger},
		{CSSColorSuccess, ColorSuccess},
		{CSSColorOnAccent, ColorOnAccent},
		{CSSColorCanvasHi, ColorCanvasHi},
		{CSSColorCanvasLo, ColorCanvasLo},
		{CSSSpaceXS, SpaceXS},
		{CSSSpaceSM, SpaceSM},
		{CSSSpaceMD, SpaceMD},
		{CSSSpaceLG, SpaceLG},
		{CSSSpaceXL, SpaceXL},
		{CSSFontSans, FontSans},
		{CSSFontMono, FontMono},
		{CSSFontDisplay, FontDisplay},
		{CSSRadiusSM, RadiusSM},
		{CSSRadiusMD, RadiusMD},
	}
}

// RenderDesignCSS returns the canonical web/tokens.css body.
func RenderDesignCSS() string {
	var b strings.Builder
	b.WriteString(":root {\n")
	for _, p := range DesignPairs() {
		fmt.Fprintf(&b, "  %s: %s;\n", p[0], p[1])
	}
	b.WriteString("}\n")
	return b.String()
}
