package components

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/costrouc/snowctl/internal/snowflake"
	"github.com/rivo/tview"
)

type SnowflakeContext struct {
	table             *tview.Table
	connectionManager *snowflake.ConnectionManager
	Schema            string
	Database          string
	Warehouse         string
	Role              string

	// Account Locator https://docs.snowflake.com/en/user-guide/admin-account-identifier#format-2-account-locator-in-a-region
	Account string
	// Account Name within Organization https://docs.snowflake.com/en/user-guide/admin-account-identifier#format-1-preferred-account-name-in-your-organization
	AccountName      string
	OrganizationName string
	Region           string
}

func NewSnowflakeContext(cm *snowflake.ConnectionManager) *SnowflakeContext {
	snowflakeContext := &SnowflakeContext{
		table:             tview.NewTable(),
		connectionManager: cm,
	}

	return snowflakeContext
}

func (s *SnowflakeContext) updateTable() {
	s.table.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf("[orange]Database: [gray]%s", s.Database)).SetAlign(tview.AlignLeft))
	s.table.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("[orange]Schema: [gray]%s", s.Schema)).SetAlign(tview.AlignLeft))
	s.table.SetCell(2, 0, tview.NewTableCell(fmt.Sprintf("[orange]Warehouse: [gray]%s", s.Warehouse)).SetAlign(tview.AlignLeft))
	s.table.SetCell(3, 0, tview.NewTableCell(fmt.Sprintf("[orange]Role: [gray]%s", s.Role)).SetAlign(tview.AlignLeft))

	s.table.SetCell(0, 1, tview.NewTableCell(fmt.Sprintf("[orange]Account: [gray]%s", s.Account)).SetAlign(tview.AlignLeft))
	s.table.SetCell(1, 1, tview.NewTableCell(fmt.Sprintf("[orange]Account Name: [gray]%s", s.AccountName)).SetAlign(tview.AlignLeft))
	s.table.SetCell(2, 1, tview.NewTableCell(fmt.Sprintf("[orange]Organization Name: [gray]%s", s.OrganizationName)).SetAlign(tview.AlignLeft))
	s.table.SetCell(3, 1, tview.NewTableCell(fmt.Sprintf("[orange]Region: [gray]%s", s.Region)).SetAlign(tview.AlignLeft))
}

func (s *SnowflakeContext) Update(ctx context.Context) error {
	type contextQuery struct {
		database         sql.NullString
		schema           sql.NullString
		warehouse        sql.NullString
		role             sql.NullString
		account          sql.NullString
		accountName      sql.NullString
		organizationName sql.NullString
		region           sql.NullString
	}

	var context contextQuery
	err := s.connectionManager.GetClient().SDKClient.GetConn().QueryRowContext(ctx, "SELECT CURRENT_DATABASE(), CURRENT_SCHEMA(), CURRENT_WAREHOUSE(), CURRENT_ROLE(), CURRENT_ACCOUNT(), CURRENT_ACCOUNT_NAME(), CURRENT_ORGANIZATION_NAME(), CURRENT_REGION();").Scan(&context.database, &context.schema, &context.warehouse, &context.role, &context.account, &context.accountName, &context.organizationName, &context.region)
	if err != nil {
		return fmt.Errorf("error fetching snowflake context %w", err)
	}

	s.Database = context.database.String
	s.Schema = context.schema.String
	s.Warehouse = context.warehouse.String
	s.Role = context.role.String
	s.Account = context.account.String
	s.AccountName = context.accountName.String
	s.OrganizationName = context.organizationName.String
	s.Region = context.region.String

	s.updateTable()

	return nil
}

func (s *SnowflakeContext) GetRender() *tview.Table {
	return s.table
}
