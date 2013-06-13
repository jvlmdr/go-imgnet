package imgnet

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

// Describes an ImageNet index constructed from a local copy.
// Provides fast access to the synset names and sizes.
// Provides a method to access an index of the images in a synset.
// Provides a method to access an image.
type Index struct {
	// Directory where ImageNet is located.
	Dir string
	// Maps synset name to size.
	Synsets map[string]int
}

// Describes an index of the images within a synset.
// Provides an unordered set of names.
type SynsetIndex struct {
	// Set of image names.
	Images map[string]struct{}
}

type Image struct {
	Synset string
	Name   string
}

// Creates an Index from a directory.
func BuildIndex(dir string, verbose bool) (*Index, error) {
	// Read the synset information from the subdirs.
	synsets, err := readSubdirs(dir, verbose)
	if err != nil {
		return nil, err
	}

	return &Index{dir, synsets}, nil
}

func (idx Index) SynsetIndex(synset string) (*SynsetIndex, error) {
	// List files in directory.
	addr := path.Join(idx.Dir, synset)
	files, err := readImagesInDir(addr)
	if err != nil {
		return nil, err
	}

	// Remove extension.
	re := regexp.MustCompile("^(.*)\\.JPEG$")
	for i, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 0 {
			err = fmt.Errorf("Could not extract image name from \"%s\"", file)
			return nil, err
		}
		files[i] = matches[1]
	}

	// Convert list to set.
	images := make(map[string]struct{}, len(files))
	for _, file := range files {
		images[file] = struct{}{}
	}

	return &SynsetIndex{images}, nil
}

func (idx Index) Open(im Image) (io.ReadCloser, error) {
	filename := path.Join(idx.Dir, im.Synset, fmt.Sprintf("%s.JPEG", im.Name))
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return file, nil
}

type Result struct {
	Name  string
	Size  int
	Error error
}

// Finds the subdirs below top dir and counts the images in each.
func readSubdirs(dir string, verbose bool) (map[string]int, error) {
	// Get a list of subdirectories.
	subdirs, err := readDirsInDir(dir)
	if err != nil {
		return nil, err
	}

	synsets := make(map[string]int, len(subdirs))

	// Initialize semaphore for limiting number of threads.
	const MaxProc = 32
	sem := make(chan int, MaxProc)
	for i := 0; i < MaxProc; i += 1 {
		sem <- 1
	}
	// Initialize channel for communicating results.
	ch := make(chan Result)

	// Index of next subdir to stat.
	i := 0
	// Number of subdirs completed.
	done := 0
	// Total number of images.
	var total int64

	for done < len(subdirs) {
		select {
		case <-sem:
			// Semaphore became free, start next subdir (if any remain).
			if i < len(subdirs) {
				go func(name string) {
					// Remember to release semaphore.
					defer func() { sem <- 1 }()
					// Stat subdirectories.
					addr := path.Join(dir, name)
					files, err := readImagesInDir(addr)
					ch <- Result{name, len(files), err}
				}(subdirs[i])
				// Move to next subdir.
				i += 1
			}

		case result := <-ch:
			// If there was an error, abort.
			if result.Error != nil {
				return nil, result.Error
			}
			// Otherwise update the index and increment the count.
			synsets[result.Name] = result.Size
			total += int64(result.Size)
			if verbose {
				fmt.Printf("%9d: \"%s\" %6d %12d\n", i, result.Name, result.Size, total)
			}
			done += 1
		}
	}

	return synsets, nil
}

// Returns a list of subdirectories below a directory.
func readDirsInDir(dir string) (subdirs []string, err error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if info.IsDir() {
			subdirs = append(subdirs, info.Name())
		}
	}

	return subdirs, nil
}

// Returns a list of the images in a directory.
func readImagesInDir(dir string) ([]string, error) {
	files, err := readFilesInDir(dir)
	if err != nil {
		return nil, err
	}

	// Check that file is an image.
	re := regexp.MustCompile("^.*\\.JPEG$")
	for _, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 0 {
			err = fmt.Errorf("Do not recognize \"%s\" as an image", file)
			return nil, err
		}
	}

	return files, nil
}

// Returns a list of the files in a directory.
func readFilesInDir(dir string) (files []string, err error) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		files = append(files, info.Name())
	}

	return files, nil
}
