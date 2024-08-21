package main

import (
	"context"
	"fmt"
	"os"

	"github.com/costrouc/snowctl/internal/components"
	"github.com/costrouc/snowctl/internal/snowflake"
)

func run() error {
	cm, err := snowflake.NewConnectionManager()
	if err != nil {
		return fmt.Errorf("creating snowflake connection manager %w", err)
	}
	connectionName := "snowsql.default"
	err = cm.SetClient(connectionName)
	if err != nil {
		return fmt.Errorf("create client from connection manager %s %w", connectionName, err)
	}

	applicationState := components.NewApplication(cm)

	ctx := context.Background()
	applicationState.Push(ctx, components.NewUsersView(cm, &components.UsersOptions{}))

	if err := applicationState.Application.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running %s\n", err.Error())
		os.Exit(1)
	}
}
