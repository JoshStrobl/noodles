package main

import (
	"errors"
	"github.com/JoshStrobl/trunk"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var newCmd = &cobra.Command{
	Use:               "new",
	Short:             "Creates a Noodles workspace, projects, or scripts",
	Long:              "Creates a Noodles workspace, projects, or scripts",
	Run:               new,
	DisableAutoGenTag: true,
}

var newProjectName string
var newScriptName string

func init() {
	newCmd.Flags().StringVarP(&newProjectName, "project", "p", "", "Name of a new project you wish to bootstrap")
	newCmd.Flags().StringVarP(&newScriptName, "script", "s", "", "Name of a new script you wish to bootstrap")
}

func new(cmd *cobra.Command, args []string) {
	if (newProjectName == "") && (newScriptName == "") { // If we're creating a new Noodles workspace
		if configInfo, statErr := os.Stat("noodles.toml"); statErr == nil { // Check if noodles.toml already exists
			if configInfo.Size() != 0 { // If the size of the file is greater than 0, meaning there is potentially content
				trunk.LogInfo("noodles.toml already exists and appears to have content. Exiting.")
				return
			}
		}

		NewWorkspacePrompt() // Perform our workspace prompting
	} else {
		if noodles.Name == "" { // Noodles workspace doesn't seem to exist
			trunk.LogErr("No Noodles workspace appears to exist. Please create a workspace first.")
			os.Exit(1)
		}

		if newProjectName != "" { // If a project is set
			if _, exists := noodles.Projects[newProjectName]; !exists { // If the project isn't already set
				NewProjectPrompt(newProjectName) // Perform our project prompting
			} else { // Project is already set
				trunk.LogErr("Project is already defined. Please choose another project name.")
				os.Exit(1)
			}
		} else if newScriptName != "" { // If a script is set
			if _, exists := noodles.Scripts[newScriptName]; !exists { // If the script isn't already set
				NewScriptPrompt(newScriptName) // Perform our script prompting
			} else { // Script is already set
				trunk.LogErr("Script is already defined. Please choose another script name.")
				os.Exit(1)
			}
		}
	}
}

// NewWorkspacePrompt will handle the necessary workspace configuration prompts
func NewWorkspacePrompt() {
	properties := []string{"name", "description", "license", "version"}
	labels := []string{"Name of Workspace", "Description of Workspace", "License", "Version"}

	var promptProperties = map[string]string{} // Set promptProperties to an empty map

	for index, key := range properties {
		label := labels[index]

		var validate func(input string) error

		if key != "version" { // If we're not needing to use a validate func
			validate = func(input string) error {
				var err error

				if len(input) == 0 { // If there is no input string
					err = errors.New("a non-empty value is required for this field")
				}

				return err
			}
		} else {
			validate = func(input string) error {
				var err error
				_, convErr := strconv.ParseFloat(input, 64) // Attempt to just do a conversion

				if convErr != nil {
					err = errors.New("invalid version number")
				}

				return err
			}
		}

		val := TextPromptValidate(label, validate)
		promptProperties[key] = val // Set this property in promptProperties
	}

	noodles.Name = promptProperties["name"]
	noodles.Description = promptProperties["description"]
	noodles.License = promptProperties["license"]

	version, _ := strconv.ParseFloat(promptProperties["version"], 64) // Convert the version to a proper num
	noodles.Version = version

	if saveErr := SaveConfig(); saveErr == nil { // Save the config
		trunk.LogSuccess("Noodles workspace created.")
	} else { // Failed to save
		trunk.LogErrRaw(saveErr)
		return
	}
}

// NewProjectPrompt will handle the necessary project creation prompts
func NewProjectPrompt(newProjectName string) {
	pluginPrompt := promptui.Select{
		Label: "Plugin",
		Items: []string{"Go", "Less", "TypeScript"},
	}

	_, plugin, pluginPromptErr := pluginPrompt.Run() // Run our plugin selection
	PromptErrorCheck(pluginPromptErr)

	plugin = strings.ToLower(plugin)

	project := NoodlesProject{
		Plugin: strings.ToLower(plugin),
	}

	if plugin == "go" {
		GoProjectPrompt(plugin, newProjectName, &project)
	} else if plugin == "less" {
		SourceDestinationPrompt(plugin, &project)
		LESSProjectPrompt(&project)
	} else if plugin == "typescript" {
		SourceDestinationPrompt(plugin, &project)
		TypeScriptProjectPrompt(&project)
	}

	if noodles.Projects == nil { // Projects doesn't exist yet (no projects yet)
		noodles.Projects = make(map[string]NoodlesProject)
	}

	noodles.Projects[newProjectName] = project
	SaveConfig()
}

// GoProjectPrompt will provide the necessary project prompts for a Go project
func GoProjectPrompt(plugin string, name string, project *NoodlesProject) {
	typePrompt := promptui.Select{
		Label: "Type",
		Items: []string{"Binary", "Package", "Plugin"},
	}

	_, goType, typePromptErr := typePrompt.Run() // Run our plugin selection
	PromptErrorCheck(typePromptErr)

	goType = strings.ToLower(goType)

	if goType != "package" { // Binary or Plugin
		SourceDestinationPrompt(plugin, project) // Request the sources and destinations

		if goType == "plugin" { // Plugin
			if filepath.Ext(project.Destination) != ".so" {
				if project.Destination == "" { // If no destination is specified
					project.Destination = filepath.Join("build", name)
				}

				project.Destination = project.Destination + ".so" // Append .so
			}
		}
	} else { // Package
		pkgName := coreutils.InputMessage("Package name")
		project.SimpleName = pkgName // Set our requested package name as the simple name
	}

	project.Type = goType

	enableGoModules := TextPromptValidate("Enable Go Modules [y/N]", TextYNValidate)
	project.EnableGoModules = IsYes(enableGoModules)

	consolidateChildDirs := TextPromptValidate("Enable nested directories [y/N]", TextYNValidate)
	project.ConsolidateChildDirs = IsYes(consolidateChildDirs)

	enableNestedEnvironment := TextPromptValidate("Enable self-contained Go workspace (forces go/src directory usage) [y/N]", TextYNValidate)
	project.DisableNestedEnvironment = !IsYes(enableNestedEnvironment) // Invert our provided value, so if we're enabling (y) then mark to disable as false
}

// LESSProjectPrompt will provide the necessary project prompts for a LESS project
func LESSProjectPrompt(project *NoodlesProject) {
	appendHashVal := TextPromptValidate("Append SHA256SUM to end of file name [y/N]", TextYNValidate)
	project.AppendHash = IsYes(appendHashVal)
}

// SourceDestinationPrompt prompts for the sources and destinations for compilation
func SourceDestinationPrompt(plugin string, project *NoodlesProject) {
	source := TextPromptValidate("Source(s)", func(input string) error {
		return PromptExtensionValidate(plugin, input)
	})

	destination := coreutils.InputMessage("Destination")

	project.Destination = destination
	project.Source = source
}

// TypeScriptProjectPrompt will provide the necessary project prompts for a TypeScript project
func TypeScriptProjectPrompt(project *NoodlesProject) {
	appendHashVal := TextPromptValidate("Append SHA256SUM to end of file name [y/N]", TextYNValidate)
	project.AppendHash = IsYes(appendHashVal)

	isCompressVal := TextPromptValidate("Compress / Minified JavaScript [y/N]", TextYNValidate)
	project.Compress = IsYes(isCompressVal)

	modePrompt := promptui.Select{
		Label: "Compiler Options Mode",
		Items: []string{"simple", "advanced", "strict"}, // Our compiler modes
	}

	_, modePromptVal, modePromptErr := modePrompt.Run()
	PromptErrorCheck(modePromptErr)

	project.Mode = modePromptVal

	targetPrompt := promptui.Select{
		Label: "Target",
		Items: ValidTypeScriptTargets,
	}

	_, targetPromptVal, targetPromptErr := targetPrompt.Run()
	PromptErrorCheck(targetPromptErr)

	project.Target = targetPromptVal
}

// NewScriptPrompt will handle the necessary script creation prompts
func NewScriptPrompt(newScriptName string) {
	description := coreutils.InputMessage("Description") // Get the description of the project
	executable := TextPromptValidate("Executable", func(input string) error {
		var execExistsErr error

		if !coreutils.ExecutableExists(input) {
			execExistsErr = errors.New("executable does not exist")
		}

		return execExistsErr
	})

	argumentsString := coreutils.InputMessage("Arguments (comma separated, optional)")
	argumentsRaw := strings.SplitN(argumentsString, ",", -1) // Split our arguments by comma
	arguments := []string{}

	if len(argumentsRaw) != 0 { // If there are arguments
		for _, arg := range argumentsRaw {
			arguments = append(arguments, strings.TrimSpace(arg)) // Trim the space around this argument
		}
	}

	directory := TextPromptValidate("Directory to run script (default is noodles root directory)", func(input string) error {
		var dirExistsErr error

		if input == "" { // Default
			input = workdir
		}

		if !coreutils.IsDir(input) {
			dirExistsErr = errors.New("directory does not exist")
		}

		return dirExistsErr
	})

	redirectOutput := TextPromptValidate("Redirect output to file [y/N]", TextYNValidate)
	redirect := IsYes(redirectOutput)

	var file string

	if redirect { // If we should redirect output to a file
		file = coreutils.InputMessage("File path and name")
	}

	script := NoodlesScript{
		Arguments:   arguments,
		Description: description,
		Directory:   directory,
		Exec:        executable,
		File:        file,
		Redirect:    redirect,
	}

	if noodles.Scripts == nil { // Scripts doesn't exist yet (no projects yet)
		noodles.Scripts = make(map[string]NoodlesScript)
	}

	noodles.Scripts[newScriptName] = script
	SaveConfig()
}
