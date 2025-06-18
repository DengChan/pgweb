package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// 全局Logger实例
	globalLogger *logrus.Logger
)

// Config Logger配置
type Config struct {
	LogDir     string // 日志目录
	FileName   string // 日志文件名
	MaxSize    int    // 单个日志文件最大大小(MB)
	MaxBackups int    // 保留的备份文件数量
	MaxAge     int    // 保留的天数
	Compress   bool   // 是否压缩旧文件
	LocalTime  bool   // 是否使用本地时间
}

// DefaultConfig 默认日志配置
var DefaultConfig = Config{
	LogDir:     "data/logs",
	FileName:   "pgweb.log",
	MaxSize:    100,   // 100MB
	MaxBackups: 0,     // 不限制备份数量，由MaxAge控制
	MaxAge:     30,    // 保留30天
	Compress:   false, // 不压缩，方便查看
	LocalTime:  true,  // 使用本地时间
}

var loggerOnce sync.Once

// Init 初始化全局Logger
func Init(config Config) error {
	var err error
	loggerOnce.Do(func() {
		// 创建日志目录
		if err = os.MkdirAll(config.LogDir, 0755); err != nil && !os.IsExist(err) {
			fmt.Println("[ERROR] create log dir failed", err.Error())
			return
		}

		// 创建lumberjack日志滚动器
		lumberjackLogger := &lumberjack.Logger{
			Filename:   filepath.Join(config.LogDir, config.FileName),
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
			LocalTime:  config.LocalTime,
		}

		// 创建全局Logger
		globalLogger = logrus.New()

		// 设置JSON格式化器，便于日志解析
		globalLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})

		// 设置同时输出到文件和控制台
		multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
		globalLogger.SetOutput(multiWriter)

		// 设置日志级别
		globalLogger.SetLevel(logrus.InfoLevel)

		// 记录初始化成功日志
		globalLogger.WithFields(logrus.Fields{
			"log_dir":     config.LogDir,
			"file_name":   config.FileName,
			"max_size":    config.MaxSize,
			"max_backups": config.MaxBackups,
			"max_age":     config.MaxAge,
			"compress":    config.Compress,
			"local_time":  config.LocalTime,
		}).Info("Global logger initialized successfully")
	})

	return err
}

// Logger 获取全局Logger实例 - 提供简洁的使用方式
func Logger() *logrus.Logger {
	if globalLogger == nil {
		// 如果全局Logger未初始化，使用默认配置初始化
		if err := Init(DefaultConfig); err != nil {
			// 如果初始化失败，创建一个基本的logger
			globalLogger = logrus.New()
			globalLogger.SetFormatter(&logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05",
			})
			globalLogger.SetOutput(os.Stdout)
			globalLogger.SetLevel(logrus.InfoLevel)
			globalLogger.WithError(err).Error("Failed to initialize global logger with file output, fallback to stdout")
		}
	}
	return globalLogger
}

// GetGlobalLogger 获取全局Logger实例 - 保持向后兼容
func GlobalLogger() *logrus.Logger {
	return Logger()
}

// SetLevel 设置日志级别
func SetLevel(level logrus.Level) {
	Logger().SetLevel(level)
}

// SetFormatter 设置日志格式化器
func SetFormatter(formatter logrus.Formatter) {
	Logger().SetFormatter(formatter)
}

// Info 记录Info级别日志
func Info(args ...interface{}) {
	Logger().Info(args...)
}
func Infof(format string, args ...interface{}) {
	Logger().Infof(format, args...)
}

// Debug 记录Debug级别日志
func Debug(args ...interface{}) {
	Logger().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Logger().Debugf(format, args...)
}

// Warn 记录Warn级别日志
func Warn(args ...interface{}) {
	Logger().Warn(args...)
}
func Warnf(format string, args ...interface{}) {
	Logger().Warnf(format, args...)
}

// Error 记录Error级别日志
func Error(args ...interface{}) {
	Logger().Error(args...)
}
func Errorf(format string, args ...interface{}) {
	Logger().Errorf(format, args...)
}

// Fatal 记录Fatal级别日志
func Fatal(args ...interface{}) {
	Logger().Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	Logger().Fatalf(format, args...)
}

// WithField 添加字段
func WithField(key string, value interface{}) *logrus.Entry {
	return Logger().WithField(key, value)
}

// WithFields 添加多个字段
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger().WithFields(fields)
}

// WithError 添加错误字段
func WithError(err error) *logrus.Entry {
	return Logger().WithError(err)
}

func init() {
	// 在包初始化时尝试初始化全局Logger
	if err := Init(DefaultConfig); err != nil {
		fmt.Println("init logger failed", err.Error(), "use basic stdout logger.")
		// 如果初始化失败，创建基本logger
		globalLogger = logrus.New()
		globalLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		globalLogger.SetOutput(os.Stdout)
		globalLogger.SetLevel(logrus.InfoLevel)
	}
}
