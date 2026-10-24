package main

import (
	"embed"

	"github.com/srz-zumix/gh-jq/cmd"
)

//go:embed skills
var skillsFS embed.FS

func main() {
	cmd.RegisterSkillsCmd(skillsFS)
	cmd.Execute()
}
