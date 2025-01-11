package bakefile

type Config struct {
	Targets map[string]Target
}

type Target struct {
	Artifacts []string
	Deps      []string
	Command   string
}
