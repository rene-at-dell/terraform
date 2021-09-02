package plugins

import (
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform/internal/addrs"
	"github.com/hashicorp/terraform/internal/getproviders"
	"github.com/hashicorp/terraform/internal/providercache"
	"github.com/hashicorp/terraform/internal/providers"
)

// Launcher is a container type for various configuration settings and other
// metadata that we will use to locate, verify, and launch plugin components,
// which currently includes both providers and provisioners.
//
// The responsibility of this type includes both true plugins, where we start
// a child process and talk to it over gRPC, and also our various
// "pseudo-plugins" which really live directly inside the Terraform executable
// but which implement the same interfaces that Terraform Core expects for
// plugin components.
//
// Launcher is not responsible for credentials helper plugins. The "svchost"
// functionality owns the overall problem of authenticating to Terraform-native
// services, which includes the idea of credentials helpers.
//
// The responsibility for configuring a Launcher is unfortunately spread over
// multiple Terraform CLI subsystems, and so this type follows the Builder
// Pattern to allow gradually adding additional constraints when needed,
// such as layering on additional checksum verification when applying a saved
// plan in order to ensure that the plugins are identical to those which
// created the plan.
type Launcher struct {
	// Providers and provisioners currently have a pretty different model
	// because providers kept moving forward to support the public registry,
	// private registries, etc while for provisioners we're just maintaining
	// the functionality as-is because they are not a mechanism we recommend
	// for new configurations.
	//
	// Unfortunately that means that this type is a bit of a two-headed
	// monster trying to make both of these appear somewhat similar to
	// callers. The configuration portions are quite distinct for each, but
	// the result for both is roughly the same: a set of factory functions
	// each responsible for verifying and launching a particular plugin.

	////// Provider-related data
	// providerDir is the object representing the directory where we keep
	// the local copies of provider plugins. The provider installer's job
	// is to write plugin packages into this directory, whereas our job
	// is to verify that the directory contains the required set of plugins
	// and then, if so, to launch them as needed.
	providerDir *providercache.Dir
	// providerBuiltins represents all of the providers belonging to the
	// special "built-in" namespace, terraform.io/builtin/<name>. These
	// are compiled directly into the Terraform executable, but Terraform
	// Core isn't aware that they are special.
	providerBuiltins map[string]providers.Factory
	// providerDevOverrides represents a set of plugins that are configured
	// in a special "dev override" mode, which means that the CLI configuration
	// specifies a fixed local directory to load from and we skip all of our
	// usual version and checksum verification. This mechanism is for provider
	// developers who want to use Terraform with as-yet-unpublished builds which
	// therefore don't have version numbers or finalized checksums yet.
	providerDevOverrides map[addrs.Provider]getproviders.PackageLocalDir
	// providersUnmanaged represents a set of plugins that Terraform doesn't
	// directly manage at all, and instead just expects to somehow already
	// be running. This is in some sense a more extreme version of
	// providerDevOverrides, and exists only to allow the plugin integration
	// test harness to make the provider's test process itself behave as an
	// already-running plugin server, respecting any special behavior set up by
	// the test program.
	providersUnmanaged map[addrs.Provider]*plugin.ReattachConfig

	////// Provisioner-related data
	// provisionerSearchDirs is a set of directories that we'll search for
	// available third-party provisioner plugins. Unlike providers, there is
	// no automatic installation mechanism for these and so users must place
	// each plugin in one of these directories manually.
	provisionerSearchDirs []string
}

// LauncherBaseSettings acts as a starting point for a new Launcher, capturing
// the subset of the settings which we're able to determine up front in
// "package main", without reference to a particular configuration or state.
//
// The immediate result of NewLauncher with some base settings will typically
// not be sufficient for real use. Callers will need to at least create a new
// derived launcher which has a set of plugin requirements derived from a
// configuration and state, and possibly also lock in a fixed set of provider
// checksums if the intended action is to apply a saved plan file.
type LauncherBaseSettings struct {
	// ProviderDir is the path of the directory where the launcher can expect
	// to find the packages placed by the provider installer.
	//
	// This directory is the _local_ (working-directory-specific) provider
	// cache directory, and not the global provider cache directory which
	// might've been set in the CLI configuration. Only the provider installer
	// cares about the global provider cache directory, and it must ensure
	// that all providers are copied or linked into the local cache directory
	// as part of its work.
	ProviderDir string

	// BuiltinProviders is a map of factories for providers that are compiled
	// directly into the Terraform CLI program, rather than launched as child
	// processes using the plugin protocol.
	//
	// The keys of this map represent the provider type portion of the
	// resulting provider address. Built-in providers always belong to the
	// namespace terraform.io/builtin/, and the provider installer knows that
	// it doesn't need to actually install providers belonging to that
	// namespace because they should always be available to launch via this
	// map.
	BuiltinProviders map[string]providers.Factory

	// ProviderDevOverrides is a map of forced local package directories for
	// particular provider addresses.
	//
	// For any provider whose address is a key in this map, the launcher will
	// skip all of the usual version and checksum verification and will
	// instead attempt to launch the provider plugin directly from the given
	// local package directory. This is a mechanism by which provider
	// developers can test against unreleased builds of their providers.
	ProviderDevOverrides map[addrs.Provider]getproviders.PackageLocalDir

	// UnmanagedProviders is a map of providers that the launcher will assume
	// are already running as externally-managed plugin processes, and so
	// the launcher will just attempt to attach using the given settings
	// rather than to start a separate process.
	// This is a mechanism by which the provider integration test harness
	// can use the provider's own test program as a plugin server, and thus
	// have it honor any special settings or behaviors activated by the test
	// cases.
	UnmanagedProviders map[addrs.Provider]*plugin.ReattachConfig
}

// NewLauncher initializes a new launcher with the given base settings, which
// are all values we expect "package main" to decide unilaterally before we
// potentially derive new launchers with additional constraints in downstream
// components.
//
// The immediate result doesn't know which plugins are actually required,
// so it will not actually offer any plugins at all. A downstream component
// must at the very least add a set of provider requirements in order to
// allow for version verification.
func NewLauncher(baseSettings LauncherBaseSettings) Launcher {
	providerDir := providercache.NewDir(baseSettings.ProviderDir)

}
