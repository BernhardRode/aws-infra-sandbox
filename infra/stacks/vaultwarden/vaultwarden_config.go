package vaultwarden

// VaultwardenConfig contains all configurable parameters for the Vaultwarden stack
type VaultwardenConfig struct {
	// Base configuration
	BaseImageName string
	BaseVersion   string
	DomainName    string
	
	// VPC configuration
	VpcCidr  string
	MaxAzs   int
	
	// ECS configuration
	ClusterName  string
	DesiredCount int
	Cpu          int
	MemoryMiB    int
	
	// EFS configuration
	FileSystemName           string
	EnableAutomaticBackups   bool
	LifecyclePolicyDays      int
	OutOfInfrequentAccessHits int
}

// DefaultVaultwardenConfig returns a configuration with sensible defaults
func DefaultVaultwardenConfig() *VaultwardenConfig {
	return &VaultwardenConfig{
		// Base configuration
		BaseImageName: "vaultwarden/server",
		BaseVersion:   "latest",
		DomainName:    "",
		
		// VPC configuration
		VpcCidr:  "20.0.0.0/24",
		MaxAzs:   2,
		
		// ECS configuration
		ClusterName:  "vaultwarden-cluster",
		DesiredCount: 1,
		Cpu:          256, // 0.25 vCPU
		MemoryMiB:    512, // 512 MB RAM
		
		// EFS configuration
		FileSystemName:           "vaultwarden-fs",
		EnableAutomaticBackups:   true,
		LifecyclePolicyDays:      14,
		OutOfInfrequentAccessHits: 1,
	}
}
