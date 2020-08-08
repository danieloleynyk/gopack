package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

var downloadCmd = &cobra.Command{
	Use: "download",
	Short: "Downloads a tar.gz of the package",
	Long: "Downloads a tar.gz of the package and the setting needed to install it later with gopack",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		download(args)
	},
}

func download(packages []string) {
	tempPath, err := ioutil.TempDir("/tmp", "gopack")
	utils.Catch(err, "An error occurred while initializing temp directory", true)
	utils.Catch(os.Setenv("GOPATH", tempPath),"An error occurred while setting temp gopath", true)

	defer func() {
		utils.RunCommand("/usr/bin/chmod", "-R", "777", tempPath)
		utils.Catch(os.RemoveAll(tempPath),"An error occurred while removing temp directory",true)
	}()

	for _, packageName := range packages {
		// Downloads the package from a public repository
		utils.RunCommand("/usr/bin/go", "get", "-d", "-v", packageName)

		reposList, err := getFileStructure(fmt.Sprintf("%s/pkg/mod", tempPath))
		utils.Catch(err, "An error occurred while generating file structure", true)

		d, err := yaml.Marshal(&reposList)
		if err != nil {
			log.Fatalf("error: %v", err)
		}

		utils.Catch(ioutil.WriteFile("/tmp/dat1", d, 0644), "An error occurred while writing dependencies to file", true)
	}
}

func getFileStructure(rootPath string) ([]utils.Repository, error) {
	var reposList []utils.Repository

	repos, err := getDirsAndFilesList(rootPath)
	if err != nil {
		return nil, err
	}

	for _, repoName := range repos {
		if repoName != "github.com" {
			continue
		}

		var maintainersList []utils.Maintainer

		maintainers, err := getDirsAndFilesList(fmt.Sprintf("%s/%s", rootPath, repoName))
		if err != nil {
			return nil, err
		}

		for _, maintainerName := range maintainers {
			var packagesList []utils.Package

			packages, err := getDirsAndFilesList(fmt.Sprintf("%s/%s/%s", rootPath, repoName, maintainerName))
			if err != nil {
				return nil, err
			}

			for _, packageName := range packages {
				packagesList = append(packagesList, utils.Package(packageName))
			}

			maintainersList = append(maintainersList, utils.Maintainer{
				Name: maintainerName,
				Packages: packagesList,
			})
		}

		reposList = append(reposList, utils.Repository{
			Name: repoName,
			Maintainers: maintainersList,
		})
	}

	return reposList, nil
}

func getDirsAndFilesList(rootPath string) ([]string, error) {
	var filesList []string

	files, err := ioutil.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filesList = append(filesList, file.Name())
	}

	return filesList, nil
}