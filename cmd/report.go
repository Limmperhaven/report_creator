package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/client"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/config"
	"gitlab.digital-spirit.ru/tn-projs/dev-tools/reports/internal/domain"
	"log"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Make report",
	Long:  `Using this command you can create reports and count cost of all work`,
	Run: func(cmd *cobra.Command, args []string) {
		outPath, err := cmd.Flags().GetString("path")
		cfgPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("error reading flags: %s", err.Error())
		}

		log.Println("Reading configs...")
		cfg, err := config.ReadConfig(cfgPath, outPath)
		if err != nil {
			log.Fatalf("error reading config: %s", err.Error())
		}

		log.Println("Initializing timetta client...")
		if err := client.InitTimettaClient(cfg.Timetta.Credentials.Email, cfg.Timetta.Credentials.Password); err != nil {
			log.Fatalf("error initializing timetta client: %s", err.Error())
		}
		cl := client.GistTimettaClient()

		domain.GenerateReport(cl, cfg)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	rootCmd.PersistentFlags().StringP("config", "c", "etc/config.yaml", "cfg file path")
	reportCmd.Flags().StringP("path", "p", "", "Path for result files")
}
