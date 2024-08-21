package snowflake

import (
	"context"
	"database/sql"
	"fmt"
)

type Listings interface {
	Create(ctx context.Context, id string, opts *CreateListingOptions) error
	Alter(ctx context.Context, id string, opts *AlterListingOptions) error
	Show(ctx context.Context) ([]Listing, error)
	Describe(ctx context.Context, id string) (*ListingDetails, error)
	Drop(ctx context.Context, id string, opts *DropListingOptions) error
}

type listings struct {
	client *Client
}

// https://other-docs.snowflake.com/en/progaccess/listing-manifest-reference
type ListingManifest struct {
	Title        string `json:"title" yaml:"title"`
	Subtitle     string `json:"subtitle" yaml:"subtitle"`
	Description  string `json:"description" yaml:"description"`
	Profile      string `json:"profile" yaml:"profile"`
	ListingTerms *struct {
		Type string `json:"type" yaml:"type"`
		Link string `json:"link,omitempty" yaml:"link,omitempty"`
	} `json:"listing_terms,omitempty" yaml:"listing_terms,omitempty"`
	Targets *struct {
		Accounts []string `json:"accounts,omitempty" yaml:"accounts,omitempty"`
		Regions  []string `json:"regions,omitempty" yaml:"regions,omitempty"`
	} `json:"targets,omitempty" yaml:"targets,omitempty"`
	AutoFulfillment *struct {
		RefreshSchedule string `json:"refresh_schedule,omitempty" yaml:"refresh_schedule,omitempty"`
		RefreshType     string `json:"refresh_type,omitempty" yaml:"refresh_type,omitempty"`
	} `json:"auto_fulfillment,omitempty" yaml:"auto_fulfillment,omitempty"`
	BusinessNeeds []struct {
		Name        string `json:"name" yaml:"name"`
		Description string `json:"description" yaml:"description"`
		Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	} `json:"business_needs,omitempty" yaml:"business_needs,omitempty"`
	Categories []string `json:"categories,omitempty" yaml:"categories,omitempty"`
	// DataAttributes
	// DataDictionary
	// UsageExamples
	Resources *struct {
		Documentation string `json:"documentation,omitempty" yaml:"documentation,omitempty"`
		Media         string `json:"media,omitempty" yaml:"media,omitempty"`
	} `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type CreateListingOptions struct {
	ListingManifest    *ListingManifest
	IfNotExists        bool
	ApplicationPackage string
	Publish            bool
	Review             bool
}

// https://other-docs.snowflake.com/en/sql-reference/sql/create-listing
func (c *listings) Create(ctx context.Context, id string, opts *CreateListingOptions) error {
	if opts == nil {
		opts = &CreateListingOptions{}
	}

	createListingTemplate := fmt.Sprintf(`
	CREATE EXTERNAL LISTING {{if .IfNotExists}}IF NOT EXISTS{{end}} %s
	    APPLICATION PACKAGE {{.ApplicationPackage}}
		AS $$ {{json .ListingManifest}} $$
		PUBLISH = {{if .Publish}}TRUE{{ else }}FALSE{{end}}
		REVIEW = {{if .Review}}TRUE{{else}}FALSE{{end}};
	`, id)
	stmt := templateToQuery(createListingTemplate, opts)
	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}

type AlterListingOptions struct {
	ListingManifest *ListingManifest
	IfExists        bool
	Publish         bool
	Review          bool
}

// https://other-docs.snowflake.com/en/sql-reference/sql/alter-listing
func (c *listings) Alter(ctx context.Context, id string, opts *AlterListingOptions) error {
	alterListingTemplate := fmt.Sprintf(`
	ALTER LISTING {{if .IfExists}}IF EXISTS{{end}} %s
		AS $$ {{json .ListingManifest}} $$
		PUBLISH = {{if .Publish}}TRUE{{ else }}FALSE{{end}}
		REVIEW = {{if .Review}}TRUE{{else}}FALSE{{end}};
	`, id)
	stmt := templateToQuery(alterListingTemplate, opts)
	_, err := c.client.SDKClient.GetConn().ExecContext(ctx, stmt)
	if err != nil {
		return err
	}

	return nil
}

type Listing struct {
	GlobalName             string         `db:"global_name"`
	Name                   string         `db:"name"`
	Title                  string         `db:"title"`
	Subtitle               sql.NullString `db:"subtitle"`
	Profile                string         `db:"profile"`
	CreatedOn              sql.NullTime   `db:"created_on"`
	UpdatedOn              sql.NullTime   `db:"updated_on"`
	PublishedOn            sql.NullTime   `db:"published_on"`
	State                  string         `db:"state"`
	ReviewState            sql.NullString `db:"review_state"`
	Comment                sql.NullString `db:"comment"`
	Owner                  string         `db:"owner"`
	OwnerRoleType          string         `db:"owner_role_type"`
	Regions                sql.NullString `db:"regions"`
	TargetAccounts         string         `db:"target_accounts"`
	IsMonetized            bool           `db:"is_monetized"`
	IsApplication          bool           `db:"is_application"`
	IsTargeted             bool           `db:"is_targeted"`
	IsLimitedTrial         bool           `db:"is_limited_trial"`
	IsByRequest            bool           `db:"is_by_request"`
	RejectedOn             sql.NullTime   `db:"rejected_on"`
	DetailedTargetAccounts sql.NullString `db:"detailed_target_accounts"`
}

// https://other-docs.snowflake.com/en/sql-reference/sql/show-listings
func (c *listings) Show(ctx context.Context) ([]Listing, error) {
	rows, err := c.client.SDKClient.GetConn().Query("SHOW LISTINGS")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	listings := make([]Listing, 0)
	for rows.Next() {
		var listing Listing

		if err := rows.Scan(&listing.GlobalName, &listing.Name, &listing.Title, &listing.Subtitle, &listing.Profile, &listing.CreatedOn, &listing.UpdatedOn, &listing.PublishedOn, &listing.State, &listing.ReviewState, &listing.Comment, &listing.Owner, &listing.OwnerRoleType, &listing.Regions, &listing.TargetAccounts, &listing.IsMonetized, &listing.IsApplication, &listing.IsTargeted, &listing.IsLimitedTrial, &listing.IsByRequest, &listing.RejectedOn, &listing.DetailedTargetAccounts); err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}

	return listings, nil
}

type ListingDetails struct {
	GlobalName               string         `db:"global_name"`
	Name                     string         `db:"name"`
	Owner                    string         `db:"owner"`
	OwnerRoleType            string         `db:"owner_role_type"`
	CreatedOn                sql.NullTime   `db:"created_on"`
	UpdatedOn                sql.NullTime   `db:"updated_on"`
	PublishedOn              sql.NullTime   `db:"published_on"`
	Title                    string         `db:"title"`
	Subtitle                 sql.NullString `db:"subtitle"`
	Description              sql.NullString `db:"description"`
	ListingTerms             sql.NullString `db:"listing_terms"`
	State                    string         `db:"state"`
	Share                    sql.NullString `db:"share"`
	ApplicationPackage       string         `db:"application_package"`
	BusinessNeeds            sql.NullString `db:"business_needs"`
	UsageExamples            sql.NullString `db:"usage_examples"`
	DataAttributes           sql.NullString `db:"data_attributes"`
	Categories               sql.NullString `db:"categories"`
	Resources                sql.NullString `db:"resources"`
	Profile                  sql.NullString `db:"profile"`
	CustomizedContactInfo    sql.NullString `db:"customized_contact_info"`
	DataDictionary           sql.NullString `db:"data_dictionary"`
	DataPreview              sql.NullString `db:"data_preview"`
	Comment                  sql.NullString `db:"comment"`
	Revisions                string         `db:"revisions"`
	TargetAccounts           sql.NullString `db:"target_accounts"`
	Regions                  sql.NullString `db:"regions"`
	RefreshSchedule          sql.NullString `db:"refresh_schedule"`
	RefreshType              sql.NullString `db:"refresh_type"`
	ReviewState              sql.NullString `db:"review_state"`
	RejectionReason          sql.NullString `db:"rejection_reason"`
	UnpublishedByAdminReason sql.NullString `db:"unpublished_by_admin_reason"`
	IsMonetized              bool           `db:"is_monetized"`
	IsApplication            bool           `db:"is_application"`
	IsTargeted               bool           `db:"is_targeted"`
	IsLimitedTrial           bool           `db:"is_limited_trial"`
	IsByRequest              bool           `db:"is_by_request"`
	LimitedTrialPlan         sql.NullString `db:"limited_trial_plan"`
	RetiredOn                sql.NullTime   `db:"retired_on"`
	ScheduledDropTime        sql.NullTime   `db:"scheduled_drop_time"`
	ManifestYAML             string         `db:"manifest_yaml"`
}

// https://other-docs.snowflake.com/en/sql-reference/sql/desc-listing
func (c *listings) Describe(ctx context.Context, id string) (*ListingDetails, error) {
	stmt := fmt.Sprintf("DESCRIBE LISTING %s", id)
	var describeListingResult ListingDetails

	err := c.client.SDKClient.GetConn().QueryRow(stmt).Scan(&describeListingResult.GlobalName, &describeListingResult.Name, &describeListingResult.Owner, &describeListingResult.OwnerRoleType, &describeListingResult.CreatedOn, &describeListingResult.UpdatedOn, &describeListingResult.PublishedOn, &describeListingResult.Title, &describeListingResult.Subtitle, &describeListingResult.Description, &describeListingResult.ListingTerms, &describeListingResult.State, &describeListingResult.Share, &describeListingResult.ApplicationPackage, &describeListingResult.BusinessNeeds, &describeListingResult.UsageExamples, &describeListingResult.DataAttributes, &describeListingResult.Categories, &describeListingResult.Resources, &describeListingResult.Profile, &describeListingResult.CustomizedContactInfo, &describeListingResult.DataDictionary, &describeListingResult.DataPreview, &describeListingResult.Comment, &describeListingResult.Revisions, &describeListingResult.TargetAccounts, &describeListingResult.Regions, &describeListingResult.RefreshSchedule, &describeListingResult.RefreshType, &describeListingResult.ReviewState, &describeListingResult.RejectionReason, &describeListingResult.UnpublishedByAdminReason, &describeListingResult.IsMonetized, &describeListingResult.IsApplication, &describeListingResult.IsTargeted, &describeListingResult.IsLimitedTrial, &describeListingResult.IsByRequest, &describeListingResult.LimitedTrialPlan, &describeListingResult.RetiredOn, &describeListingResult.ScheduledDropTime, &describeListingResult.ManifestYAML)
	if err != nil {
		return nil, err
	}

	return &describeListingResult, nil
}

type DropListingOptions struct {
	IfExists bool
}

// https://other-docs.snowflake.com/en/sql-reference/sql/drop-listing
func (c *listings) Drop(ctx context.Context, id string, opts *DropListingOptions) error {
	if opts == nil {
		opts = &DropListingOptions{}
	}

	dropListingTemplate := fmt.Sprintf("DROP LISTING %s {{if .IfExists}}IF EXISTS{{end}};", id)
	stmt := templateToQuery(dropListingTemplate, opts)
	_, err := c.client.SDKClient.GetConn().Exec(stmt)
	if err != nil {
		return err
	}

	return nil
}
