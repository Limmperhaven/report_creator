package cmd

import (
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain"
	"log"

	"github.com/spf13/cobra"
)

var resourcesCmd = &cobra.Command{
	Use:   "resources",
	Short: "Get users by projects",
	Long:  `Using this command you can get users by projects`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("error reading flags: %s", err.Error())
		}

		log.Println("Reading configs...")
		cfg, err := config.ReadConfig(cfgPath, "")
		if err != nil {
			log.Fatalf("error reading config: %s", err.Error())
		}

		log.Println("Initializing timetta client...")
		if err := client.InitTimettaClient(cfg.Timetta.Credentials.Email, cfg.Timetta.Credentials.Password); err != nil {
			log.Fatalf("error initializing timetta client: %s", err.Error())
		}
		cl := client.GistTimettaClient()

		domain.PrintMembers(cl, cfg)
	},
}

func init() {
	rootCmd.AddCommand(resourcesCmd)
}
