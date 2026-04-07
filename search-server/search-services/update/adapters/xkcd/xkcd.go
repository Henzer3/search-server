package xkcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"yadro.com/course/update/core"
)

type NumLastId struct {
	NumsComics int `json:"num"`
}

type InformationPicture struct {
	ID         int    `json:"num"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	ImageURL   string `json:"img"`
	Title      string `json:"title"`
}

type Client struct {
	log    *slog.Logger
	client http.Client
	url    string
}

func NewClient(url string, timeout time.Duration, log *slog.Logger) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("empty base url specified")
	}
	return &Client{
		client: http.Client{Timeout: timeout},
		log:    log,
		url:    url,
	}, nil
}

func (c *Client) Get(ctx context.Context, id int) (core.XKCDInfo, error) {
	if id <= 0 {
		c.log.Error("id should be more than 0", "err", core.ErrBadArguments)
		return core.XKCDInfo{}, core.ErrBadArguments
	}
	adress := fmt.Sprintf("%s/%d/info.0.json", c.url, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, adress, nil)
	if err != nil {
		c.log.Error("create request to xkcd", "id", id, "err", err)
		return core.XKCDInfo{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("can't get comic info", "id", id, "err", err)
		return core.XKCDInfo{}, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Error("cant close resp.Body on xkcd", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("xkcd returned non-200 status", "status", resp.StatusCode)
		return core.XKCDInfo{}, fmt.Errorf("xkcd returned status %d", resp.StatusCode)
	}

	var ans InformationPicture
	if err := json.NewDecoder(resp.Body).Decode(&ans); err != nil {
		c.log.Error("cant decode in Get", "err", err)
		return core.XKCDInfo{}, err
	}

	description := fmt.Sprintf("%s %s %s %s", ans.SafeTitle, ans.Title, ans.Alt, ans.Transcript)

	return core.XKCDInfo{ID: ans.ID, URL: ans.ImageURL, Title: ans.Title, Description: description}, nil
}

func (c *Client) LastID(ctx context.Context) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/info.0.json", c.url), nil)

	if err != nil {
		c.log.Error("cant create request lastid", "err", err)
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		c.log.Error("can't get last id", "err", err)
		return 0, fmt.Errorf("do request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.log.Error("cant close resp.Body on xkcd", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.log.Error("xkcd returned non-200 status", "status", resp.StatusCode)
		return 0, fmt.Errorf("xkcd returned status %d", resp.StatusCode)
	}

	var ans NumLastId

	if err := json.NewDecoder(resp.Body).Decode(&ans); err != nil {
		c.log.Error("cant decode in lastId", "err", err)
		return 0, err
	}

	return ans.NumsComics, nil
}
