package queue

type Settings struct {
	LogCompile bool
	LogRun     bool
	LogScore   bool
}

var DefaultSettings = Settings{LogCompile: true, LogRun: true, LogScore: true}

type Option func(Settings) Settings

func CompileLogs(enable bool) Option {
	return func(o Settings) Settings {
		o.LogCompile = enable
		return o
	}
}

func RunLogs(enable bool) Option {
	return func(o Settings) Settings {
		o.LogRun = enable
		return o
	}
}

func ScoreLogs(enable bool) Option {
	return func(o Settings) Settings {
		o.LogScore = enable
		return o
	}
}
