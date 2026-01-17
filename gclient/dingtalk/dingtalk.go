// Package alert 提供钉钉机器人消息推送功能
//
// 本包实现了钉钉群机器人的完整消息类型支持，包括：
//   - 文本消息 (Text)
//   - 链接消息 (Link)
//   - Markdown消息
//   - ActionCard消息（单按钮/多按钮）
//   - FeedCard消息
//
// # 快速开始
//
//	// 创建机器人客户端
//	robot := alert.NewRobot(
//	    alert.WithAccessToken("your_access_token"),
//	    alert.WithSignSecret("your_secret"),
//	)
//
//	// 发送文本消息
//	robot.Text("服务器异常告警！").AtAll().Send()
//
//	// 发送Markdown消息
//	robot.Markdown("告警通知", "## 异常详情\n- 时间: 2024-01-01").
//	    AtMobiles("13800138000").
//	    Send()
//
// # 从环境变量创建
//
//	// 设置环境变量: DINGTALK_ACCESS_TOKEN_P0, DINGTALK_SECRET_P0
//	robot, err := alert.NewRobotFromEnv("P0")
package alert

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// 常量定义
// ============================================================================

const (
	// DefaultHost 钉钉机器人API默认主机地址
	DefaultHost = "oapi.dingtalk.com"

	// DefaultTimeout 默认HTTP请求超时时间
	DefaultTimeout = 10 * time.Second

	// DefaultRetryCount 默认重试次数
	DefaultRetryCount = 3

	// DefaultRetryInterval 默认重试间隔基数
	DefaultRetryInterval = time.Second

	// UserAgent 请求的User-Agent标识
	UserAgent = "DingTalk-Robot-SDK/2.0"
)

// msgType 消息类型枚举
type msgType string

const (
	msgTypeText       msgType = "text"
	msgTypeLink       msgType = "link"
	msgTypeMarkdown   msgType = "markdown"
	msgTypeActionCard msgType = "actionCard"
	msgTypeFeedCard   msgType = "feedCard"
)

// ============================================================================
// 错误定义
// ============================================================================

// Error 钉钉API返回的错误信息
type Error struct {
	Code    int    // 钉钉错误码
	Message string // 错误描述
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("钉钉API错误 [%d]: %s", e.Code, e.Message)
}

// IsError 判断 error 是否为钉钉API错误
func IsError(err error) (*Error, bool) {
	if e, ok := err.(*Error); ok {
		return e, true
	}
	return nil, false
}

// ============================================================================
// 机器人配置选项
// ============================================================================

// Option 机器人配置选项函数
type Option func(*Robot)

// WithAccessToken 设置机器人的 access_token（必需）
func WithAccessToken(token string) Option {
	return func(r *Robot) { r.accessToken = token }
}

// WithSignSecret 设置签名密钥（加签模式必需）
func WithSignSecret(secret string) Option {
	return func(r *Robot) { r.signSecret = secret }
}

// WithHost 设置自定义API主机地址
func WithHost(host string) Option {
	return func(r *Robot) { r.host = host }
}

// WithTimeout 设置HTTP请求超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(r *Robot) { r.timeout = timeout }
}

// WithRetry 设置重试次数
func WithRetry(count int) Option {
	return func(r *Robot) { r.retryCount = count }
}

// WithHTTPClient 设置自定义HTTP客户端
func WithHTTPClient(client *http.Client) Option {
	return func(r *Robot) { r.httpClient = client }
}

// ============================================================================
// Robot 机器人客户端
// ============================================================================

// Robot 钉钉群机器人客户端
type Robot struct {
	accessToken   string        // 访问令牌
	signSecret    string        // 签名密钥
	host          string        // API主机
	httpClient    *http.Client  // HTTP客户端
	timeout       time.Duration // 超时时间
	retryCount    int           // 重试次数
	retryInterval time.Duration // 重试间隔
	webhookURL    string        // Webhook地址
}

// NewRobot 创建钉钉机器人客户端
//
// 示例：
//
//	robot := NewRobot(
//	    WithAccessToken("your_token"),
//	    WithSignSecret("your_secret"),
//	)
func NewRobot(opts ...Option) *Robot {
	r := &Robot{
		host:          DefaultHost,
		timeout:       DefaultTimeout,
		retryCount:    DefaultRetryCount,
		retryInterval: DefaultRetryInterval,
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.httpClient == nil {
		r.httpClient = &http.Client{Timeout: r.timeout}
	}

	r.webhookURL = fmt.Sprintf("https://%s/robot/send?access_token=%s", r.host, r.accessToken)
	return r
}

// NewRobotFromEnv 从环境变量创建机器人
//
// 环境变量：
//   - DINGTALK_ACCESS_TOKEN_{level}
//   - DINGTALK_SECRET_{level}
//
// 示例：
//
//	robot, err := NewRobotFromEnv("P0")
func NewRobotFromEnv(level string) (*Robot, error) {
	level = strings.ToUpper(level)

	tokenKey := fmt.Sprintf("DINGTALK_ACCESS_TOKEN_%s", level)
	secretKey := fmt.Sprintf("DINGTALK_SECRET_%s", level)

	token := os.Getenv(tokenKey)
	if token == "" {
		return nil, fmt.Errorf("环境变量 %s 未设置", tokenKey)
	}

	secret := os.Getenv(secretKey)
	if secret == "" {
		return nil, fmt.Errorf("环境变量 %s 未设置", secretKey)
	}

	return NewRobot(WithAccessToken(token), WithSignSecret(secret)), nil
}

// ============================================================================
// @功能
// ============================================================================

// AtInfo @功能配置
type AtInfo struct {
	AtMobiles []string `json:"atMobiles,omitempty"` // 被@人的手机号
	AtUserIds []string `json:"atUserIds,omitempty"` // 被@人的用户ID
	IsAtAll   bool     `json:"isAtAll,omitempty"`   // 是否@所有人
}

// ============================================================================
// 文本消息
// ============================================================================

// TextBuilder 文本消息构建器
type TextBuilder struct {
	robot   *Robot
	content string
	at      *AtInfo
}

// Text 创建文本消息
//
// 示例：
//
//	robot.Text("告警：CPU使用率90%").AtAll().Send()
func (r *Robot) Text(content string) *TextBuilder {
	return &TextBuilder{robot: r, content: content, at: &AtInfo{}}
}

// AtMobiles 通过手机号@指定用户
func (b *TextBuilder) AtMobiles(mobiles ...string) *TextBuilder {
	b.at.AtMobiles = append(b.at.AtMobiles, mobiles...)
	return b
}

// AtUserIds 通过用户ID@指定用户
func (b *TextBuilder) AtUserIds(userIds ...string) *TextBuilder {
	b.at.AtUserIds = append(b.at.AtUserIds, userIds...)
	return b
}

// AtAll @所有人
func (b *TextBuilder) AtAll() *TextBuilder {
	b.at.IsAtAll = true
	return b
}

// Send 发送消息
func (b *TextBuilder) Send() error {
	return b.SendWithContext(context.Background())
}

// SendWithContext 使用指定 Context 发送
func (b *TextBuilder) SendWithContext(ctx context.Context) error {
	return b.robot.send(ctx, map[string]any{
		"msgtype": msgTypeText,
		"text":    map[string]string{"content": b.content},
		"at":      b.at,
	})
}

// ============================================================================
// 链接消息
// ============================================================================

// LinkBuilder 链接消息构建器
type LinkBuilder struct {
	robot      *Robot
	title      string
	text       string
	messageURL string
	picURL     string
}

// Link 创建链接消息
//
// 示例：
//
//	robot.Link("版本发布", "v2.0已发布", "https://github.com/xxx").
//	    WithPicture("https://xxx/logo.png").
//	    Send()
func (r *Robot) Link(title, text, messageURL string) *LinkBuilder {
	return &LinkBuilder{robot: r, title: title, text: text, messageURL: messageURL}
}

// WithPicture 设置消息配图
func (b *LinkBuilder) WithPicture(picURL string) *LinkBuilder {
	b.picURL = picURL
	return b
}

// Send 发送消息
func (b *LinkBuilder) Send() error {
	return b.SendWithContext(context.Background())
}

// SendWithContext 使用指定 Context 发送
func (b *LinkBuilder) SendWithContext(ctx context.Context) error {
	link := map[string]string{
		"title":      b.title,
		"text":       b.text,
		"messageUrl": b.messageURL,
	}
	if b.picURL != "" {
		link["picUrl"] = b.picURL
	}
	return b.robot.send(ctx, map[string]any{"msgtype": msgTypeLink, "link": link})
}

// ============================================================================
// Markdown消息
// ============================================================================

// MarkdownBuilder Markdown消息构建器
type MarkdownBuilder struct {
	robot *Robot
	title string
	text  string
	at    *AtInfo
}

// Markdown 创建Markdown消息
//
// 支持的语法：# 标题、> 引用、**粗体**、[链接](URL)、![图片](URL)、- 列表
//
// 示例：
//
//	robot.Markdown("告警", "## 详情\n- 服务: api\n- 状态: **异常**").
//	    AtMobiles("13800138000").
//	    Send()
func (r *Robot) Markdown(title, text string) *MarkdownBuilder {
	return &MarkdownBuilder{robot: r, title: title, text: text, at: &AtInfo{}}
}

// AtMobiles 通过手机号@指定用户
func (b *MarkdownBuilder) AtMobiles(mobiles ...string) *MarkdownBuilder {
	b.at.AtMobiles = append(b.at.AtMobiles, mobiles...)
	return b
}

// AtUserIds 通过用户ID@指定用户
func (b *MarkdownBuilder) AtUserIds(userIds ...string) *MarkdownBuilder {
	b.at.AtUserIds = append(b.at.AtUserIds, userIds...)
	return b
}

// AtAll @所有人
func (b *MarkdownBuilder) AtAll() *MarkdownBuilder {
	b.at.IsAtAll = true
	return b
}

// Send 发送消息
func (b *MarkdownBuilder) Send() error {
	return b.SendWithContext(context.Background())
}

// SendWithContext 使用指定 Context 发送
func (b *MarkdownBuilder) SendWithContext(ctx context.Context) error {
	return b.robot.send(ctx, map[string]any{
		"msgtype":  msgTypeMarkdown,
		"markdown": map[string]string{"title": b.title, "text": b.text},
		"at":       b.at,
	})
}

// ============================================================================
// ActionCard消息
// ============================================================================

// ActionCardBuilder ActionCard消息构建器
type ActionCardBuilder struct {
	robot          *Robot
	title          string
	text           string
	btnOrientation string // "0": 竖直, "1": 横向
	singleTitle    string
	singleURL      string
	buttons        []Button
}

// Button ActionCard按钮
type Button struct {
	Title     string `json:"title"`
	ActionURL string `json:"actionURL"`
}

// ActionCard 创建ActionCard消息
//
// 示例（单按钮）：
//
//	robot.ActionCard("确认", "是否发布?").SingleButton("确认", "https://xxx").Send()
//
// 示例（多按钮）：
//
//	robot.ActionCard("审批", "张三申请休假").
//	    AddButton("同意", "https://xxx/yes").
//	    AddButton("拒绝", "https://xxx/no").
//	    Horizontal().
//	    Send()
func (r *Robot) ActionCard(title, text string) *ActionCardBuilder {
	return &ActionCardBuilder{
		robot:          r,
		title:          title,
		text:           text,
		btnOrientation: "0",
		buttons:        make([]Button, 0),
	}
}

// SingleButton 设置单按钮（整个卡片可点击）
func (b *ActionCardBuilder) SingleButton(title, actionURL string) *ActionCardBuilder {
	b.singleTitle = title
	b.singleURL = actionURL
	return b
}

// AddButton 添加按钮（多按钮模式）
func (b *ActionCardBuilder) AddButton(title, actionURL string) *ActionCardBuilder {
	b.buttons = append(b.buttons, Button{Title: title, ActionURL: actionURL})
	return b
}

// Horizontal 按钮横向排列
func (b *ActionCardBuilder) Horizontal() *ActionCardBuilder {
	b.btnOrientation = "1"
	return b
}

// Vertical 按钮竖直排列（默认）
func (b *ActionCardBuilder) Vertical() *ActionCardBuilder {
	b.btnOrientation = "0"
	return b
}

// Send 发送消息
func (b *ActionCardBuilder) Send() error {
	return b.SendWithContext(context.Background())
}

// SendWithContext 使用指定 Context 发送
func (b *ActionCardBuilder) SendWithContext(ctx context.Context) error {
	card := map[string]any{
		"title":          b.title,
		"text":           b.text,
		"btnOrientation": b.btnOrientation,
	}

	if b.singleTitle != "" {
		card["singleTitle"] = b.singleTitle
		card["singleURL"] = b.singleURL
	} else if len(b.buttons) > 0 {
		card["btns"] = b.buttons
	}

	return b.robot.send(ctx, map[string]any{"msgtype": msgTypeActionCard, "actionCard": card})
}

// ============================================================================
// FeedCard消息
// ============================================================================

// FeedCardBuilder FeedCard消息构建器
type FeedCardBuilder struct {
	robot *Robot
	links []FeedLink
}

// FeedLink FeedCard链接项
type FeedLink struct {
	Title      string `json:"title"`
	MessageURL string `json:"messageURL"`
	PicURL     string `json:"picURL"`
}

// FeedCard 创建FeedCard消息
//
// 示例：
//
//	robot.FeedCard().
//	    AddLink("新闻1", "https://xxx/1", "https://xxx/pic1.jpg").
//	    AddLink("新闻2", "https://xxx/2", "https://xxx/pic2.jpg").
//	    Send()
func (r *Robot) FeedCard() *FeedCardBuilder {
	return &FeedCardBuilder{robot: r, links: make([]FeedLink, 0)}
}

// AddLink 添加链接
func (b *FeedCardBuilder) AddLink(title, messageURL, picURL string) *FeedCardBuilder {
	b.links = append(b.links, FeedLink{Title: title, MessageURL: messageURL, PicURL: picURL})
	return b
}

// Send 发送消息
func (b *FeedCardBuilder) Send() error {
	return b.SendWithContext(context.Background())
}

// SendWithContext 使用指定 Context 发送
func (b *FeedCardBuilder) SendWithContext(ctx context.Context) error {
	return b.robot.send(ctx, map[string]any{
		"msgtype":  msgTypeFeedCard,
		"feedCard": map[string]any{"links": b.links},
	})
}

// ============================================================================
// 核心发送逻辑
// ============================================================================

// send 发送消息（自动重试）
func (r *Robot) send(ctx context.Context, message any) error {
	var lastErr error

	for attempt := 0; attempt <= r.retryCount; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * r.retryInterval):
			}
		}

		if err := r.doSend(ctx, message); err == nil {
			return nil
		} else {
			lastErr = err
			// API错误不重试
			if _, isDingErr := err.(*Error); isDingErr {
				return err
			}
		}
	}

	return fmt.Errorf("发送失败，已重试 %d 次: %w", r.retryCount, lastErr)
}

// doSend 执行单次发送
func (r *Robot) doSend(ctx context.Context, message any) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}

	requestURL, err := r.buildRequestURL()
	if err != nil {
		return fmt.Errorf("构建URL失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return &Error{Code: result.ErrCode, Message: result.ErrMsg}
	}

	return nil
}

// buildRequestURL 构建请求URL（含签名）
func (r *Robot) buildRequestURL() (string, error) {
	if r.signSecret == "" {
		return r.webhookURL, nil
	}

	timestamp := time.Now().UnixMilli()
	sign, err := r.calculateSign(timestamp)
	if err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("timestamp", strconv.FormatInt(timestamp, 10))
	params.Set("sign", sign)

	return fmt.Sprintf("%s&%s", r.webhookURL, params.Encode()), nil
}

// calculateSign 计算签名
func (r *Robot) calculateSign(timestamp int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, r.signSecret)
	h := hmac.New(sha256.New, []byte(r.signSecret))
	if _, err := h.Write([]byte(stringToSign)); err != nil {
		return "", fmt.Errorf("计算签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
