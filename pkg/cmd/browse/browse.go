package browse

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewCmdBrowse(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "browse <command>",
		Short: "",
		Long:  `Open GitHub in the browser`,
		Example: heredoc.Doc(`
			$ gh browse issue "217"
			$ gh browse file "src/main.java"
			$ gh browse branch "master"
		`),
		Annotations: map[string]string{
			"IsCore": "true", 
			"help:arguments": heredoc.Doc(`
				Branch names, pr numbers, issue numbers, and file paths
				can be supplied as arguments in the following formats:
		 		- by number, e.g. “123”; or
		 		- by description, e.g. “index.js”
		 
			`),
		},
	}

	cmdutil.EnableRepoOverride(cmd, f) 


	return cmd
}