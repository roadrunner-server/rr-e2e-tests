package logger

import (
	"context"
	"strings"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if config section exists.
	Has(name string) bool
}

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

type TestPlugin struct {
	config Configurer
	log    *zap.Logger
}

type Loggable struct{}

func (l *Loggable) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("error", "Example marshaller error")
	return nil
}

func (p1 *TestPlugin) Init(cfg Configurer, log Logger) error {
	p1.config = cfg
	p1.log = log.NamedLogger("test")
	return nil
}

func (p1 *TestPlugin) Serve() chan error {
	errCh := make(chan error, 1)
	p1.log.Error("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Info("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Debug("error", zap.Error(errors.E(errors.Str("test"))))
	p1.log.Warn("error", zap.Error(errors.E(errors.Str("test"))))

	field := zap.String("error", "Example field error")

	p1.log.Error("error", field)
	p1.log.Info("error", field)
	p1.log.Debug("error", field)
	p1.log.Warn("error", field)

	marshalledObject := &Loggable{}

	p1.log.Error("error", zap.Any("object", marshalledObject))
	p1.log.Info("error", zap.Any("object", marshalledObject))
	p1.log.Debug("error", zap.Any("object", marshalledObject))
	p1.log.Warn("error", zap.Any("object", marshalledObject))

	p1.log.Error("error", zap.String("test", ""))
	p1.log.Info("error", zap.String("test", ""))
	p1.log.Debug("error", zap.String("test", ""))
	p1.log.Warn("error", zap.String("test", ""))

	// test the `raw` mode
	messageJSON := []byte(`{"field": "value"}`)
	p1.log.Debug(strings.TrimRight(string(messageJSON), " \n\t"))

	return errCh
}

func (p1 *TestPlugin) Stop(context.Context) error {
	return nil
}

func (p1 *TestPlugin) Name() string {
	return "logger_plugin"
}
