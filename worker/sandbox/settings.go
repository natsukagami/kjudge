package sandbox

type Settings struct {
	LogSandbox    bool
	IgnoreWarning bool
}

var DefaultSettings = Settings{LogSandbox: true, IgnoreWarning: false}

type Option func(Settings) Settings

func IgnoreWarnings(ignore bool) Option {
	return func(o Settings) Settings {
		o.IgnoreWarning = ignore
		return o
	}
}

func EnableSandboxLogs(enable bool) Option {
	return func(o Settings) Settings {
		o.LogSandbox = enable
		return o
	}
}
