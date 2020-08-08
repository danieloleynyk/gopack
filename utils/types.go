package utils

type Package string

type Maintainer struct {
	Name string `yaml:"maintainerName"`
	Packages []Package
}

type Repository struct {
	Name string `yaml:"repositoryName"`
	Maintainers []Maintainer
}
