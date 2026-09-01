package tokens

import "html"

// Magic link open page copy (RU) — GET /v1/auth/magic-link/open error HTML.
const (
	MagicLinkOpenCopyInvalid      = "Ссылка недействительна или уже использована."
	MagicLinkOpenCopyMissingToken = "В ссылке нет токена."
	MagicLinkOpenPageTitle        = ProductNameLocal + " — вход"
)

// RenderMagicLinkOpenErrorHTML returns minimal HTML for open-link errors.
func RenderMagicLinkOpenErrorHTML(message string) string {
	msg := html.EscapeString(message)
	return `<!DOCTYPE html>
<html lang="` + LocaleHTML + `">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>` + MagicLinkOpenPageTitle + `</title>
  <link rel="stylesheet" href="` + PathTokensCSS + `" />
</head>
<body style="font-family:system-ui,sans-serif;padding:2rem;max-width:32rem;margin:auto">
  <p>` + msg + `</p>
  <p><a href="` + PathRoot + `">` + html.EscapeString(ProductNameLocal) + `</a></p>
</body>
</html>`
}
