package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopack/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
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

	for _, packageRepoName := range packages {

		// Generates package name
		r := regexp.MustCompile(`.*/(.*)$`)
		regexResult := r.FindStringSubmatch(packageRepoName)
		packageName := regexResult[len(regexResult) - 1]

		log.Printf("downloading %s\n", packageName)

		// Downloads the package from a public repository
		utils.RunCommand("/usr/bin/go", "get", "-d", "-v", packageRepoName)

		// Creates a temp dir for the package artifacts
		artifactsPath := path.Join(tempPath, packageName)
		downloadedPackagePath := path.Join(tempPath, "pkg", "mod")

		utils.Catch(
			os.Mkdir(artifactsPath, 0755),
			"An error occurred while creating artifacts dir",
			true,
		)

		reposList, err := getFileStructure(downloadedPackagePath)
		utils.Catch(
			err,
			"An error occurred while generating file structure",
			true,
		)

		utils.Catch(
			dumpRequirementsFile(reposList, artifactsPath),
			"An error occurred while dumping requirements",
			true,
		)
		log.Println("prepared requirements")

		utils.Catch(
			dumpPackages(downloadedPackagePath, artifactsPath),
			"An error occurred while dumping packages",
			true,
		)
		log.Println("prepared the package and its dependency")

		utils.CompressTarball(artifactsPath)

		fmt.Print("end")
	}
}

func dumpRequirementsFile(reposList []*utils.Repository, dumpPath string) error {
	requirementsByteArray, err := yaml.Marshal(&reposList)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		path.Join(dumpPath, utils.RequirementsFileName),
		requirementsByteArray,
		0644,
	)
	if err != nil {
		return err
	}

	return nil
}

func dumpPackages(downloadPackagesPath, artifactsPath string) error {
	return utils.CopyDirectory(downloadPackagesPath, artifactsPath)
}

func getFileStructure(rootPath string) ([]*utils.Repository, error) {
	var reposList []*utils.Repository

	repos, err := utils.GetDirsAndFilesList(rootPath)
	if err != nil {
		return nil, err
	}

	for _, repoName := range repos {
		if repoName != "github.com" {
			continue
		}

		var maintainersList []*utils.Maintainer

		maintainers, err := utils.GetDirsAndFilesList(path.Join(rootPath, repoName))
		if err != nil {
			return nil, err
		}

		for _, maintainerName := range maintainers {
			var packagesList []utils.Package

			packages, err := utils.GetDirsAndFilesList(path.Join(rootPath, repoName, maintainerName))
			if err != nil {
				return nil, err
			}

			for _, packageName := range packages {
				packagesList = append(packagesList, utils.Package(packageName))
			}

			maintainersList = append(maintainersList, &utils.Maintainer{
				Name: maintainerName,
				Packages: packagesList,
			})
		}

		reposList = append(reposList, &utils.Repository{
			Name: repoName,
			Maintainers: maintainersList,
		})
	}

	return reposList, nil
}
