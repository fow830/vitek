package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"

	"vitek/internal/tokens"
)

// RodPageFetcher opens Avito pages with a pooled server-side Chrome (go-rod).
// One browser process per (proxy, user-data-dir); pages are opened/closed per fetch.
type RodPageFetcher struct {
	userDataDir string
	httpBase    string
	timeout     time.Duration
	headless    bool

	mu       sync.Mutex
	sessions map[string]*rodChromeSession
}

type rodChromeSession struct {
	mu      sync.Mutex
	browser *rod.Browser
	proxy   string
	dir     string
}

func NewRodPageFetcher(userDataDir, httpBase string) *RodPageFetcher {
	return &RodPageFetcher{
		userDataDir: strings.TrimSpace(userDataDir),
		httpBase:    strings.TrimSpace(httpBase),
		timeout:     tokens.RodPageFetchTimeout,
		headless:    tokens.RodHeadlessDefault,
		sessions:    make(map[string]*rodChromeSession),
	}
}

func (f *RodPageFetcher) FetchHTML(ctx context.Context, proxyEndpoint, userDataDir, pageURL string) (string, error) {
	pageURL = tokens.ListingSearchRewriteHTTPBase(pageURL, f.httpBase)
	dir := strings.TrimSpace(userDataDir)
	if dir == "" {
		dir = f.userDataDir
	}
	proxyEndpoint = strings.TrimSpace(proxyEndpoint)

	sess, err := f.session(ctx, proxyEndpoint, dir)
	if err != nil {
		return "", err
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()

	page, err := sess.browser.Page(proto.TargetCreateTarget{URL: tokens.RodAboutBlankURL})
	if err != nil {
		_ = f.dropSession(proxyEndpoint, dir)
		return "", fmt.Errorf("%s: %w", tokens.ErrMsgRodConnect, err)
	}
	defer func() { _ = page.Close() }()

	page = page.Context(ctx).Timeout(f.timeout)
	prepareRodChromePage(page)
	return rodNavigateHTML(ctx, page, pageURL, f.timeout)
}

// WarmProfile opens Avito home, dwells/scrolls, then a public feed — sticky profile stays pooled.
func (f *RodPageFetcher) WarmProfile(ctx context.Context, proxyEndpoint, userDataDir string) (string, error) {
	dir := strings.TrimSpace(userDataDir)
	if dir == "" {
		dir = f.userDataDir
	}
	proxyEndpoint = strings.TrimSpace(proxyEndpoint)
	timeout := tokens.RodChromeWarmTimeout

	sess, err := f.session(ctx, proxyEndpoint, dir)
	if err != nil {
		return "", err
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()

	page, err := sess.browser.Page(proto.TargetCreateTarget{URL: tokens.RodAboutBlankURL})
	if err != nil {
		_ = f.dropSession(proxyEndpoint, dir)
		return "", fmt.Errorf("%s: %w", tokens.ErrMsgRodConnect, err)
	}
	defer func() { _ = page.Close() }()

	page = page.Context(ctx).Timeout(timeout)
	prepareRodChromePage(page)

	home := tokens.ListingSearchRewriteHTTPBase(tokens.ProxyProbeURL(), f.httpBase)
	if _, err := rodNavigateHTML(ctx, page, home, timeout); err != nil {
		return "", err
	}
	if err := rodDwell(ctx, tokens.RodChromeWarmHomeDwell); err != nil {
		return "", err
	}
	if err := rodWarmScroll(ctx, page); err != nil {
		return "", err
	}

	second := tokens.ListingSearchRewriteHTTPBase(tokens.RodChromeWarmSecondaryURL(), f.httpBase)
	html, err := rodNavigateHTML(ctx, page, second, timeout)
	if err != nil {
		return "", err
	}
	if err := rodDwell(ctx, tokens.RodChromeWarmStepDwell); err != nil {
		return "", err
	}
	return html, nil
}

func rodNavigateHTML(ctx context.Context, page *rod.Page, pageURL string, timeout time.Duration) (string, error) {
	var load proto.PageLoadEventFired
	waitLoad := page.WaitEvent(&load)
	if err := page.Navigate(pageURL); err != nil {
		return "", newAvitoFetchFailure(tokens.AvitoFetchStateOther, 0, false)
	}
	rodWaitEvent(ctx, waitLoad, timeout)

	html, err := waitRodChromeContent(ctx, page, timeout)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(html) == "" {
		return "", newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, 0, false)
	}
	return html, nil
}

func rodDwell(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func rodWarmScroll(ctx context.Context, page *rod.Page) error {
	for i := 0; i < tokens.RodChromeWarmScrollSteps; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		_ = proto.InputDispatchMouseEvent{
			Type:   proto.InputDispatchMouseEventTypeMouseWheel,
			X:      tokens.RodChromeWarmScrollX,
			Y:      tokens.RodChromeWarmScrollY,
			DeltaX: 0,
			DeltaY: float64(tokens.RodChromeWarmScrollDelta),
		}.Call(page)
		if err := rodDwell(ctx, tokens.RodChromeWarmScrollPause); err != nil {
			return err
		}
	}
	return nil
}

func prepareRodChromePage(page *rod.Page) {
	meta := &proto.EmulationUserAgentMetadata{
		Brands: []*proto.EmulationUserAgentBrandVersion{
			{Brand: tokens.RodChromeBrandChrome, Version: tokens.RodChromeBrandChromeMajor},
			{Brand: tokens.RodChromeBrandChromium, Version: tokens.RodChromeBrandChromeMajor},
			{Brand: tokens.RodChromeBrandNotA, Version: tokens.RodChromeBrandNotAVersion},
		},
		FullVersionList: []*proto.EmulationUserAgentBrandVersion{
			{Brand: tokens.RodChromeBrandChrome, Version: tokens.RodChromeUAFullVersion},
			{Brand: tokens.RodChromeBrandChromium, Version: tokens.RodChromeUAFullVersion},
			{Brand: tokens.RodChromeBrandNotA, Version: tokens.RodChromeBrandNotAVersion + ".0.0.0"},
		},
		FullVersion:     tokens.RodChromeUAFullVersion,
		Platform:        tokens.RodChromeUADataPlatform,
		PlatformVersion: tokens.RodChromeUADataPlatformVersion,
		Architecture:    tokens.RodChromeUADataArchitecture,
		Model:           "",
		Mobile:          false,
		Bitness:         tokens.RodChromeUADataBitness,
		Wow64:           false,
	}
	_ = proto.EmulationSetUserAgentOverride{
		UserAgent:         tokens.RodChromeUserAgent,
		AcceptLanguage:    tokens.RodChromeAcceptLanguage,
		Platform:          tokens.RodChromePlatform,
		UserAgentMetadata: meta,
	}.Call(page)
	_ = proto.EmulationSetTimezoneOverride{TimezoneID: tokens.RodChromeTimezone}.Call(page)
	_ = proto.EmulationSetLocaleOverride{Locale: tokens.RodChromeLang}.Call(page)
	_, _ = page.EvalOnNewDocument(rodChromeStealthJS())
}

func rodChromeStealthJS() string {
	return fmt.Sprintf(`() => {
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		Object.defineProperty(navigator, 'languages', { get: () => ['ru-RU', 'ru', 'en-US', 'en'] });
		Object.defineProperty(navigator, 'language', { get: () => 'ru-RU' });
		Object.defineProperty(navigator, 'platform', { get: () => 'Win32' });
		Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => %d });
		Object.defineProperty(navigator, 'deviceMemory', { get: () => %d });
		Object.defineProperty(navigator, 'maxTouchPoints', { get: () => 0 });
		window.chrome = window.chrome || { runtime: {}, app: { isInstalled: false } };
		const FakePC = function() {
			this.onicecandidate = null;
			this.createDataChannel = () => ({ close() {}, addEventListener() {}, removeEventListener() {}, dispatchEvent() { return false; } });
			this.createOffer = () => Promise.resolve({ type: 'offer', sdp: '' });
			this.createAnswer = () => Promise.resolve({ type: 'answer', sdp: '' });
			this.setLocalDescription = () => Promise.resolve();
			this.setRemoteDescription = () => Promise.resolve();
			this.addIceCandidate = () => Promise.resolve();
			this.close = () => {};
			this.addEventListener = () => {};
			this.removeEventListener = () => {};
			this.dispatchEvent = () => false;
			this.getStats = () => Promise.resolve(new Map());
			this.connectionState = 'connected';
			this.iceConnectionState = 'connected';
			this.signalingState = 'stable';
		};
		try {
			Object.defineProperty(window, 'RTCPeerConnection', { configurable: true, writable: true, value: FakePC });
			Object.defineProperty(window, 'webkitRTCPeerConnection', { configurable: true, writable: true, value: FakePC });
		} catch (e) {}
	}`, tokens.RodChromeHardwareConcurrency, tokens.RodChromeDeviceMemoryGB)
}

func waitRodChromeContent(ctx context.Context, page *rod.Page, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	var last string
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		html, err := rodDOMOuterHTML(page)
		if err != nil {
			return "", err
		}
		last = html
		if tokens.AvitoSERPHTMLReady(html) {
			return html, nil
		}
		if time.Now().After(deadline) {
			if strings.TrimSpace(last) == "" {
				return "", newAvitoFetchFailure(tokens.AvitoFetchStateEmpty, 0, false)
			}
			return last, nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(tokens.RodChromeContentPoll):
		}
	}
}

// rodDOMOuterHTML reads document HTML via DOM domain (no Runtime.enable / Eval).
func rodDOMOuterHTML(page *rod.Page) (string, error) {
	_ = proto.DOMEnable{}.Call(page)
	doc, err := proto.DOMGetDocument{}.Call(page)
	if err != nil {
		return "", err
	}
	if doc.Root == nil {
		return "", nil
	}
	out, err := proto.DOMGetOuterHTML{NodeID: doc.Root.NodeID}.Call(page)
	if err != nil {
		return "", err
	}
	return out.OuterHTML, nil
}

func rodWaitEvent(ctx context.Context, wait func(), timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	case <-time.After(timeout):
	}
}

func (f *RodPageFetcher) session(ctx context.Context, proxyEndpoint, dir string) (*rodChromeSession, error) {
	key := tokens.RodChromeSessionKey(proxyEndpoint, dir)

	f.mu.Lock()
	if s := f.sessions[key]; s != nil {
		f.mu.Unlock()
		return s, nil
	}
	f.mu.Unlock()

	browser, err := launchRodChrome(ctx, proxyEndpoint, dir, f.headless)
	if err != nil {
		return nil, err
	}
	s := &rodChromeSession{browser: browser, proxy: proxyEndpoint, dir: dir}

	f.mu.Lock()
	defer f.mu.Unlock()
	if existing := f.sessions[key]; existing != nil {
		_ = browser.Close()
		return existing, nil
	}
	f.sessions[key] = s
	return s, nil
}

func (f *RodPageFetcher) dropSession(proxyEndpoint, dir string) error {
	key := tokens.RodChromeSessionKey(proxyEndpoint, dir)
	f.mu.Lock()
	s := f.sessions[key]
	delete(f.sessions, key)
	f.mu.Unlock()
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.browser != nil {
		return s.browser.Close()
	}
	return nil
}

func launchRodChrome(ctx context.Context, proxyEndpoint, dir string, headless bool) (*rod.Browser, error) {
	_ = os.Setenv(tokens.RodChromeEnvTZ, tokens.RodChromeTimezone)
	_ = os.Setenv(tokens.RodChromeEnvLANG, tokens.RodChromeLANGValue)
	l := launcher.New().NoSandbox(true).
		Set(tokens.RodChromeArgDisableDevShm).
		Set(tokens.RodChromeArgLang, tokens.RodChromeLang).
		Set(tokens.RodChromeArgDisableBlink, tokens.RodChromeBlinkAutomation).
		Set(tokens.RodChromeArgForceWebRTCIPHandling, tokens.RodChromeWebRTCIPHandlingPolicy).
		Set(tokens.RodChromeArgUserAgent, tokens.RodChromeUserAgent).
		Set(tokens.RodChromeArgWindowSize, tokens.RodChromeWindowSize)
	for _, flag := range tokens.RodChromeStealthDeleteFlags {
		l = l.Delete(flags.Flag(flag))
	}
	bin := strings.TrimSpace(os.Getenv(tokens.EnvRodBrowser))
	if bin == "" {
		bin = tokens.DefaultRodChromeBin
	}
	if bin != "" {
		l = l.Bin(bin)
	}
	switch {
	case headless:
		l = l.HeadlessNew(true)
	case tokens.RodChromeXVFBDefault:
		l = l.XVFB(tokens.RodChromeXVFBArgAuto, tokens.RodChromeXVFBArgScreen, tokens.RodChromeXVFBScreen).Headless(false)
	default:
		l = l.Headless(false)
	}
	if dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("%s: %w", tokens.ErrMsgRodLaunch, err)
		}
		if err := ensureRodChromeProfilePrefs(dir); err != nil {
			return nil, fmt.Errorf("%s: %w", tokens.ErrMsgRodLaunch, err)
		}
		l = l.UserDataDir(dir)
	}
	if proxyEndpoint != "" && !tokens.IsDirectProxyEndpoint(proxyEndpoint) {
		proxyURL, err := url.Parse(proxyEndpoint)
		if err != nil {
			return nil, err
		}
		l = l.Proxy(proxyURL.String())
	}
	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", tokens.ErrMsgRodLaunch, err)
	}
	browser := rod.New().ControlURL(controlURL).Context(ctx).NoDefaultDevice()
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("%s: %w", tokens.ErrMsgRodConnect, err)
	}
	return browser, nil
}

func ensureRodChromeProfilePrefs(dir string) error {
	prefDir := filepath.Join(dir, tokens.RodChromeProfileDirName)
	if err := os.MkdirAll(prefDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(prefDir, tokens.RodChromePreferencesFile)
	prefs := map[string]any{}
	if raw, err := os.ReadFile(path); err == nil && len(raw) > 0 {
		_ = json.Unmarshal(raw, &prefs)
	}
	webrtc, _ := prefs[tokens.RodChromePrefWebRTC].(map[string]any)
	if webrtc == nil {
		webrtc = map[string]any{}
	}
	webrtc[tokens.RodChromePrefWebRTCIPHandling] = tokens.RodChromeWebRTCIPHandlingPolicy
	webrtc[tokens.RodChromePrefWebRTCMultipleRoutes] = false
	webrtc[tokens.RodChromePrefWebRTCNonProxiedUDP] = false
	prefs[tokens.RodChromePrefWebRTC] = webrtc
	intl, _ := prefs[tokens.RodChromePrefIntl].(map[string]any)
	if intl == nil {
		intl = map[string]any{}
	}
	langList := tokens.RodChromeLang + tokens.RodChromeLangFallbackList
	intl[tokens.RodChromePrefSelectedLanguages] = langList
	intl[tokens.RodChromePrefAcceptLanguages] = langList
	prefs[tokens.RodChromePrefIntl] = intl
	profile, _ := prefs[tokens.RodChromePrefProfile].(map[string]any)
	if profile == nil {
		profile = map[string]any{}
	}
	profile[tokens.RodChromePrefExitType] = tokens.RodChromePrefExitTypeNormal
	profile[tokens.RodChromePrefExitedCleanly] = true
	prefs[tokens.RodChromePrefProfile] = profile
	out, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}
