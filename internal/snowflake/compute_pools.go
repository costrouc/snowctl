package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type ComputePools interface {
	Create(ctx context.Context, id sdk.AccountObjectIdentifier, opts *CreateComputePoolOptions) error
	Show(ctx context.Context) ([]ComputePool, error)
	Drop(ctx context.Context, id sdk.AccountObjectIdentifier, opts *DropComputePoolOptions) error
	Describe(ctx context.Context, id sdk.AccountObjectIdentifier) (*ComputePoolDetails, error)
	Alter(ctx context.Context, id sdk.AccountObjectIdentifier, opts *AlterComputePoolOptions) error
	AlterState(ctx context.Context, id sdk.AccountObjectIdentifier, opts *AlterComputePoolStateOptions) error
}

type computepools struct {
	client *Client
}

type ComputePoolInstance string

const (
	CPU_X64_XS    ComputePoolInstance = "CPU_X64_XS"
	CPU_X64_S     ComputePoolInstance = "CPU_X64_S"
	CPU_X64_M     ComputePoolInstance = "CPU_X64_M"
	CPU_X64_L     ComputePoolInstance = "CPU_X64_L"
	HIGHMEM_X64_S ComputePoolInstance = "HIGHMEM_X64_S"
	HIGHMEM_X64_M ComputePoolInstance = "HIGHMEM_X64_M"
	HIGHMEM_X64_L ComputePoolInstance = "HIGHMEM_X64_L"
)

type CreateComputePoolOptions struct {
	IfNotExists        bool
	Application        *sdk.AccountObjectIdentifier
	MinNodes           int
	MaxNodes           int
	InstanceFamily     ComputePoolInstance
	AutoResume         *bool
	InitiallySuspended *bool
	AutoSuspendSecs    *int
	Comment            *string
}

func (c *computepools) Create(ctx context.Context, id sdk.AccountObjectIdentifier, opts *CreateComputePoolOptions) error {
	createTemplate := fmt.Sprintf(`
	CREATE COMPUTE POOL {{ if .IfNotExists }}IF NOT EXISTS{{ end }} %s
	    {{ if .Application }}FOR APPLICATION {{ .Application.FullQualifiedName }}{{ end }}
	    MIN_NODES = {{ .MinNodes }}
		MAX_NODES = {{ .MaxNodes }}
		INSTANCE_FAMILY = {{ .InstanceFamily }}
		{{ if .AutoResume }}AUTO_RESUME = {{ if .AutoResume }}TRUE{{ else }}FALSE{{ end }}{{ end }}
		{{ if .InitiallySuspended }}INITIALLY_SUSPENDED = {{ if .InitiallySuspended }}TRUE{{ else }}FALSE{{ end }}{{ end }}
		{{ if .AutoSuspendSecs }}AUTO_SUSPEND_SECS = {{ .AutoSuspendSecs }}{{ end }}
		{{ if .Comment }}COMMENT = '{{ .Comment }}'{{ end }};
	`, id.FullyQualifiedName())
	stmt := templateToQuery(createTemplate, opts)

	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}

type ComputePool struct {
	Name            string         `db:"name"`
	State           string         `db:"state"`
	MinNodes        int            `db:"min_nodes"`
	MaxNodes        int            `db:"max_nodes"`
	InstanceFamily  string         `db:"instance_family"`
	NumServices     int            `db:"num_services"`
	NumJobs         int            `db:"num_jobs"`
	AutoSuspendSecs int            `db:"auto_suspend_secs"`
	AutoResume      bool           `db:"auto_resume"`
	ActiveNodes     int            `db:"active_nodes"`
	IdleNodes       int            `db:"idle_nodes"`
	CreatedOn       sql.NullTime   `db:"created_on"`
	ResumedOn       sql.NullTime   `db:"resumed_on"`
	UpdatedOn       sql.NullTime   `db:"updated_on"`
	Owner           string         `db:"owner"`
	Comment         sql.NullString `db:"comment"`
	IsExclusive     bool           `db:"is_exclusive"`
	Application     sql.NullString `db:"application"`
}

func (c *computepools) Show(ctx context.Context) ([]ComputePool, error) {
	rows, err := c.client.SDKClient.GetConn().Query("SHOW COMPUTE POOLS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	computepools := make([]ComputePool, 0)
	for rows.Next() {
		var computepool ComputePool

		if err := rows.Scan(&computepool.Name, &computepool.State, &computepool.MinNodes, &computepool.MaxNodes, &computepool.InstanceFamily, &computepool.NumServices, &computepool.NumJobs, &computepool.AutoSuspendSecs, &computepool.AutoResume, &computepool.ActiveNodes, &computepool.IdleNodes, &computepool.CreatedOn, &computepool.ResumedOn, &computepool.UpdatedOn, &computepool.Owner, &computepool.Comment, &computepool.IsExclusive, &computepool.Application); err != nil {
			return nil, err
		}
		computepools = append(computepools, computepool)
	}

	return computepools, nil
}

type DropComputePoolOptions struct {
	IfExists bool
}

func (c *computepools) Drop(ctx context.Context, id sdk.AccountObjectIdentifier, opts *DropComputePoolOptions) error {
	dropTemplate := fmt.Sprintf("DROP COMPUTE POOL {{ if .IfExists }}IF EXISTS{{ end }} %s", id.FullyQualifiedName())
	stmt := templateToQuery(dropTemplate, opts)

	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}

type ComputePoolDetails struct {
	Name            string         `db:"name"`
	State           string         `db:"state"`
	MinNodes        int            `db:"min_nodes"`
	MaxNodes        int            `db:"max_nodes"`
	InstanceFamily  string         `db:"instance_family"`
	NumServices     int            `db:"num_services"`
	NumJobs         int            `db:"num_jobs"`
	AutoSuspendSecs int            `db:"auto_suspend_secs"`
	AutoResume      bool           `db:"auto_resume"`
	ActiveNodes     int            `db:"active_nodes"`
	IdleNodes       int            `db:"idle_nodes"`
	CreatedOn       sql.NullTime   `db:"created_on"`
	ResumedOn       sql.NullTime   `db:"resumed_on"`
	UpdatedOn       sql.NullTime   `db:"updated_on"`
	Owner           string         `db:"owner"`
	Comment         string         `db:"comment"`
	IsExclusive     bool           `db:"is_exclusive"`
	Application     sql.NullString `db:"application"`
}

func (c *computepools) Describe(ctx context.Context, id sdk.AccountObjectIdentifier) (*ComputePoolDetails, error) {
	stmt := fmt.Sprintf("DESCRIBE COMPUTE POOL %s", id.FullyQualifiedName())
	var describeComputePoolResult ComputePoolDetails

	err := c.client.SDKClient.GetConn().QueryRow(stmt).Scan(&describeComputePoolResult.Name, &describeComputePoolResult.State, &describeComputePoolResult.MinNodes, &describeComputePoolResult.MaxNodes, &describeComputePoolResult.InstanceFamily, &describeComputePoolResult.NumServices, &describeComputePoolResult.NumJobs, &describeComputePoolResult.AutoSuspendSecs, &describeComputePoolResult.AutoResume, &describeComputePoolResult.ActiveNodes, &describeComputePoolResult.IdleNodes, &describeComputePoolResult.CreatedOn, &describeComputePoolResult.ResumedOn, &describeComputePoolResult.UpdatedOn, &describeComputePoolResult.Owner, &describeComputePoolResult.Comment, &describeComputePoolResult.IsExclusive, &describeComputePoolResult.Application)
	if err != nil {
		return nil, err
	}

	return &describeComputePoolResult, nil
}

type AlterComputePoolOptions struct {
	IfExists        bool
	MinNodes        int
	MaxNodes        int
	AutoResume      bool
	AutoSuspendSecs int
	Comment         string
}

// https://docs.snowflake.com/en/sql-reference/sql/alter-compute-pool
func (c *computepools) Alter(ctx context.Context, id sdk.AccountObjectIdentifier, opts *AlterComputePoolOptions) error {
	alterTemplate := fmt.Sprintf(`
	ALTER COMPUTE POOL {{ if .IfExists }}IF EXISTS{{ end }} %s
		SET
		{{ if .MinNodes }}MIN NODES = {{ .MinNodes }}{{ end }}
		{{ if .MaxNodes }}MAX NODES = {{ .MaxNodes }}{{ end }}
		{{ if .AutoResume }}AUTO RESUME = {{ .AutoResume }}{{ end }}
		{{ if .AutoSuspendSecs }}AUTO SUSPEND = {{ .AutoSuspendSecs }}{{ end }}
		{{ if .Comment }}COMMENT = '{{ .Comment }}'{{ end }};
	`, id.FullyQualifiedName())
	stmt := templateToQuery(alterTemplate, opts)

	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}

type ComputePoolStateAction string

const (
	ComputePoolStateActionSuspend ComputePoolStateAction = "SUSPEND"
	ComputePoolStateActionResume  ComputePoolStateAction = "RESUME"
	ComputePoolStateActionStopAll ComputePoolStateAction = "STOP ALL"
)

type AlterComputePoolStateOptions struct {
	IfExists    bool
	StateAction ComputePoolStateAction
}

// https://docs.snowflake.com/en/sql-reference/sql/alter-compute-pool
func (c *computepools) AlterState(ctx context.Context, id sdk.AccountObjectIdentifier, opts *AlterComputePoolStateOptions) error {
	alterTemplate := fmt.Sprintf("ALTER COMPUTE POOL {{ if .IfExists }}IF EXISTS{{ end }} %s {{ .StateAction }};", id.FullyQualifiedName())
	stmt := templateToQuery(alterTemplate, opts)

	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}
