package tokens

import (
	"fmt"
	"html"
	"strings"
)

// Landing DOM ids.
const (
	LandingDOMMagicForm  = "landing-magic-form"
	LandingDOMEmailInput = "landing-email"
	LandingDOMAuthHint   = "landing-hint"
)

// Landing UI copy (RU).
const (
	LandingCopyLede         = "Поиск похожих объявлений на Avito. Вход по Magic Link — без паролей."
	LandingCopyEmailLabel   = "Email"
	LandingCopySendLink     = "Прислать ссылку"
	LandingCopySentHint     = "Проверьте почту — ссылка ведёт на " + ProductDomainApp + "."
	LandingCopyRequestFailed = "Ошибка запроса Magic Link"
)

// RenderLandingHTML returns the marketing landing (magic link request only).
func RenderLandingHTML() string {
	var b strings.Builder
	esc := html.EscapeString
	stem := esc(ProductBrandStem())
	accent := esc(ProductBrandAccent())
	title := esc(ProductNameLocal)

	b.WriteString("<!DOCTYPE html>\n")
	fmt.Fprintf(&b, "<html lang=\"%s\">\n<head>\n", LocaleHTML)
	b.WriteString("  <meta charset=\"utf-8\" />\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	fmt.Fprintf(&b, "  <title>%s</title>\n", title)
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" />\n", esc(FontsGooglePreconnect))
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" crossorigin />\n", esc(FontsGoogleStatic))
	fmt.Fprintf(&b, "  <link href=\"%s\" rel=\"stylesheet\" />\n", esc(FontsGoogleCSSURL))
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\" />\n", esc(PathTokensCSS))
	b.WriteString(landingStyleBlock())
	b.WriteString("</head>\n<body>\n")
	b.WriteString("  <main class=\"landing\">\n    <div class=\"card\">\n")
	fmt.Fprintf(&b, "      <p class=\"brand\">%s<span>%s</span></p>\n", stem, accent)
	fmt.Fprintf(&b, "      <p class=\"lede\">%s</p>\n", esc(LandingCopyLede))
	fmt.Fprintf(&b, "      <form id=\"%s\" onsubmit=\"return landingMagicLink(event)\">\n", LandingDOMMagicForm)
	b.WriteString("        <div class=\"field\">\n")
	fmt.Fprintf(&b, "          <label for=\"%s\">%s</label>\n", LandingDOMEmailInput, esc(LandingCopyEmailLabel))
	fmt.Fprintf(&b, "          <input id=\"%s\" type=\"email\" required placeholder=\"%s\" autocomplete=\"email\" />\n",
		LandingDOMEmailInput, esc(AppCopyEmailPlaceholder))
	b.WriteString("        </div>\n")
	fmt.Fprintf(&b, "        <button class=\"btn\" type=\"submit\">%s</button>\n", esc(LandingCopySendLink))
	b.WriteString("      </form>\n")
	fmt.Fprintf(&b, "      <p class=\"hint\" id=\"%s\">%s</p>\n", LandingDOMAuthHint, esc(LandingCopySentHint))
	b.WriteString("    </div>\n  </main>\n")
	b.WriteString(landingScriptBlock())
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

func landingStyleBlock() string {
	return `  <style>
    * { box-sizing: border-box; }
    html, body {
      margin: 0; min-height: 100%;
      background:
        radial-gradient(1200px 600px at 10% -10%, color-mix(in srgb, var(--color-accent) 18%, transparent), transparent 60%),
        linear-gradient(180deg, var(--color-canvas-hi) 0%, var(--color-canvas) 45%, var(--color-canvas-lo) 100%);
      color: var(--color-text); font-family: var(--font-sans);
    }
    button, input { font: inherit; }
    .landing { min-height: 100vh; display: grid; place-items: center; padding: var(--space-lg); }
    .card { width: min(440px, 100%); }
    .brand { font-family: var(--font-display); font-weight: 800; font-size: clamp(3rem, 8vw, 4.5rem); letter-spacing: -.04em; line-height: .95; margin: 0 0 var(--space-md); }
    .brand span { color: var(--color-accent); }
    .lede { margin: 0 0 var(--space-xl); color: var(--color-text-muted); font-size: 1.05rem; line-height: 1.5; max-width: 34ch; }
    .field { display: grid; gap: var(--space-sm); margin-bottom: var(--space-md); }
    .field label { font-size: .8rem; letter-spacing: .08em; text-transform: uppercase; color: var(--color-text-muted); }
    .field input {
      width: 100%; padding: 14px 16px; border-radius: var(--radius-md);
      border: 1px solid var(--color-border); background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas));
      color: var(--color-text); outline: none;
    }
    .btn {
      display: inline-flex; align-items: center; justify-content: center; width: 100%;
      padding: 14px 18px; border: 0; border-radius: var(--radius-md);
      background: var(--color-accent); color: var(--color-on-accent); font-weight: 700; cursor: pointer;
    }
    .hint { margin-top: var(--space-lg); font-size: .9rem; color: var(--color-text-muted); }
  </style>
`
}

func landingScriptBlock() string {
	return `  <script>
    const hint = document.getElementById('` + LandingDOMAuthHint + `');
    async function landingMagicLink(e) {
      e.preventDefault();
      const email = document.getElementById('` + LandingDOMEmailInput + `').value.trim();
      if (!email.includes('@')) return false;
      const res = await fetch('` + PathV1AuthMagicLink + `', {
        method: 'POST',
        headers: { '` + HeaderContentType + `': '` + MIMEApplicationJSON + `' },
        body: JSON.stringify({ '` + JSONFieldEmail + `': email })
      });
      if (!res.ok) { hint.textContent = '` + LandingCopyRequestFailed + `'; return false; }
      hint.textContent = '` + LandingCopySentHint + `';
      return false;
    }
  </script>
`
}
