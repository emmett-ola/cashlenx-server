package migration_cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/macar-x/cashlenx-server/migrations"
	"github.com/macar-x/cashlenx-server/util"
	"github.com/macar-x/cashlenx-server/util/database"
	"github.com/spf13/cobra"
)

var MigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Inspect database migration state",
}

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify that MongoDB migration history is complete and immutable",
	RunE: func(cmd *cobra.Command, args []string) error {
		if util.GetConfigByKey("db.type") != "mongodb" {
			return fmt.Errorf("migration verify currently supports MongoDB; selected database is %q", util.GetConfigByKey("db.type"))
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		if err := migrations.VerifyMongo(ctx, database.GetMongoDatabase()); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "MongoDB migration history verified.")
		return nil
	},
}

func init() {
	MigrationCmd.AddCommand(verifyCmd)
}
