package utils

import (
	"gorm.io/gorm/logger"

	"context"
	"regexp"
	"strings"
	"time"
)

type PrefixedLogger struct {
	logger.Interface
	Prefix string
}

var whitespaceRE = regexp.MustCompile(`\s+`)

func clean(msg string) string {
	return whitespaceRE.ReplaceAllString(strings.TrimSpace(msg), " ")
}

func (l PrefixedLogger) LogMode(level logger.LogLevel) logger.Interface {
	return PrefixedLogger {
		Interface: l.Interface.LogMode(level),
		Prefix: l.Prefix,
	}
}

func (l PrefixedLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Interface.Info(ctx, l.Prefix+clean(msg), data...)
}

func (l PrefixedLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Interface.Warn(ctx, l.Prefix+clean(msg), data...)
}

func (l PrefixedLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Interface.Error(ctx, l.Prefix+clean(msg), data...)
}

func (l PrefixedLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	l.Interface.Trace(ctx, begin, func() (string, int64) {
		sql, rows := fc()
		return l.Prefix + clean(sql), rows
	}, err)
}

