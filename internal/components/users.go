package components

import (
	"context"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type UsersView struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	options           *UsersOptions
}

type UsersOptions struct{}

func NewUsersView(connectionManager *snowflake.ConnectionManager, opts *UsersOptions) *UsersView {
	users := &UsersView{
		table:             tview.NewTable(),
		connectionManager: connectionManager,
	}

	users.table.SetFixed(1, 0).SetSelectable(true, false).SetBorder(true)

	return users
}

func (v *UsersView) Update(ctx context.Context) error {
	table, err := v.getData(ctx)
	if err != nil {
		return fmt.Errorf("updating users data %w", err)
	}

	updateTable(v.table, table)

	return nil
}

func (v *UsersView) getData(ctx context.Context) (*Table, error) {
	users, err := v.connectionManager.GetClient().SDKClient.Users.Show(ctx, &sdk.ShowUserOptions{})
	if err != nil {
		return nil, fmt.Errorf("calling snowflake show users %w", err)
	}

	columns := []string{"Name", "Email", "First Name", "Last Name", "Default Role"}
	rows := make([][]string, 0)

	for _, user := range users {
		rows = append(rows, []string{
			user.Name,
			user.Email,
			user.FirstName,
			user.LastName,
			user.DefaultRole,
		})
	}

	return &Table{
		Title:   "users",
		Columns: columns,
		Rows:    rows,
	}, nil
}

func (v *UsersView) GetRender() tview.Primitive {
	return v.table
}

func (t *UsersView) GetBindings(ctx context.Context, applicationState *ApplicationState) []*KeyBinding {
	return []*KeyBinding{}
}
