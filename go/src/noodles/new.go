package main

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"strconv"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a Noodles workspace, projects, or scripts",
	Long:  "Creates a Noodles workspace, projects, or scripts",
	Run:   new,
}

var newProjectName string
var newScriptName string

func init() {
	newCmd.Flags().StringVarP(&newProjectName, "project", "p", "", "Name of a new project you wish to bootstrap.")
	newCmd.Flags().StringVarP(&newScriptName, "script", "s", "", "Name of a new script you wish to bootstrap")
}

func new(cmd *cobra.Command, args []string) {
	if (newProjectName == "") && (newScriptName == "") { // If we're creating a new Noodles workspace
		if configInfo, statErr := os.Stat("noodles.toml"); statErr == nil { // Check if noodles.toml already exists
			if configInfo.Size() != 0 { // If the size of the file is greater than 0, meaning there is potentially content
				fmt.Println("noodles.toml already exists and appears to have content. Exiting.")
				return
			}
		}

		NewWorkspacePrompt() // Perform our workspace prompting
	} else {
		if noodles.Name == "" { // Noodles workspace doesn't seem to exist
			fmt.Println("No Noodles workspace appears to exist. Please create a workspace first.")
			os.Exit(1)
		}

		if newProjectName != "" { // If a project is set
			if _, exists := noodles.Projects[newProjectName]; !exists { // If the project isn't already set
				pluginPrompt := promptui.Select{
					Label: "Plugin",
					Items: []string{"Go", "Less", "TypeScript"},
				}

				_, plugin, pluginPromptErr := pluginPrompt.Run() // Run our plugin selection
				PromptErrorCheck(pluginPromptErr)

				source := TextPromptValidate("Source(s)", func(input string) error {
					return PromptExtensionValidate(plugin, input)
				})

				destination := coreutils.InputMessage("Destination")

				project := NoodlesProject{
					Destination: destination,
					Plugin:      plugin,
					Source:      source,
				}

				switch plugin {
				case "Go":
					GoProjectPrompt(&project)
					break
				case "LESS":
					LESSProjectPrompt(&project)
					break
				case "TypeScript":
					TypeScriptProjectPrompt(&project)
					break
				}

				noodles.Projects[newProjectName] = project
				SaveConfig()
			} else { // Project is already set
				fmt.Println("Project is already defined. Please choose another project name.")
				os.Exit(1)
			}
		} else if newScriptName != "" { // If a script is set
			if _, exists := noodles.Scripts[newScriptName]; !exists { // If the script isn't already set

			} else { // Script is already set
				fmt.Println("Script is already defined. Please choose another script name.")
				os.Exit(1)
			}
		}
	}
}

// NewWorkspacePrompt will handle the necessary workspace configuration prompts
func NewWorkspacePrompt() {
	var properties = map[string]string{ // Create a map of properties to ask for
		"name":        "Name of Workspace",
		"description": "Description of Workspace",
		"license":     "License",
		"version":     "Version",
	}

	var promptProperties = map[string]string{} // Set promptProperties to an empty map

	for key, label := range properties {
		var validate func(input string) error

		if key != "version" { // If we're not needing to use a validate func
			validate = func(input string) error {
				var err error

				if len(input) == 0 { // If there is no input string
					err = errors.New("A non-empty value is required for this field.")
				}

				return err
			}
		} else {
			validate = func(input string) error {
				var err error
				_, convErr := strconv.ParseFloat(input, 64) // Attempt to just do a conversion

				if convErr != nil {
					err = errors.New("Invalid Version Number.")
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
		fmt.Println("Noodles workspace now created!")
	} else { // Failed to save
		fmt.Println(saveErr.Error())
		return
	}
}

// GoProjectPrompt will provide the necessary project prompts for a Go project
func GoProjectPrompt(project *NoodlesProject) {
	isBinaryVal := TextPromptValidate("Is A Binary [y/N]", TextYNValidate)

	isBinary := (isBinaryVal == "y") || (isBinaryVal == "yes")
	project.Binary = isBinary
}

// LessProjectPrompt will provide the necessary project prompts for a LESS project
func LESSProjectPrompt(project *NoodlesProject) {
	appendHashVal := TextPromptValidate("Append SHA256SUM to end of file name [y/N]", TextYNValidate)

	project.AppendHash = (appendHashVal == "y") || (appendHashVal == "yes")
}

// TypeScriptProjectPrompt will provide the necessary project prompts for a TypeScript project
func TypeScriptProjectPrompt(project *NoodlesProject) {
	appendHashVal := TextPromptValidate("Append SHA256SUM to end of file name [y/N]", TextYNValidate)
	project.AppendHash = (appendHashVal == "y") || (appendHashVal == "yes")

	isCompressVal := TextPromptValidate("Compress / Minified JavaScript [y/N]", TextYNValidate)
	project.Compress = (isCompressVal == "y") || (isCompressVal == "yes")

	modePrompt := promptui.Select{
		Label: "Compiler Options Mode",
		Items: []string{"simple", "advanced", "strict"}, // Our compiler modes
	}

	_, modePromptVal, modePromptErr := modePrompt.Run()
	PromptErrorCheck(modePromptErr)

	project.Mode = modePromptVal

	targetPrompt := promptui.Select{
		Label: "Target",
		Items: []string{"ES5", "ES6", "ES7"},
	}

	_, targetPromptVal, targetPromptErr := targetPrompt.Run()
	PromptErrorCheck(targetPromptErr)

	project.Target = targetPromptVal
}
