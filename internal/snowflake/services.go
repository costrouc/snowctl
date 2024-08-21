package snowflake

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
)

type Services interface {
	Create(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *CreateServiceOptions) error
	Show(ctx context.Context, opts *ShowServiceOptions) ([]Service, error)
	Drop(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *DropServiceOptions) error
	Describe(ctx context.Context, id sdk.SchemaObjectIdentifier) (*ServiceDetails, error)
	Alter(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *AlterServiceOptions) error
}

type services struct {
	client *Client
}

type ManifestContainer struct {
	Name    string   `json:"name"`
	Image   string   `json:"image"`
	Command []string `json:"command"`
	Args    []string `json:"args"`
	Env     []string `json:"env"`

	ReadinessProbe struct {
		Port int    `json:"port"`
		Path string `json:"path"`
	} `json:"readinessProbe"`

	VolumeMounts []struct {
		Name      string `json:"name"`
		MountPath string `json:"mountPath"`
	}

	Resources struct {
		Requests struct {
			Memory       string `json:"memory"`
			CPU          string `json:"cpu"`
			NvidiaComGpu int    `json:"nvidia.com/gpu"`
		} `json:"requests"`

		Limits struct {
			Memory       string `json:"memory"`
			CPU          string `json:"cpu"`
			NvidiaComGpu int    `json:"nvidia.com/gpu"`
		} `json:"limits"`
	} `json:"resources"`

	Secrets []struct {
		SnowflakeSecret string `json:"snowflake_secret"`
		SecretKeyRef    string `json:"secretKeyRef"`
		EnvVarName      string `json:"envVarName"`
		DirectoryPath   string `json:"directoryPath"`
	} `json:"secrets"`
}

type ManifestEndpoint struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Public   bool   `json:"public"`
	Protocol string `json:"protocol"`
}

type ManifestVolume struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Size        string `json:"size"`
	BlockConfig *struct {
		InitialContents struct {
			FromSnapshot string `json:"fromSnapshot"`
		} `json:"initialContents"`
	} `json:"blockConfig"`
	Uid int `json:"uid"`
	Gid int `json:"gid"`
}

type ManifestServiceRole struct {
	Name      string   `json:"name"`
	Endpoints []string `json:"endpoints"`
}

type ServiceManifest struct {
	Spec struct {
		Containers []*ManifestContainer `json:"containers"`
		Endpoints  []*ManifestEndpoint  `json:"endpoints"`
		Volumes    []*ManifestVolume    `json:"volumes"`

		LogExporters struct {
			EventTableConfig struct {
				LogLevel string `json:"logLevel"`
			} `json:"eventTableConfig"`
		}

		ServiceRoles []*ManifestServiceRole `json:"serviceRoles"`
	} `json:"spec"`
}

type CreateServiceOptions struct {
	IfNotExists                bool
	ComputePool                sdk.AccountObjectIdentifier
	ServiceManifest            *ServiceManifest
	ExternalAccessIntegrations []sdk.AccountObjectIdentifier
	AutoResume                 bool
	MinInstances               int
	MaxInstances               int
	QueryWarehouse             sdk.AccountObjectIdentifier
	Comment                    string
}

func (s *services) Create(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *CreateServiceOptions) error {
	createComputePoolTemplate := fmt.Sprintf(`
	CREATE SERVICE {{if .IfNotExists}}IF NOT EXISTS{{end}} %s
	    IF COMPUTE POOL {{ .ComputePool }}
	    {{ if .ExternalAccessIntegrations }}EXTERNAL_ACCESS_INTEGRATIONS = ({{ range .ExternalAccessIntegrations }}{{ .Name }},{{ end }}){{ end }}
		AUTO_RESUME = {{ .AutoResume }}
		MIN_INSTANCES = {{ .MinInstances }}
		MAX_INSTANCES = {{ .MaxInstances }}
		{{ if .QueryWarehouse }}QUERY_WAREHOUSE = {{ .QueryWarehouse }}{{ end }}
		{{ if .Comment }}COMMENT = '{{ .Comment }}'{{ end }};
	`, id.FullyQualifiedName())
	stmt := templateToQuery(createComputePoolTemplate, opts)
	_, err := s.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}

type Service struct {
	Name                      string         `db:"name"`
	Status                    string         `db:"status"`
	DatabaseName              string         `db:"database_name"`
	SchemaName                string         `db:"schema_name"`
	Owner                     string         `db:"owner"`
	ComputePool               string         `db:"compute_pool"`
	DNSName                   string         `db:"dns_name"`
	MinInstances              int            `db:"min_instances"`
	MaxInstances              int            `db:"max_instances"`
	AutoResume                bool           `db:"auto_resume"`
	ExternalAccessIntegration string         `db:"external_access_integration"`
	CreatedOn                 sql.NullTime   `db:"created_on"`
	UpdatedOn                 sql.NullTime   `db:"updated_on"`
	ResumedOn                 sql.NullTime   `db:"resumed_on"`
	Comment                   sql.NullString `db:"comment"`
	OwnerRoleType             string         `db:"owner_role_type"`
	QueryWarehouse            sql.NullString `db:"query_warehouse"`
	IsJob                     bool           `db:"is_job"`
}

type ShowServiceOptions struct {
	ComputePool *sdk.AccountObjectIdentifier
	Database    *sdk.AccountObjectIdentifier
	Schema      *sdk.DatabaseObjectIdentifier
}

func (s *services) Show(ctx context.Context, opts *ShowServiceOptions) ([]Service, error) {
	query := "SHOW SERVICES"
	if opts.ComputePool != nil {
		query = fmt.Sprintf("SHOW SERVICES IN COMPUTE POOL %s", opts.ComputePool.FullyQualifiedName())
	}
	if opts.Database != nil {
		query = fmt.Sprintf("SHOW SERVICES IN DATABASE %s", opts.Database.FullyQualifiedName())
	}
	if opts.Schema != nil {
		query = fmt.Sprintf("SHOW SERVICES IN SCHEMA %s", opts.Schema.FullyQualifiedName())
	}

	if query == "" {
		return nil, fmt.Errorf("one of the show users options must be not nil")
	}

	rows, err := s.client.SDKClient.GetConn().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]Service, 0)
	for rows.Next() {
		var service Service

		if err := rows.Scan(&service.Name, &service.Status, &service.DatabaseName, &service.SchemaName, &service.Owner, &service.ComputePool, &service.DNSName, &service.MinInstances, &service.MaxInstances, &service.AutoResume, &service.ExternalAccessIntegration, &service.CreatedOn, &service.UpdatedOn, &service.ResumedOn, &service.Comment, &service.OwnerRoleType, &service.QueryWarehouse, &service.IsJob); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

type DropServiceOptions struct {
	IfExists bool
}

func (s *services) Drop(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *DropServiceOptions) error {
	template := fmt.Sprintf("DROP SERVICE {{ if .IfExists }}IF EXISTS{{ end}} %s", id.FullyQualifiedName())
	stmt := templateToQuery(template, opts)
	_, err := s.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}

type ServiceDetails struct {
	Name                      string       `db:"name"`
	DatabaseName              string       `db:"database_name"`
	SchemaName                string       `db:"schema_name"`
	Owner                     string       `db:"owner"`
	ComputePool               string       `db:"compute_pool"`
	DNSName                   string       `db:"dns_name"`
	MinInstances              int          `db:"min_instances"`
	MaxInstances              int          `db:"max_instances"`
	AutoResume                bool         `db:"auto_resume"`
	ExternalAccessIntegration string       `db:"external_access_integration"`
	CreatedOn                 sql.NullTime `db:"created_on"`
	UpdatedOn                 sql.NullTime `db:"updated_on"`
	ResumedOn                 sql.NullTime `db:"resumed_on"`
	Comment                   string       `db:"comment"`
	OwnerRoleType             string       `db:"owner_role_type"`
	QueryWarehouse            string       `db:"query_warehouse"`
	IsJob                     bool         `db:"is_job"`
}

func (s *services) Describe(ctx context.Context, id sdk.SchemaObjectIdentifier) (*ServiceDetails, error) {
	stmt := fmt.Sprintf("DESCRIBE SERVICE %s", id.FullyQualifiedName())
	var describeServiceResult ServiceDetails

	err := s.client.SDKClient.GetConn().QueryRow(stmt).Scan(&describeServiceResult.Name, &describeServiceResult.DatabaseName, &describeServiceResult.SchemaName, &describeServiceResult.Owner, &describeServiceResult.ComputePool, &describeServiceResult.DNSName, &describeServiceResult.MinInstances, &describeServiceResult.MaxInstances, &describeServiceResult.AutoResume, &describeServiceResult.ExternalAccessIntegration, &describeServiceResult.CreatedOn, &describeServiceResult.UpdatedOn, &describeServiceResult.ResumedOn, &describeServiceResult.Comment, &describeServiceResult.OwnerRoleType, &describeServiceResult.QueryWarehouse, &describeServiceResult.IsJob)
	if err != nil {
		return nil, err
	}

	return &describeServiceResult, nil
}

type AlterServiceOptions struct {
	IfExists                  bool
	MinInstances              *int
	MaxInstances              *int
	QueryWarehouse            *sdk.AccountObjectIdentifier
	AutoResume                *bool
	ExternalAccessIntegration []sdk.AccountObjectIdentifier
	Comment                   *string
}

func (c *services) Alter(ctx context.Context, id sdk.SchemaObjectIdentifier, opts *AlterServiceOptions) error {
	if opts == nil {
		opts = &AlterServiceOptions{}
	}

	alterServiceTemplate := fmt.Sprintf(`
	ALTER SERVICE {{ if .IfExists }}IF EXISTS{{ end }} %s
	     SET
		 {{ if .MinInstances }}MIN_INSTANCES = {{ .MinInstances }}{{ end }}
		 {{ if .MaxInstances }}MAX_INSTANCES = {{ .MaxInstances }}{{ end }}
		 {{ if .QueryWarehouse }}QUERY_WAREHOUSE = {{ .QueryWarehouse }}{{ end }}
		 {{ if .AutoResume }}AUTO_RESUME = {{ .AutoResume }}{{ end }}
		 {{ if .ExternalAccessIntegration }}EXTERNAL_ACCESS_INTEGRATIONS = ({{ range .ExternalAccessIntegration }}{{ .Name }},{{ end }}){{ end }}
		 {{ if .Comment }}COMMENT = '{{ .Comment }}'{{ end }};
	`, id.FullyQualifiedName())
	stmt := templateToQuery(alterServiceTemplate, opts)
	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	return err
}
