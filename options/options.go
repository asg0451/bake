package options

type Opts struct {
	BakefilePath string `short:"p" long:"path" description:"path to Bakefile.toml" default:"Bakefile.toml"`
}
