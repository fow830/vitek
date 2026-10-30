package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"vitek/internal/tokens"
)

// TelegramAPIClient talks to api.telegram.org with the Bot HTTP API (no extra deps).
type TelegramAPIClient struct {
	token  string
	base   string
	client *http.Client
}

func NewTelegramAPIClient(token string) *TelegramAPIClient {
	return &TelegramAPIClient{
		token: strings.TrimSpace(token),
		base:  tokens.TelegramAPIBase,
		client: &http.Client{
			Timeout: tokens.TelegramHTTPTimeout,
		},
	}
}

func (c *TelegramAPIClient) Send(ctx context.Context, chatID, text string) error {
	return c.sendMessage(ctx, chatID, text, nil)
}

// SendWithKeyboard sends text with the persistent watch control keyboard.
func (c *TelegramAPIClient) SendWithKeyboard(ctx context.Context, chatID, text string) error {
	kb := tokens.TelegramControlKeyboard()
	return c.sendMessage(ctx, chatID, text, kb)
}

func (c *TelegramAPIClient) sendMessage(ctx context.Context, chatID, text string, keyboard map[string]any) error {
	if c == nil || c.token == "" {
		return fmt.Errorf("%s", tokens.ErrMsgTelegramTokenMissing)
	}
	body := map[string]any{
		tokens.TelegramJSONChatID: chatID,
		tokens.TelegramJSONText:   text,
	}
	if keyboard != nil {
		body[tokens.TelegramJSONReplyMarkup] = keyboard
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	endpoint := c.base + "/bot" + c.token + "/" + tokens.TelegramAPISendMessage
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set(tokens.HeaderContentType, tokens.MIMEApplicationJSON)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, int64(tokens.AvitoHTTPProbeDiscardBytes)))
	if !tokens.HTTPStatusOK(res.StatusCode) {
		return fmt.Errorf("%s: status %d", tokens.ErrMsgTelegramSendFailed, res.StatusCode)
	}
	return nil
}

type telegramUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

type telegramUpdatesResponse struct {
	OK     bool             `json:"ok"`
	Result []telegramUpdate `json:"result"`
}

func (c *TelegramAPIClient) GetUpdates(ctx context.Context, offset, timeoutSec int) ([]telegramUpdate, error) {
	if c == nil || c.token == "" {
		return nil, fmt.Errorf("%s", tokens.ErrMsgTelegramTokenMissing)
	}
	v := url.Values{}
	v.Set(tokens.TelegramJSONOffset, strconv.Itoa(offset))
	v.Set(tokens.TelegramJSONTimeout, strconv.Itoa(timeoutSec))
	v.Set(tokens.TelegramJSONAllowedUpdates, `["message"]`)
	endpoint := c.base + "/bot" + c.token + "/" + tokens.TelegramAPIGetUpdates + "?" + v.Encode()

	// Long poll may exceed default client timeout; use request-scoped client.
	client := &http.Client{Timeout: time.Duration(timeoutSec)*time.Second + tokens.TelegramPollClientSlack}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, int64(tokens.AvitoHTTPMaxBodyBytes)))
	if err != nil {
		return nil, err
	}
	if !tokens.HTTPStatusOK(res.StatusCode) {
		return nil, fmt.Errorf("%s: status %d", tokens.ErrMsgTelegramUpdatesFailed, res.StatusCode)
	}
	var out telegramUpdatesResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, fmt.Errorf("%s", tokens.ErrMsgTelegramUpdatesNotOK)
	}
	return out.Result, nil
}
