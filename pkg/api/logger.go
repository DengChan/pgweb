package api

import (
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/sosedoff/pgweb/pkg/command"
	"github.com/sosedoff/pgweb/pkg/logger"
)

var (
	reConnectToken = regexp.MustCompile("/connect/(.*)")
)

func RequestLogger(loggerInstance *logrus.Logger) gin.HandlerFunc {
	// 如果传入nil，使用全局logger
	if loggerInstance == nil {
		loggerInstance = logger.Logger()
	}

	debug := loggerInstance.Level > logrus.InfoLevel
	logForwardedUser := command.Opts.LogForwardedUser

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		if !debug {
			// Skip static assets logging
			if strings.Contains(path, "/static/") {
				return
			}

			path = sanitizeLogPath(path)
		}

		status := c.Writer.Status()
		end := time.Now()
		latency := end.Sub(start)

		fields := logrus.Fields{
			"status":      status,
			"method":      c.Request.Method,
			"remote_addr": c.ClientIP(),
			"duration":    latency.String(),
			"duration_ms": latency.Milliseconds(),
			"path":        path,
		}

		if reqID := getRequestID(c); reqID != "" {
			fields["id"] = reqID
		}

		if logForwardedUser {
			if forwardedUser := c.GetHeader("X-Forwarded-User"); forwardedUser != "" {
				fields["forwarded_user"] = forwardedUser
			}
			if forwardedEmail := c.GetHeader("X-Forwarded-Email"); forwardedEmail != "" {
				fields["forwarded_email"] = forwardedEmail
			}
		}

		if err := c.Errors.Last(); err != nil {
			fields["error"] = err.Error()
		}

		// Additional fields for debugging
		if debug {
			fields["raw_query"] = c.Request.URL.RawQuery

			if c.Request.Method != http.MethodGet {
				fields["raw_form"] = c.Request.Form
			}
		}

		msg := "http_request"

		switch {
		case status >= http.StatusBadRequest && status < http.StatusInternalServerError:
			loggerInstance.WithFields(fields).Warn(msg)
		case status >= http.StatusInternalServerError:
			loggerInstance.WithFields(fields).Error(msg)
		default:
			loggerInstance.WithFields(fields).Info(msg)
		}
	}
}

func sanitizeLogPath(str string) string {
	return reConnectToken.ReplaceAllString(str, "/connect/REDACTED")
}

func getRequestID(c *gin.Context) string {
	id := c.GetHeader("x-request-id")
	if id == "" {
		id = c.GetHeader("x-amzn-trace-id")
	}
	return id
}
