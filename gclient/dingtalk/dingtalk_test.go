package alert

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// 测试辅助
// ============================================================================

// mockServer 创建模拟服务器
func mockServer(handler http.HandlerFunc) (*httptest.Server, func()) {
	server := httptest.NewServer(handler)
	return server, server.Close
}

// successHandler 成功响应
func successHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": 0, "errmsg": "ok"})
	}
}

// errorHandler 错误响应
func errorHandler(code int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": code, "errmsg": msg})
	}
}

// createTestRobot 创建测试机器人
func createTestRobot(serverURL string, opts ...Option) *Robot {
	host := strings.TrimPrefix(serverURL, "http://")
	allOpts := append([]Option{WithAccessToken("test_token"), WithHost(host), WithRetry(0)}, opts...)
	robot := NewRobot(allOpts...)
	robot.webhookURL = strings.Replace(robot.webhookURL, "https://", "http://", 1)
	return robot
}

// ============================================================================
// Robot 构造测试
// ============================================================================

func TestNewRobot(t *testing.T) {
	t.Run("默认配置", func(t *testing.T) {
		robot := NewRobot(WithAccessToken("my_token"))

		assert.Contains(t, robot.webhookURL, "my_token")
		assert.Contains(t, robot.webhookURL, DefaultHost)
		assert.Equal(t, DefaultTimeout, robot.timeout)
		assert.Equal(t, DefaultRetryCount, robot.retryCount)
	})

	t.Run("完整配置", func(t *testing.T) {
		robot := NewRobot(
			WithAccessToken("token"),
			WithSignSecret("secret"),
			WithHost("custom.host"),
			WithTimeout(30*time.Second),
			WithRetry(5),
		)

		assert.Contains(t, robot.webhookURL, "custom.host")
		assert.Equal(t, "secret", robot.signSecret)
		assert.Equal(t, 30*time.Second, robot.timeout)
		assert.Equal(t, 5, robot.retryCount)
	})

	t.Run("自定义HTTP客户端", func(t *testing.T) {
		client := &http.Client{Timeout: 60 * time.Second}
		robot := NewRobot(WithAccessToken("token"), WithHTTPClient(client))
		assert.Equal(t, client, robot.httpClient)
	})
}

func TestNewRobotFromEnv(t *testing.T) {
	defer func() {
		_ = os.Unsetenv("DINGTALK_ACCESS_TOKEN_TEST")
		_ = os.Unsetenv("DINGTALK_SECRET_TEST")
	}()

	t.Run("环境变量完整", func(t *testing.T) {
		_ = os.Setenv("DINGTALK_ACCESS_TOKEN_TEST", "token")
		_ = os.Setenv("DINGTALK_SECRET_TEST", "secret")

		robot, err := NewRobotFromEnv("TEST")

		require.NoError(t, err)
		assert.Contains(t, robot.webhookURL, "token")
		assert.Equal(t, "secret", robot.signSecret)
	})

	t.Run("缺少TOKEN", func(t *testing.T) {
		_ = os.Unsetenv("DINGTALK_ACCESS_TOKEN_MISS")
		_ = os.Setenv("DINGTALK_SECRET_MISS", "secret")

		_, err := NewRobotFromEnv("MISS")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ACCESS_TOKEN")
	})

	t.Run("缺少SECRET", func(t *testing.T) {
		_ = os.Setenv("DINGTALK_ACCESS_TOKEN_NOSEC", "token")
		_ = os.Unsetenv("DINGTALK_SECRET_NOSEC")

		_, err := NewRobotFromEnv("NOSEC")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "SECRET")
	})

	t.Run("自动转大写", func(t *testing.T) {
		_ = os.Setenv("DINGTALK_ACCESS_TOKEN_CASE", "token")
		_ = os.Setenv("DINGTALK_SECRET_CASE", "secret")

		robot, err := NewRobotFromEnv("case")
		require.NoError(t, err)
		assert.NotNil(t, robot)
	})
}

// ============================================================================
// Error 测试
// ============================================================================

func TestError(t *testing.T) {
	t.Run("格式化输出", func(t *testing.T) {
		err := &Error{Code: 400101, Message: "access_token不存在"}
		assert.Equal(t, "钉钉API错误 [400101]: access_token不存在", err.Error())
	})

	t.Run("IsError判断", func(t *testing.T) {
		var err error = &Error{Code: 123, Message: "test"}
		dingErr, ok := IsError(err)
		assert.True(t, ok)
		assert.Equal(t, 123, dingErr.Code)

		_, ok = IsError(assert.AnError)
		assert.False(t, ok)
	})
}

// ============================================================================
// 文本消息测试
// ============================================================================

func TestTextMessage(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*TextBuilder)
		validate func(*testing.T, map[string]any)
	}{
		{
			name: "基础消息",
			validate: func(t *testing.T, msg map[string]any) {
				assert.Equal(t, "text", msg["msgtype"])
				assert.Equal(t, "测试内容", msg["text"].(map[string]any)["content"])
			},
		},
		{
			name:  "@所有人",
			setup: func(b *TextBuilder) { b.AtAll() },
			validate: func(t *testing.T, msg map[string]any) {
				assert.True(t, msg["at"].(map[string]any)["isAtAll"].(bool))
			},
		},
		{
			name:  "@手机号",
			setup: func(b *TextBuilder) { b.AtMobiles("138", "139") },
			validate: func(t *testing.T, msg map[string]any) {
				mobiles := msg["at"].(map[string]any)["atMobiles"].([]any)
				assert.Len(t, mobiles, 2)
			},
		},
		{
			name:  "@用户ID",
			setup: func(b *TextBuilder) { b.AtUserIds("u1", "u2") },
			validate: func(t *testing.T, msg map[string]any) {
				ids := msg["at"].(map[string]any)["atUserIds"].([]any)
				assert.Len(t, ids, 2)
			},
		},
		{
			name: "组合@",
			setup: func(b *TextBuilder) {
				b.AtMobiles("138").AtUserIds("u1").AtAll()
			},
			validate: func(t *testing.T, msg map[string]any) {
				at := msg["at"].(map[string]any)
				assert.True(t, at["isAtAll"].(bool))
				assert.NotEmpty(t, at["atMobiles"])
				assert.NotEmpty(t, at["atUserIds"])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var received map[string]any
			server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(body, &received)
				successHandler()(w, r)
			})
			defer cleanup()

			robot := createTestRobot(server.URL)
			builder := robot.Text("测试内容")
			if tc.setup != nil {
				tc.setup(builder)
			}

			require.NoError(t, builder.Send())
			tc.validate(t, received)
		})
	}
}

// ============================================================================
// 链接消息测试
// ============================================================================

func TestLinkMessage(t *testing.T) {
	t.Run("基础链接", func(t *testing.T) {
		var received map[string]any
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &received)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.Link("标题", "描述", "https://example.com").Send()

		require.NoError(t, err)
		assert.Equal(t, "link", received["msgtype"])
		link := received["link"].(map[string]any)
		assert.Equal(t, "标题", link["title"])
		assert.Equal(t, "描述", link["text"])
		assert.Equal(t, "https://example.com", link["messageUrl"])
	})

	t.Run("带图片", func(t *testing.T) {
		var received map[string]any
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &received)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.Link("标题", "内容", "https://url").WithPicture("https://pic").Send()

		require.NoError(t, err)
		assert.Equal(t, "https://pic", received["link"].(map[string]any)["picUrl"])
	})
}

// ============================================================================
// Markdown消息测试
// ============================================================================

func TestMarkdownMessage(t *testing.T) {
	var received map[string]any
	server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		successHandler()(w, r)
	})
	defer cleanup()

	robot := createTestRobot(server.URL)
	err := robot.Markdown("告警", "## 详情\n- 状态: **异常**").AtMobiles("138").Send()

	require.NoError(t, err)
	assert.Equal(t, "markdown", received["msgtype"])
	md := received["markdown"].(map[string]any)
	assert.Equal(t, "告警", md["title"])
	assert.Contains(t, md["text"], "## 详情")
}

// ============================================================================
// ActionCard消息测试
// ============================================================================

func TestActionCardMessage(t *testing.T) {
	t.Run("单按钮", func(t *testing.T) {
		var received map[string]any
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &received)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.ActionCard("确认", "是否发布?").SingleButton("确认", "https://url").Send()

		require.NoError(t, err)
		card := received["actionCard"].(map[string]any)
		assert.Equal(t, "确认", card["singleTitle"])
		assert.Equal(t, "https://url", card["singleURL"])
	})

	t.Run("多按钮横向", func(t *testing.T) {
		var received map[string]any
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &received)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.ActionCard("审批", "申请休假").
			AddButton("同意", "https://yes").
			AddButton("拒绝", "https://no").
			Horizontal().
			Send()

		require.NoError(t, err)
		card := received["actionCard"].(map[string]any)
		assert.Equal(t, "1", card["btnOrientation"])
		btns := card["btns"].([]any)
		assert.Len(t, btns, 2)
	})

	t.Run("竖直排列", func(t *testing.T) {
		var received map[string]any
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &received)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.ActionCard("标题", "内容").AddButton("按钮", "url").Vertical().Send()

		require.NoError(t, err)
		assert.Equal(t, "0", received["actionCard"].(map[string]any)["btnOrientation"])
	})
}

// ============================================================================
// FeedCard消息测试
// ============================================================================

func TestFeedCardMessage(t *testing.T) {
	var received map[string]any
	server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		successHandler()(w, r)
	})
	defer cleanup()

	robot := createTestRobot(server.URL)
	err := robot.FeedCard().
		AddLink("新闻1", "https://1", "https://pic1").
		AddLink("新闻2", "https://2", "https://pic2").
		Send()

	require.NoError(t, err)
	assert.Equal(t, "feedCard", received["msgtype"])
	links := received["feedCard"].(map[string]any)["links"].([]any)
	assert.Len(t, links, 2)
	assert.Equal(t, "新闻1", links[0].(map[string]any)["title"])
}

// ============================================================================
// 签名测试
// ============================================================================

func TestSignature(t *testing.T) {
	t.Run("有密钥时包含签名", func(t *testing.T) {
		var requestURL string
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			requestURL = r.URL.String()
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL, WithSignSecret("secret"))
		_ = robot.Text("测试").Send()

		assert.Contains(t, requestURL, "timestamp=")
		assert.Contains(t, requestURL, "sign=")
	})

	t.Run("无密钥时无签名", func(t *testing.T) {
		var requestURL string
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			requestURL = r.URL.String()
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		_ = robot.Text("测试").Send()

		assert.NotContains(t, requestURL, "timestamp=")
	})

	t.Run("签名一致性", func(t *testing.T) {
		robot := NewRobot(WithSignSecret("SECtest123"))
		timestamp := int64(1609459200000)

		sign1, _ := robot.calculateSign(timestamp)
		sign2, _ := robot.calculateSign(timestamp)

		assert.Equal(t, sign1, sign2)
		assert.NotEmpty(t, sign1)
	})
}

// ============================================================================
// 错误处理测试
// ============================================================================

func TestErrorHandling(t *testing.T) {
	t.Run("API错误", func(t *testing.T) {
		server, cleanup := mockServer(errorHandler(400101, "token无效"))
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.Text("测试").Send()

		require.Error(t, err)
		dingErr, ok := IsError(err)
		require.True(t, ok)
		assert.Equal(t, 400101, dingErr.Code)
	})

	t.Run("HTTP错误", func(t *testing.T) {
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		err := robot.Text("测试").Send()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "500")
	})

	t.Run("连接失败", func(t *testing.T) {
		robot := NewRobot(WithAccessToken("token"), WithHost("localhost:1"), WithRetry(0), WithTimeout(time.Second))
		err := robot.Text("测试").Send()
		require.Error(t, err)
	})
}

// ============================================================================
// 重试测试
// ============================================================================

func TestRetry(t *testing.T) {
	t.Run("网络错误重试成功", func(t *testing.T) {
		var count int32
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&count, 1) < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL, WithRetry(3))
		robot.retryInterval = 10 * time.Millisecond

		require.NoError(t, robot.Text("测试").Send())
		assert.Equal(t, int32(3), atomic.LoadInt32(&count))
	})

	t.Run("API错误不重试", func(t *testing.T) {
		var count int32
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&count, 1)
			errorHandler(400101, "错误")(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL, WithRetry(3))
		err := robot.Text("测试").Send()

		require.Error(t, err)
		assert.Equal(t, int32(1), atomic.LoadInt32(&count))
	})
}

// ============================================================================
// Context支持测试
// ============================================================================

func TestContext(t *testing.T) {
	t.Run("Context取消", func(t *testing.T) {
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := robot.Text("测试").SendWithContext(ctx)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Context正常", func(t *testing.T) {
		server, cleanup := mockServer(successHandler())
		defer cleanup()

		robot := createTestRobot(server.URL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		require.NoError(t, robot.Text("测试").SendWithContext(ctx))
	})
}

// ============================================================================
// HTTP请求测试
// ============================================================================

func TestHTTPRequest(t *testing.T) {
	t.Run("请求头", func(t *testing.T) {
		var headers http.Header
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			headers = r.Header.Clone()
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		_ = robot.Text("测试").Send()

		assert.Equal(t, "application/json; charset=utf-8", headers.Get("Content-Type"))
		assert.Equal(t, UserAgent, headers.Get("User-Agent"))
	})

	t.Run("POST方法", func(t *testing.T) {
		var method string
		server, cleanup := mockServer(func(w http.ResponseWriter, r *http.Request) {
			method = r.Method
			successHandler()(w, r)
		})
		defer cleanup()

		robot := createTestRobot(server.URL)
		_ = robot.Text("测试").Send()

		assert.Equal(t, http.MethodPost, method)
	})
}

// ============================================================================
// 示例
// ============================================================================

func ExampleNewRobot() {
	robot := NewRobot(
		WithAccessToken("your_access_token"),
		WithSignSecret("your_sign_secret"),
	)

	_ = robot.Text("服务器告警：CPU使用率超过90%").AtAll().Send()
}

func ExampleRobot_Markdown() {
	robot := NewRobot(WithAccessToken("your_token"))

	content := `## 部署通知
- **环境**: 生产环境
- **版本**: v2.0.0
- **状态**: ✅ 成功`

	_ = robot.Markdown("部署通知", content).AtMobiles("13800138000").Send()
}

func ExampleRobot_ActionCard() {
	robot := NewRobot(WithAccessToken("your_token"))

	// 单按钮
	_ = robot.ActionCard("确认", "是否发布?").SingleButton("确认发布", "https://example.com").Send()

	// 多按钮
	_ = robot.ActionCard("审批", "张三申请休假").
		AddButton("同意", "https://yes").
		AddButton("拒绝", "https://no").
		Horizontal().
		Send()
}
