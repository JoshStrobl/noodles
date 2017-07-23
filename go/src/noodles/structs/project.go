//// Project plugin configuration

// SharedProjectPluginProperties
// Shared properties across all project plugins
type SharedProjectPluginProperties struct {
	*SharedPluginProperties
	// Required steps prior to execution of this specific project plugin
	requires string[]
	// Source directory of code we're using with this specific project plugin
	sourceDir string
}

// ProjectHtmlPluginProperties
// Properties for the HTML plugin in a project
type ProjectHtmlPluginProperties struct {
	*SharedGlobalHtmlPluginProperties
	*SharedProjectPluginProperties
}

// ProjectGolangPluginProperties
// Properties for the Golang plugin in a project
type ProjectGolangPluginProperties struct {
	*SharedGlobalGolangPluginProperties
	*SharedProjectPluginProperties
}

// ProjectLessPluginProperties
// Properties for the LESS plugin in a project
type ProjectLessPluginProperties struct {
	*SharedGlobalLessPluginProperties
	*SharedProjectPluginProperties
}

// ProjectTypeScriptPluginProperties
type ProjectTypeScriptPluginProperties struct {
	*SharedGlobalTypeScriptPluginProperties
	*SharedProjectPluginProperties
}