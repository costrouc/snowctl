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

	err = cm.SetDefault()
	if err != nil {
		return fmt.Errorf("create client from connection manager %w", err)
	}

	applicationState := components.NewApplication(cm)

	ctx := context.Background()
	applicationState.Push(ctx, components.NewRolesView(cm, &components.RolesOptions{}))

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
