package utils

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	copy2 "github.com/otiai10/copy"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCommand used for running a command and piping its output to stdout
func RunCommand(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmdReader, err := cmd.StderrPipe()
	Catch(err, "An error occurred while setting up stderr pipeline", false)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("\t gopack > %s\n", scanner.Text())
		}
	}()

	Catch(cmd.Start(), "An error occurred in command start", true)
	Catch(cmd.Wait(), "An error occurred in command wait", true)
}

func GetDirsAndFilesList(rootPath string) ([]string, error) {
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

// CopyDirectory is a wrapper around the copy function of github.com/otiai10/copy, for future changes
func CopyDirectory(srcPath, dstPath string) error {
	return copy2.Copy(srcPath, dstPath)
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func CompressTarball(src string, writers ...io.Writer) error {

	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	file, err := os.Create("qwe.tar")
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create tarball file '%s', got error '%s'", "out.tar", err.Error()))
	}
	defer file.Close()

	mw := io.MultiWriter(file)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		fmt.Println(file)
		f.Close()

		return nil
	})
}