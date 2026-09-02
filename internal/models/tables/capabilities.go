package tables

// DeviceCapabilities matches device.Capabilities JSON tags (no omitempty).
type DeviceCapabilities struct {
	CloudProjects      bool `json:"cloud_projects"`
	OrganizationAssets bool `json:"organization_assets"`
	ProposalBranding   bool `json:"proposal_branding"`
	EngineeringHandoff bool `json:"engineering_handoff"`
	MaxSeats           int  `json:"max_seats"`
	MaxDevices         int  `json:"max_devices"`
	MaxCloudProjects   int  `json:"max_cloud_projects"`
}

// PlanCapabilities matches billing.Capabilities JSON tags (omitempty on limits).
type PlanCapabilities struct {
	CloudProjects      bool `json:"cloud_projects"`
	OrganizationAssets bool `json:"organization_assets"`
	ProposalBranding   bool `json:"proposal_branding"`
	EngineeringHandoff bool `json:"engineering_handoff"`
	MaxSeats           int  `json:"max_seats,omitempty"`
	MaxDevices         int  `json:"max_devices,omitempty"`
	MaxCloudProjects   int  `json:"max_cloud_projects,omitempty"`
}
