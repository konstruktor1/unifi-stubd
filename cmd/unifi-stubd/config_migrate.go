package main

import (
	"fmt"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

func runConfigMigration(flags runtimeFlags) error {
	if flags.configMigrate && flags.configMigrateDryRun {
		return fmt.Errorf("-config-migrate and -config-migrate-dry-run cannot be used together")
	}
	path := strings.TrimSpace(flags.configPath)
	if path == "" {
		return fmt.Errorf("-config is required for config migration")
	}
	migration, err := appconfig.MigrateFile(path, flags.configMigrateDryRun)
	if err != nil {
		return fmt.Errorf("migrate config %s: %w", path, err)
	}
	for _, action := range migration.Result.Actions {
		fmt.Printf("config migration: %s\n", action)
	}
	for _, warning := range migration.Result.Warnings {
		fmt.Printf("config migration warning: %s\n", warning)
	}
	if !migration.Result.Changed {
		fmt.Printf("config migration: no changes needed: %s\n", path)
		return nil
	}
	if flags.configMigrateDryRun {
		fmt.Println("config migration: dry run, no changes written")
		fmt.Println("migrated_config_yaml:")
		fmt.Print(string(migration.Result.Data))
		return nil
	}
	fmt.Printf("config migration backup: %s\n", migration.BackupPath)
	fmt.Printf("config migration: wrote %s\n", path)
	return nil
}
