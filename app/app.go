package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/AvengeMedia/dankgo/log"
	"github.com/AvengeMedia/dankgo/paths"
	"github.com/spf13/cobra"
)

type Info struct {
	Name      string
	ID        string
	Version   string
	Commit    string
	BuildTime string
}

type App struct {
	Info Info
	root *cobra.Command
}

func New(info Info, root *cobra.Command) *App {
	log.SetEnvPrefix(strings.ToUpper(info.ID))
	log.ApplyEnvOverrides()

	if root.Version == "" {
		root.Version = versionString(info)
	}
	return &App{Info: info, root: root}
}

func versionString(info Info) string {
	if info.Commit == "" {
		return info.Version
	}
	return fmt.Sprintf("%s (%s)", info.Version, info.Commit)
}

func (a *App) Paths() paths.App { return paths.New(a.Info.ID) }

func (a *App) Execute() {
	if err := a.root.Execute(); err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
