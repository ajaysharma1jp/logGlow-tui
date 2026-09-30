package main

type LogLevel struct{
	ShowError 	bool
	ShowInfo 	bool
	ShowWarning bool
	ShowDebug 	bool
}

func NewLogLevel() LogLevel{
	return LogLevel{
		ShowError: true,
		ShowInfo: true,
		ShowWarning: true,
		ShowDebug: true,
	}
}

func (logLevel *LogLevel) handleKeyStroke(key string) bool{
	switch key{
	case "e":
		logLevel.ShowError = true
		logLevel.ShowInfo = false
		logLevel.ShowWarning = false
		logLevel.ShowDebug = false
		return true
	case "i":
		logLevel.ShowError = false
		logLevel.ShowInfo = true
		logLevel.ShowWarning = false
		logLevel.ShowDebug = false
		return true
	case "w":
		logLevel.ShowError = false
		logLevel.ShowInfo = false
		logLevel.ShowWarning = true
		logLevel.ShowDebug = false
		return true
	case "d":
		logLevel.ShowError = false
		logLevel.ShowInfo = false
		logLevel.ShowWarning = false
		logLevel.ShowDebug = true
		return true
	case "a":
		logLevel.ShowError = true
		logLevel.ShowInfo = true
		logLevel.ShowWarning = true
		logLevel.ShowDebug = true
		return true
	}
	return false
}

func (logLevel LogLevel) IsVisible(level string) bool{
	switch level{
	case "error", "fatal", "err":
		return logLevel.ShowError
	case "info":
		return logLevel.ShowInfo
	case "warn", "warning":
		return logLevel.ShowWarning
	case "debug":
		return logLevel.ShowDebug
	default:
		return true
	}
}