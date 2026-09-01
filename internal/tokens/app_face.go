package tokens

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

// App face title / live clock label.
const (
	AppFaceTitleSuffix = "Platform"
	AppCopyLiveClock   = "LIVE"
)

// AppNav IDs (shell views).
const (
	AppNavIDOverview = "overview"
	AppNavIDSearch   = "search"
	AppNavIDServices = "services"
	AppNavIDAvito    = "avito"
	AppNavIDProxies  = "proxies"
)

// App DOM ids — shared by face HTML, face JS, and Datastar SSE patches.
const (
	AppDOMScreenAuth  = "screen-auth"
	AppDOMScreenPlatform = "screen-platform"
	AppDOMNav         = "app-nav"
	AppDOMStatAvito   = "stat-avito"
	AppDOMStatProxy   = "stat-proxy"
	AppDOMStatShipped = "stat-shipped"
	AppDOMTick        = "tick"
	AppDOMAuthHint    = "auth-hint"
	AppDOMMagicForm   = "magic-form"
	AppDOMEmailInput  = "email"
	AppDOMSearchForm    = "search-form"
	AppDOMSearchURL     = "search-url"
	AppDOMSearchResults = "search-results"
	AppDOMSearchStatus  = "search-status"
	AppClassIsActive  = "is-active"
	AppClassStatusOk  = "status-ok"
	AppClassStatusOff = "status-off"
	AppDOMViewPrefix  = "view-"
)

// App UI copy (RU) — SoT for generated app/face.html.
const (
	AppCopyAuthLede         = "Платформа. Вход только по Magic Link — без паролей."
	AppCopyEmailLabel       = "Email"
	AppCopyEmailPlaceholder = "you@company.ru"
	AppCopySendLink         = "Прислать ссылку"
	AppCopyAuthHint         = "Magic Link: письмо → " + ProductDomainApp + "."
	AppCopySentHint         = "Ссылка создана. Локально token вернётся в ответе API."
	AppCopyRequestFailed    = "Ошибка запроса Magic Link"
	AppCopyConsumeFailed    = "Ошибка consume"
	AppCopyOverviewTitle    = "Обзор"
	AppCopyOverviewLede     = "Платформа сервисов. Shipped и reserved — из каталога токенов."
	AppCopySearchTitle      = "Поиск похожих"
	AppCopySearchLede       = "Вставьте ссылку на фильтр Avito. Первый запуск сохраняет базу — дальше только новые объявления."
	AppCopySearchURLLabel   = "Ссылка на фильтр"
	AppCopySearchSubmit     = "Искать"
	AppCopySearchStatusIdle = "Создайте задачу или выберите из списка."
	AppCopySearchFailed     = "Ошибка поиска"
	AppCopyServicesTitle    = "Сервисы"
	AppCopyServicesLede     = "Каталог product_services (SoT = tokens + БД)."
	AppCopyAvitoTitle       = "Аккаунты Авито"
	AppCopyAvitoLede        = "Пул аккаунтов для worker."
	AppCopyProxiesTitle     = "Прокси"
	AppCopyProxiesLede      = "Прокси для запросов к Avito."
	AppCopyEmptyPanel       = "Нет строк — создайте через platform API."
	AppCopyLogout           = "Выйти"
	AppCopyStatAvito        = "Аккаунты Авито"
	AppCopyStatProxies      = "Прокси ACTIVE"
	AppCopyStatShipped      = "Сервисы shipped"
	AppCopyColCode          = "Code"
	AppCopyColTitle         = "Title"
	AppCopyColShipped       = "Shipped"
)

// AppSSEStatsPatch is the Datastar element patch for overview counters.
func AppSSEStatsPatch(avitoN, proxyN, shippedN int) string {
	return `<b id="` + AppDOMStatAvito + `">` + strconv.Itoa(avitoN) + `</b>` +
		`<b id="` + AppDOMStatProxy + `">` + strconv.Itoa(proxyN) + `</b>` +
		`<b id="` + AppDOMStatShipped + `">` + strconv.Itoa(shippedN) + `</b>`
}

// AppNavItem is a shell navigation entry.
type AppNavItem struct {
	ID    string
	Label string
}

// AppNav is the platform shell navigation SoT.
var AppNav = []AppNavItem{
	{ID: AppNavIDOverview, Label: AppCopyOverviewTitle},
	{ID: AppNavIDSearch, Label: AppCopySearchTitle},
	{ID: AppNavIDServices, Label: AppCopyServicesTitle},
	{ID: AppNavIDAvito, Label: AppCopyAvitoTitle},
	{ID: AppNavIDProxies, Label: AppCopyProxiesTitle},
}

// FixtureSessionEmail is the side-meta identity shown before login (ProductDomain-derived).
func FixtureSessionEmail() string {
	return ProductEmail("admin")
}

// ProductBrandStem / ProductBrandAccent split ProductNameLocal for display accent.
func ProductBrandStem() string {
	r := []rune(ProductNameLocal)
	if len(r) < 2 {
		return ProductNameLocal
	}
	return string(r[:len(r)-2])
}

func ProductBrandAccent() string {
	r := []rune(ProductNameLocal)
	if len(r) < 2 {
		return ""
	}
	return string(r[len(r)-2:])
}

// RenderAppFaceHTML returns the canonical web/app/face.html body (tokens only).
func RenderAppFaceHTML() string {
	var b strings.Builder
	esc := html.EscapeString
	stem := esc(ProductBrandStem())
	accent := esc(ProductBrandAccent())
	title := esc(ProductNameLocal + " — " + AppFaceTitleSuffix)
	emailPh := esc(AppCopyEmailPlaceholder)
	fixtureEmail := esc(FixtureSessionEmail())

	b.WriteString("<!DOCTYPE html>\n")
	fmt.Fprintf(&b, "<html lang=\"%s\" %s>\n<head>\n", LocaleHTML, AttrDataStar)
	b.WriteString("  <meta charset=\"utf-8\" />\n")
	b.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n")
	fmt.Fprintf(&b, "  <title>%s</title>\n", title)
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" />\n", esc(FontsGooglePreconnect))
	fmt.Fprintf(&b, "  <link rel=\"preconnect\" href=\"%s\" crossorigin />\n", esc(FontsGoogleStatic))
	fmt.Fprintf(&b, "  <link href=\"%s\" rel=\"stylesheet\" />\n", esc(FontsGoogleCSSURL))
	fmt.Fprintf(&b, "  <link rel=\"stylesheet\" href=\"%s\" />\n", esc(PathTokensCSS))
	fmt.Fprintf(&b, "  <script type=\"module\" src=\"%s\" data-star></script>\n", esc(DatastarCDNURL))
	b.WriteString(appFaceStyleBlock())
	b.WriteString("</head>\n<body>\n")

	// Auth
	fmt.Fprintf(&b, "  <section id=\"%s\" class=\"screen %s\">\n", AppDOMScreenAuth, AppClassIsActive)
	b.WriteString("    <div class=\"auth\">\n      <div class=\"auth-card\">\n")
	fmt.Fprintf(&b, "        <p class=\"brand\">%s<span>%s</span></p>\n", stem, accent)
	fmt.Fprintf(&b, "        <p class=\"lede\">%s</p>\n", esc(AppCopyAuthLede))
	fmt.Fprintf(&b, "        <form id=\"%s\" onsubmit=\"return enterApp(event)\">\n", AppDOMMagicForm)
	b.WriteString("          <div class=\"field\">\n")
	fmt.Fprintf(&b, "            <label for=\"%s\">%s</label>\n", AppDOMEmailInput, esc(AppCopyEmailLabel))
	fmt.Fprintf(&b, "            <input id=\"%s\" type=\"email\" required placeholder=\"%s\" autocomplete=\"email\" />\n", AppDOMEmailInput, emailPh)
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <button class=\"btn\" type=\"submit\">%s</button>\n", esc(AppCopySendLink))
	b.WriteString("        </form>\n")
	fmt.Fprintf(&b, "        <p class=\"hint\" id=\"%s\">%s</p>\n", AppDOMAuthHint, esc(AppCopyAuthHint))
	b.WriteString("      </div>\n    </div>\n  </section>\n\n")

	// Platform shell
	fmt.Fprintf(&b, "  <section id=\"%s\" class=\"screen\">\n    <div class=\"shell\">\n", AppDOMScreenPlatform)
	b.WriteString("      <aside class=\"side\">\n")
	fmt.Fprintf(&b, "        <div class=\"side-brand\">%s<em>%s</em></div>\n", stem, accent)
	fmt.Fprintf(&b, "        <nav class=\"nav\" id=\"%s\">\n", AppDOMNav)
	for i, item := range AppNav {
		active := ""
		if i == 0 {
			active = ` class="` + AppClassIsActive + `"`
		}
		fmt.Fprintf(&b, "          <a href=\"#%s\"%s data-view=\"%s\">%s</a>\n",
			esc(item.ID), active, esc(item.ID), esc(item.Label))
	}
	b.WriteString("        </nav>\n        <div class=\"side-meta\">\n")
	fmt.Fprintf(&b, "          <span class=\"pulse\"><i></i> %s</span>\n", esc(AppCopyLiveClock))
	fmt.Fprintf(&b, "          <span>%s</span>\n", fixtureEmail)
	fmt.Fprintf(&b, "          <a href=\"#\" onclick=\"return backAuth()\">%s</a>\n", esc(AppCopyLogout))
	b.WriteString("        </div>\n      </aside>\n\n      <main class=\"main\">\n")

	// Overview
	shippedN := strconv.Itoa(len(ShippedServiceCodes()))
	fmt.Fprintf(&b, "        <div class=\"view %s\" id=\"%s%s\">\n", AppClassIsActive, AppDOMViewPrefix, AppNavIDOverview)
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(&b, "            <h1>%s</h1>\n", esc(AppCopyOverviewTitle))
	fmt.Fprintf(&b, "            <p>%s</p>\n", esc(AppCopyOverviewLede))
	b.WriteString("          </div>\n")
	fmt.Fprintf(&b, "          <div class=\"live-bar\">%s <samp id=\"%s\">--:--:--</samp></div>\n", esc(AppCopyLiveClock), AppDOMTick)
	b.WriteString("          </div>\n          <div class=\"grid\">\n")
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">—</b></div>\n", esc(AppCopyStatAvito), AppDOMStatAvito)
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">—</b></div>\n", esc(AppCopyStatProxies), AppDOMStatProxy)
	fmt.Fprintf(&b, "            <div class=\"stat\"><span>%s</span><b id=\"%s\">%s</b></div>\n", esc(AppCopyStatShipped), AppDOMStatShipped, esc(shippedN))
	b.WriteString("          </div>\n          <div class=\"panel\">\n")
	b.WriteString("            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopyServicesTitle))
	b.WriteString("</h2><span class=\"chip\">")
	b.WriteString(esc(AuthMethodMagicLink))
	b.WriteString("</span></div>\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Search view (listing_search UI)
	fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AppDOMViewPrefix, AppNavIDSearch)
	writeViewHeader(&b, AppCopySearchTitle, AppCopySearchLede)
	b.WriteString("          <div class=\"panel\">\n            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopySearchTitle))
	b.WriteString("</h2></div>\n            <div style=\"padding:18px\">\n")
	fmt.Fprintf(&b, "              <form id=\"%s\" onsubmit=\"return submitSearch(event)\">\n", AppDOMSearchForm)
	b.WriteString("                <div class=\"field\">\n")
	fmt.Fprintf(&b, "                  <label for=\"%s\">%s</label>\n", AppDOMSearchURL, esc(AppCopySearchURLLabel))
	fmt.Fprintf(&b, "                  <input id=\"%s\" type=\"url\" required placeholder=\"%s\" />\n", AppDOMSearchURL, esc(FixtureListingURL))
	b.WriteString("                </div>\n")
	fmt.Fprintf(&b, "                <button class=\"btn\" type=\"submit\">%s</button>\n", esc(AppCopySearchSubmit))
	b.WriteString("              </form>\n")
	fmt.Fprintf(&b, "              <p class=\"hint\" id=\"%s\">%s</p>\n", AppDOMSearchStatus, esc(AppCopySearchStatusIdle))
	fmt.Fprintf(&b, "              <div id=\"%s\"></div>\n", AppDOMSearchResults)
	b.WriteString("            </div>\n          </div>\n        </div>\n\n")

	// Services view
	fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AppDOMViewPrefix, AppNavIDServices)
	writeViewHeader(&b, AppCopyServicesTitle, AppCopyServicesLede)
	b.WriteString("          <div class=\"panel\">\n")
	writeServicesTable(&b)
	b.WriteString("          </div>\n        </div>\n\n")

	// Avito admin panel
	fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AppDOMViewPrefix, AppNavIDAvito)
	writeViewHeader(&b, AppCopyAvitoTitle, AppCopyAvitoLede)
	writeAdminAvitoPanel(&b)
	b.WriteString("        </div>\n\n")

	// Proxies admin panel
	fmt.Fprintf(&b, "        <div class=\"view\" id=\"%s%s\">\n", AppDOMViewPrefix, AppNavIDProxies)
	writeViewHeader(&b, AppCopyProxiesTitle, AppCopyProxiesLede)
	writeAdminProxiesPanel(&b)
	b.WriteString("        </div>\n\n")

	b.WriteString("      </main>\n    </div>\n  </section>\n")
	b.WriteString(appFaceScriptBlock())
	b.WriteString("</body>\n</html>\n")
	return b.String()
}

// RenderAppFaceHTMLLoggedIn returns platform shell for an authenticated session.
// SSE boot is attached only when withSSE is true (ADMIN).
func RenderAppFaceHTMLLoggedIn(email, _ string, withSSE bool) string {
	out := RenderAppFaceHTML()
	out = strings.Replace(out,
		`id="`+AppDOMScreenAuth+`" class="screen `+AppClassIsActive+`"`,
		`id="`+AppDOMScreenAuth+`" class="screen"`, 1)
	platformActive := `id="` + AppDOMScreenPlatform + `" class="screen ` + AppClassIsActive + `"`
	if withSSE {
		platformActive += `" data-on:load="@get('` + PathAppSSE + `')"`
	} else {
		platformActive += `"`
	}
	out = strings.Replace(out,
		`id="`+AppDOMScreenPlatform+`" class="screen"`,
		platformActive, 1)
	out = strings.Replace(out, html.EscapeString(FixtureSessionEmail()), html.EscapeString(email), 1)
	extra := appFaceSearchScriptBlock()
	if withSSE {
		extra += appFaceAdminScriptBlock()
	}
	out = strings.Replace(out, "</body>", extra+"</body>", 1)
	return out
}

func writeViewHeader(b *strings.Builder, title, lede string) {
	esc := html.EscapeString
	b.WriteString("          <div class=\"top\"><div>\n")
	fmt.Fprintf(b, "            <h1>%s</h1>\n", esc(title))
	fmt.Fprintf(b, "            <p>%s</p>\n", esc(lede))
	b.WriteString("          </div></div>\n")
}

func writeAdminProxiesPanel(b *strings.Builder) {
	esc := html.EscapeString
	b.WriteString("          <div class=\"panel\">\n            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopyProxiesTitle))
	b.WriteString("</h2></div>\n            <table>\n              <thead><tr>")
	fmt.Fprintf(b, "<th>%s</th><th>%s</th><th>%s</th><th>%s</th>",
		esc(AppCopyColLabel), esc(AppCopyColEndpoint), esc(AppCopyColStatus), esc(AppCopyColActions))
	b.WriteString("</tr></thead>\n")
	fmt.Fprintf(b, "              <tbody id=\"%s\"><tr><td colspan=\"4\">%s</td></tr></tbody>\n", AppDOMProxiesTable, esc(AppCopyAdminEmpty))
	b.WriteString("            </table>\n          </div>\n          <div class=\"panel\" style=\"margin-top:16px\">\n            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopyAdminCreate))
	b.WriteString("</h2></div>\n            <div style=\"padding:18px\">\n")
	fmt.Fprintf(b, "              <form id=\"%s\" onsubmit=\"return saveProxy(event)\">\n", AppDOMProxiesForm)
	fmt.Fprintf(b, "                <input type=\"hidden\" id=\"%s\" value=\"\" />\n", AppDOMProxiesEditID)
	b.WriteString("                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMProxiesLabel, esc(AppCopyColLabel))
	fmt.Fprintf(b, "                  <input id=\"%s\" required />\n", AppDOMProxiesLabel)
	b.WriteString("                </div>\n                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMProxiesEndpoint, esc(AppCopyColEndpoint))
	fmt.Fprintf(b, "                  <input id=\"%s\" required placeholder=\"%s\" />\n", AppDOMProxiesEndpoint, esc(FixtureAdminProxyEndpoint))
	b.WriteString("                </div>\n                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMProxiesStatus, esc(AppCopyColStatus))
	fmt.Fprintf(b, "                  <select id=\"%s\" required>\n", AppDOMProxiesStatus)
	writeStatusOptions(b, ProxyStatusValues)
	b.WriteString("                  </select>\n                </div>\n")
	fmt.Fprintf(b, "                <button class=\"btn\" type=\"submit\" id=\"%s-submit\">%s</button>\n", AppDOMProxiesForm, esc(AppCopyAdminCreate))
	fmt.Fprintf(b, "                <button class=\"btn btn-ghost\" type=\"button\" onclick=\"return resetProxyForm()\" style=\"margin-top:8px\">%s</button>\n", esc(AppCopyAdminCancel))
	b.WriteString("              </form>\n")
	fmt.Fprintf(b, "              <p class=\"hint\" id=\"%s\"></p>\n", AppDOMProxiesFormStatus)
	b.WriteString("            </div>\n          </div>\n")
}

func writeAdminAvitoPanel(b *strings.Builder) {
	esc := html.EscapeString
	b.WriteString("          <div class=\"panel\">\n            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopyAvitoTitle))
	b.WriteString("</h2></div>\n            <table>\n              <thead><tr>")
	fmt.Fprintf(b, "<th>%s</th><th>%s</th><th>%s</th><th>%s</th>",
		esc(AppCopyColLabel), esc(AppCopyColExternalRef), esc(AppCopyColStatus), esc(AppCopyColActions))
	b.WriteString("</tr></thead>\n")
	fmt.Fprintf(b, "              <tbody id=\"%s\"><tr><td colspan=\"4\">%s</td></tr></tbody>\n", AppDOMAvitoTable, esc(AppCopyAdminEmpty))
	b.WriteString("            </table>\n          </div>\n          <div class=\"panel\" style=\"margin-top:16px\">\n            <div class=\"panel-head\"><h2>")
	b.WriteString(esc(AppCopyAdminCreate))
	b.WriteString("</h2></div>\n            <div style=\"padding:18px\">\n")
	fmt.Fprintf(b, "              <form id=\"%s\" onsubmit=\"return saveAvito(event)\">\n", AppDOMAvitoForm)
	fmt.Fprintf(b, "                <input type=\"hidden\" id=\"%s\" value=\"\" />\n", AppDOMAvitoEditID)
	b.WriteString("                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMAvitoLabel, esc(AppCopyColLabel))
	fmt.Fprintf(b, "                  <input id=\"%s\" required />\n", AppDOMAvitoLabel)
	b.WriteString("                </div>\n                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMAvitoExternalRef, esc(AppCopyColExternalRef))
	fmt.Fprintf(b, "                  <input id=\"%s\" required />\n", AppDOMAvitoExternalRef)
	b.WriteString("                </div>\n                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\">%s</label>\n", AppDOMAvitoStatus, esc(AppCopyColStatus))
	fmt.Fprintf(b, "                  <select id=\"%s\" required>\n", AppDOMAvitoStatus)
	writeStatusOptions(b, AvitoAccountStatusValues)
	b.WriteString("                  </select>\n                </div>\n                <div class=\"field\">\n")
	fmt.Fprintf(b, "                  <label for=\"%s\" id=\"%s-label\">%s</label>\n", AppDOMAvitoPassword, AppDOMAvitoPassword, esc(AppCopyAdminPasswordRequired))
	fmt.Fprintf(b, "                  <input id=\"%s\" type=\"password\" autocomplete=\"new-password\" required />\n", AppDOMAvitoPassword)
	b.WriteString("                </div>\n")
	fmt.Fprintf(b, "                <button class=\"btn\" type=\"submit\" id=\"%s-submit\">%s</button>\n", AppDOMAvitoForm, esc(AppCopyAdminCreate))
	fmt.Fprintf(b, "                <button class=\"btn btn-ghost\" type=\"button\" onclick=\"return resetAvitoForm()\" style=\"margin-top:8px\">%s</button>\n", esc(AppCopyAdminCancel))
	b.WriteString("              </form>\n")
	fmt.Fprintf(b, "              <p class=\"hint\" id=\"%s\"></p>\n", AppDOMAvitoFormStatus)
	b.WriteString("            </div>\n          </div>\n")
}

func writeStatusOptions(b *strings.Builder, values []string) {
	esc := html.EscapeString
	for _, v := range values {
		fmt.Fprintf(b, "                    <option value=\"%s\">%s</option>\n", esc(v), esc(v))
	}
}

func writeServicesTable(b *strings.Builder) {
	esc := html.EscapeString
	b.WriteString("            <table>\n              <thead><tr>")
	fmt.Fprintf(b, "<th>%s</th><th>%s</th><th>%s</th>", esc(AppCopyColCode), esc(AppCopyColTitle), esc(AppCopyColShipped))
	b.WriteString("</tr></thead>\n              <tbody>\n")
	for _, s := range ProductServiceCatalog {
		shippedClass := AppClassStatusOff
		shippedVal := BoolStringFalse
		if s.Shipped {
			shippedClass = AppClassStatusOk
			shippedVal = BoolStringTrue
		}
		fmt.Fprintf(b, "                <tr><td class=\"mono\">%s</td><td>%s</td><td><span class=\"status %s\">%s</span></td></tr>\n",
			esc(s.Code), esc(s.Title), shippedClass, shippedVal)
	}
	b.WriteString("              </tbody>\n            </table>\n")
}

func appFaceStyleBlock() string {
	// Layout chrome only; colors/fonts/spacing via CSS variables from tokens.css.
	return `  <style>
    * { box-sizing: border-box; }
    html, body {
      margin: 0; min-height: 100%;
      background:
        radial-gradient(1200px 600px at 10% -10%, color-mix(in srgb, var(--color-accent) 18%, transparent), transparent 60%),
        radial-gradient(900px 500px at 100% 0%, color-mix(in srgb, var(--color-success) 10%, transparent), transparent 55%),
        linear-gradient(180deg, var(--color-canvas-hi) 0%, var(--color-canvas) 45%, var(--color-canvas-lo) 100%);
      color: var(--color-text); font-family: var(--font-sans);
    }
    button, input { font: inherit; }
    a { color: inherit; text-decoration: none; }
    .screen { display: none; min-height: 100vh; }
    .screen.` + AppClassIsActive + ` { display: block; }
    .auth { min-height: 100vh; display: grid; place-items: center; padding: var(--space-lg); position: relative; overflow: hidden; }
    .auth::before {
      content: ""; position: absolute; inset: auto -20% -30% -20%; height: 55%;
      background:
        linear-gradient(90deg, transparent, color-mix(in srgb, var(--color-accent) 12%, transparent), transparent),
        repeating-linear-gradient(90deg, color-mix(in srgb, var(--color-border) 55%, transparent) 0 1px, transparent 1px 48px);
      transform: perspective(500px) rotateX(58deg); pointer-events: none; opacity: .55;
    }
    .auth-card { width: min(440px, 100%); position: relative; z-index: 1; }
    .brand { font-family: var(--font-display); font-weight: 800; font-size: clamp(3rem, 8vw, 4.5rem); letter-spacing: -.04em; line-height: .95; margin: 0 0 var(--space-md); }
    .brand span { color: var(--color-accent); }
    .lede { margin: 0 0 var(--space-xl); color: var(--color-text-muted); font-size: 1.05rem; line-height: 1.5; max-width: 34ch; }
    .field { display: grid; gap: var(--space-sm); margin-bottom: var(--space-md); }
    .field label { font-size: .8rem; letter-spacing: .08em; text-transform: uppercase; color: var(--color-text-muted); }
    .field input {
      width: 100%; padding: 14px 16px; border-radius: var(--radius-md);
      border: 1px solid var(--color-border);       background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas));
      color: var(--color-text); outline: none;
    }
    .field input:focus { border-color: color-mix(in srgb, var(--color-accent) 70%, var(--color-border)); box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent) 22%, transparent); }
    .btn {
      display: inline-flex; align-items: center; justify-content: center; gap: var(--space-sm);
      width: 100%; padding: 14px 18px; border: 0; border-radius: var(--radius-md);
      background: var(--color-accent); color: var(--color-on-accent); font-weight: 700; cursor: pointer;
    }
    .btn:hover { filter: brightness(1.06); }
    .btn-ghost { background: transparent; color: var(--color-text-muted); border: 1px solid var(--color-border); }
    .btn-sm { width: auto; padding: 6px 12px; font-size: .85rem; margin-right: 6px; }
    select { width: 100%; padding: 14px 16px; border-radius: var(--radius-md); border: 1px solid var(--color-border); background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas)); color: var(--color-text); }
    .row-actions { white-space: nowrap; }
    .hint { margin-top: var(--space-lg); font-size: .9rem; color: var(--color-text-muted); }
    .shell { display: grid; grid-template-columns: 260px 1fr; min-height: 100vh; }
    @media (max-width: 900px) { .shell { grid-template-columns: 1fr; } .side { position: sticky; top: 0; z-index: 5; } }
    .side { padding: var(--space-lg); border-right: 1px solid var(--color-border); background: color-mix(in srgb, var(--color-surface) 70%, transparent); backdrop-filter: blur(10px); }
    .side-brand { font-family: var(--font-display); font-weight: 800; font-size: 1.6rem; letter-spacing: -.03em; margin-bottom: var(--space-xl); }
    .side-brand em { font-style: normal; color: var(--color-accent); }
    .nav { display: grid; gap: 4px; }
    .nav a { padding: 10px 12px; border-radius: var(--radius-sm); color: var(--color-text-muted); font-weight: 500; }
    .nav a.` + AppClassIsActive + `, .nav a:hover { background: color-mix(in srgb, var(--color-accent) 12%, transparent); color: var(--color-text); }
    .side-meta { margin-top: var(--space-xl); padding-top: var(--space-lg); border-top: 1px solid var(--color-border); font-size: .85rem; color: var(--color-text-muted); display: grid; gap: 6px; }
    .pulse { display: inline-flex; align-items: center; gap: 8px; color: var(--color-success); font-family: var(--font-mono); font-size: .78rem; }
    .pulse i { width: 8px; height: 8px; border-radius: 50%; background: var(--color-success); animation: ping 1.6s infinite; }
    @keyframes ping { 0% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--color-success) 55%, transparent); } 70% { box-shadow: 0 0 0 10px transparent; } 100% { box-shadow: 0 0 0 0 transparent; } }
    .main { padding: var(--space-xl) var(--space-lg); }
    .top { display: flex; flex-wrap: wrap; align-items: end; justify-content: space-between; gap: var(--space-md); margin-bottom: var(--space-xl); }
    .top h1 { margin: 0; font-family: var(--font-display); font-size: clamp(1.8rem, 3vw, 2.4rem); letter-spacing: -.03em; }
    .top p { margin: 6px 0 0; color: var(--color-text-muted); max-width: 48ch; }
    .chip { font-family: var(--font-mono); font-size: .75rem; padding: 6px 10px; border-radius: var(--radius-md); border: 1px solid var(--color-border); color: var(--color-text-muted); background: color-mix(in srgb, var(--color-surface) 80%, transparent); }
    .grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: var(--space-md); margin-bottom: var(--space-xl); }
    @media (max-width: 900px) { .grid { grid-template-columns: 1fr; } }
    .stat { padding: var(--space-lg); border: 1px solid var(--color-border); border-radius: var(--radius-md); background: linear-gradient(160deg, color-mix(in srgb, var(--color-surface) 92%, var(--color-text)), var(--color-surface)); }
    .stat b { display: block; font-family: var(--font-display); font-size: 2rem; letter-spacing: -.03em; margin-top: var(--space-sm); }
    .stat span { color: var(--color-text-muted); font-size: .9rem; }
    .panel { border: 1px solid var(--color-border); border-radius: var(--radius-md); background: color-mix(in srgb, var(--color-surface) 88%, var(--color-canvas)); overflow: hidden; }
    .panel.empty { padding: var(--space-xl); color: var(--color-text-muted); }
    .panel-head { display: flex; justify-content: space-between; align-items: center; gap: var(--space-md); padding: var(--space-md) var(--space-lg); border-bottom: 1px solid var(--color-border); }
    .panel-head h2 { margin: 0; font-size: 1rem; font-weight: 600; }
    table { width: 100%; border-collapse: collapse; }
    th, td { text-align: left; padding: 14px 18px; border-bottom: 1px solid color-mix(in srgb, var(--color-border) 70%, transparent); font-size: .92rem; }
    th { color: var(--color-text-muted); font-weight: 500; font-size: .75rem; letter-spacing: .06em; text-transform: uppercase; }
    tr:last-child td { border-bottom: 0; }
    .status { font-family: var(--font-mono); font-size: .78rem; padding: 4px 8px; border-radius: var(--radius-sm); }
    .` + AppClassStatusOk + ` { color: var(--color-success); background: color-mix(in srgb, var(--color-success) 14%, transparent); }
    .` + AppClassStatusOff + ` { color: var(--color-text-muted); background: color-mix(in srgb, var(--color-border) 35%, transparent); }
    .mono { font-family: var(--font-mono); font-size: .85rem; color: var(--color-text-muted); }
    .view { display: none; }
    .view.` + AppClassIsActive + ` { display: block; animation: in .35s ease; }
    @keyframes in { from { opacity: 0; transform: translateY(6px); } to { opacity: 1; transform: none; } }
    .live-bar { display: flex; align-items: center; gap: var(--space-sm); color: var(--color-text-muted); font-size: .85rem; }
    .live-bar samp { font-family: var(--font-mono); color: var(--color-accent); }
  </style>
`
}
