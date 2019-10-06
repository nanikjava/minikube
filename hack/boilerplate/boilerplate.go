package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var (
	cwd            string
	boilerplatedir *string
	rootdir        *string
	refs           map[string][]string
	skipped_dirs          = [...]string{"Godeps", "third_party", "gopath", "_output", ".git", "cluster/env.sh", "vendor", "test/e2e/generated/bindata.go"}
	filestoprocess []string
	m                                = make(map[string]*regexp.Regexp)
)

func init(){
	cwd , _ = os.Getwd()
	boilerplatedir  = flag.String("boilerplate-dir", cwd, "Boilerplate directory for boilerplate files")
	cwd = cwd + "/../../"
	rootdir  = flag.String("rootdir",  filepath.Dir(cwd), "Root directory to examine")
}

func main() {
	flag.Parse()
	if (*boilerplatedir == cwd){
		fmt.Println("-----same directory")
	}
	fmt.Println("-----> Template/boilerplate directory  = " + *boilerplatedir)
	fmt.Println("-----> Root directory = " + *rootdir)

	// the different regex that will be used for processing the files
	/*
		go_build_constraints --> to check whether the file contains // +build - <something_something>
		year --> to check whether the file contains the string YEAR instead of the real year
		date --> to check whether the file contain the year 2013,etc
		shebang --> used for .sh file to check whether it contains !/bin/sh string
	*/
	m["go_build_constraints"] = regexp.MustCompile(`(?m)^(// \+build.*\n)+\n`)
	m["year"] = regexp.MustCompile(`YEAR`)
	m["date"]= regexp.MustCompile(`(20[123]\d)`)
	m["shebang"]= regexp.MustCompile(`^(#!.*\n)\n*`)

	getRefs()
	getFiles()

	// process the file
	for _, filename := range filestoprocess {
		if (!isFileValid(filename)) {
			fmt.Printf("\nFilename = %s %s", filename , " is not valid")
		}
	}

	fmt.Printf("\nProcessing done....\n")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

/**
Function to check whether the processed file
is valid.

Returning false means that the file does not the
proper boilerplate template
*/
func isFileValid(filename string) (bool) {
	dat, err := ioutil.ReadFile(filename)
	check(err)

	filecontent := string(dat)
	filextension :=  filepath.Ext(filename)
	var re = regexp.MustCompile(`(?m)(.[^.]+)`)

	currentfilextension := re.FindString(filextension)
	templatecontent := refs[currentfilextension]
	currentRegex := m["go_build_constraints"]

	// if the file has .go extension than
	// use the go_build_constraints regex
	if (currentfilextension == ".go") {
		currentRegex = m["go_build_constraints"]
		if (currentRegex.MatchString(filecontent)) {
			// if it has a match then replace the matched string with empty string
			// this will remove the '+build' string
			filecontent = currentRegex.ReplaceAllString(filecontent, "")
		}
	}

	// use the shebang regex for .sh file
	if (currentfilextension == ".sh") {
		currentRegex = m["shebang"]
		if (currentRegex.MatchString(filecontent)) {
			// this will replace the !/bin/bash with empty string
			filecontent = currentRegex.ReplaceAllString(filecontent, "")
		}
	}

	// split the full string into array
	filecontentstringarray := strings.Split(filecontent, "\n")

	// check to make sure that the length of the data read from
	// the file is not less than template. If this case does
	// happen that means either the file is corrupt or
	// something else..
	if len(templatecontent) > len(filecontentstringarray) {
		fmt.Printf("\n -------->  The template is longer than the file ( %s )", filename)
		return false
	}

	// cut down the original content of the file read to make it the
	// same size like the template file
	filecontentstringarray = filecontentstringarray[:len(templatecontent)]

	// use the 'year' regex to check whether
	// the file contain the YEAR string
	currentRegex = m["year"]
	for _, content := range filecontentstringarray {
		if (currentRegex.MatchString(content)) {
			// if there is a match than means the template
			// is no good...
			return false
		}
	}

	// use the date regex to check for the year value
	currentRegex = m["date"]
	for ctr, content := range filecontentstringarray {
		if len(currentRegex.FindStringIndex(content)) > 0 {
			// if it found then replace it with the word YEAR.
			filecontentstringarray[ctr] = currentRegex.ReplaceAllString(content, "YEAR")
			break
		}
	}

	/*
		once the code reach here that means all regex has been executed so what's
		left is to check the complete string of the file and template to see
		if they are the same
	*/
	if (! assertEq(filecontentstringarray, templatecontent) ) {
		return false
	}

	// if everything is good that means the file does not need to
	// be worked on ....
	return true
}

/*
Function to just a normal deep equal check
between 2 different string array
*/
func assertEq(test []string, ans []string) bool {
	return reflect.DeepEqual(test, ans)
}

/**
Function to get all the files from the directory
*/
func getFiles() {
	var ff = func(fpath string, fileinfo os.FileInfo, err error) error {
		// first thing to do, check error. and decide what to do about it
		if err != nil {
			fmt.Printf("Error %v for path %s\n", err, fpath)
			return err
		}

		// check the see if the directory being send to this
		// function is in the skipped_dirs array
		for _, skipdir := range skipped_dirs {
			if fileinfo.IsDir() && fileinfo.Name() == skipdir {
				fmt.Printf("\nDirectory %s will be skipped", fileinfo.Name())
				// return with filepath.SkipDir so that this
				// directory will not be traversed into..
				return filepath.SkipDir
			}
		}

		// if directory is not to be skipped then check the file extension
		// check to make sure that this is not a directory
		if (! fileinfo.IsDir()) {
			// check if the extension exist in the ref variable
			filextension :=  filepath.Ext(fpath)
			var re = regexp.MustCompile(`(?m)(.[^.]+)`)
			for k,_:= range refs {

				// if the key (that contain the extension name)
				// is the same as the filename extension then we have
				// a match, so we need to add the filename to the
				// availabefiles variables to be processed later
				if ( k == re.FindString(filextension)) {
					filestoprocess = append(filestoprocess, fpath)
				}
			}
		}
		return nil
	}

	// walk through the rootdir path
	err := filepath.Walk(*rootdir, ff)

	// if there is an error report to the user
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", *rootdir, err)
	}

}

/**
This function is to populate the refs variable with the
different boilerplate/template for different extension. The
end result of the map will be like:

	.go --> <template_string>
	.sh --> <template_string>

*/
func getRefs(){
	matches, err := filepath.Glob(*boilerplatedir + "/boilerplate.*.txt")


	if err != nil {
		fmt.Println(err)
	}

	if len(matches) != 0 {
		refs=make(map[string][]string,len(matches))

		for i := 0; i < len(matches); i++ {
			var splitString []string
			// grab the filename only from the complete directory
			// eg: the directory will be <dir_name>/hack/boilerplate/boilerplate.go.txt
			_, file := filepath.Split(matches[i])

			// open the file
			_file, err := os.Open(matches[i])

			//TODO: Need to just display an error and
			//      exit
			if err != nil {
				fmt.Println("Error processing the boilerplate files. Exiting !")

			}
			defer _file.Close()

			// create scanner for reading the file
			s := bufio.NewScanner(_file)
			for s.Scan() {
				// read per line
				read_line := s.Text()

				// append it into the string
				splitString = append(splitString, read_line)
			}

			// regex to get the first . of the file as the filename
			// will be in the format eg: boilerplate.<lang>.txt
			// The first occurence of the . (dot) will be the extension
			// eg: boilerplate.go.txt
			// it will give us .go
			var re = regexp.MustCompile(`(?m)(.[^.]+)`)
			var matchall  =  re.FindAllString(file, -1)

			// get the extension of the file and use it as a key
			if len(matchall) > 0 {
				refs[matchall[1]] = splitString
			}
		}
	}
}