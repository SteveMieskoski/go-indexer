package utils

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func InitializeLogger() {
	//config := zap.NewProductionEncoderConfig()
	//config.EncodeTime = zapcore.ISO8601TimeEncoder
	////fileEncoder := zapcore.NewJSONEncoder(config)
	//consoleEncoder := zapcore.NewConsoleEncoder(config)
	////logFile, _ := os.OpenFile("text.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	////writer := zapcore.AddSync(logFile)
	//defaultLogLevel := zapcore.DebugLevel
	//core := zapcore.NewTee(
	//	//zapcore.NewCore(fileEncoder, writer, defaultLogLevel),
	//	zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
	//)
	//Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	logger1, _ := zap.NewProduction()
	defer logger1.Sync() // flushes buffer, if any
	Logger = logger1.Sugar()
}

//
//func setupLogger() {
//	rawJSON := []byte(`{
//	  "level": "debug",
//	  "encoding": "json",
//	  "outputPaths": ["stdout", "/tmp/logs"],
//	  "errorOutputPaths": ["stderr"],
//	  "initialFields": {"foo": "bar"},
//	  "encoderConfig": {
//	    "messageKey": "message",
//	    "levelKey": "level",
//	    "levelEncoder": "lowercase"
//	  }
//	}`)
//
//	var cfg zap.Config
//	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
//		panic(err)
//	}
//	logger := zap.Must(cfg.Build())
//	defer logger.Sync()
//
//}
