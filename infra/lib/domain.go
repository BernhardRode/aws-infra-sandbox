package lib

import (
	"fmt"
)

// DomainConfig contains domain configuration for the entire infrastructure
type DomainConfig struct {
	// The root domain name (e.g., "ebbo.dev")
	RootDomain string
	
	// The hosted zone ID for the root domain
	HostedZoneId string
}

// GetAppDomain returns the full domain for an application in the current environment
// Format: <app>.<environment>.rootDomain (e.g., vaultwarden.staging.ebbo.dev)
func (d *DomainConfig) GetAppDomain(appName string, env Environment) string {
	return fmt.Sprintf("%s.%s.%s", appName, env.GetEnvPrefix(), d.RootDomain)
}

// GetEnvironmentDomain returns the base domain for the current environment
// Format: <environment>.rootDomain (e.g., staging.ebbo.dev)
func (d *DomainConfig) GetEnvironmentDomain(env Environment) string {
	return fmt.Sprintf("%s.%s", env.GetEnvPrefix(), d.RootDomain)
}

// DefaultDomainConfig returns a configuration with default values
func DefaultDomainConfig() *DomainConfig {
	return &DomainConfig{
		RootDomain:   "ebbo.dev",
		HostedZoneId: "Z02287733RP9AY57D3IRQ", // The hosted zone ID from your existing code
	}
}
