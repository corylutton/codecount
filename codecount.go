/*
###############################################################################
# Copyright 2014 Cory Lutton                                                  #
#                                                                             #
# Licensed under the Apache License, Version 2.0 (the "License");             #
# you may not use this file except in compliance with the License.            #
# You may obtain a copy of the License at                                     #
#                                                                             #
#    http://www.apache.org/licenses/LICENSE-2.0                               #
#                                                                             #
# Unless required by applicable law or agreed to in writing, software         #
# distributed under the License is distributed on an "AS IS" BASIS,           #
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.    #
# See the License for the specific language governing permissions and         #
# limitations under the License.                                              #
###############################################################################
*/
// Count the lines in text type files recursively though a directory
// based on the extension of the file.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

var (
	VERSION     = "0.2"
	ROOT        = string(".")
	ARG_JSON    = flag.Bool("json", false, "Output JSON")
	ARG_VERSION = flag.Bool("v", false, "Display Version")
	ARG_BYFILE  = flag.Bool("f", false, "Report by File")
	ARG_BYPATH  = flag.Bool("p", false, "Report by Path")
	ARG_DEBUG   = flag.Bool("d", false, "Enable Debug output")
	ARG_INCLUDE = flag.Bool("i", false, "Include Duplicate Files")
	ARG_PROFILE = flag.String("cpuprofile", "", "Write cpu profile to file")
	ARG_MEMORY  = flag.String("memprofile", "", "Write mem profile to file")
)

type File struct {
	path     string      // Path of the file
	info     os.FileInfo // Complete file info returned by ioutil
	lang     Language    // Language
	scanned  bool        // Was this scanned
	lines    int         // Total Lines
	comments int         // Comment Lines
	blanks   int         // Blank Lintes
	code     int         // Code Lines
	// hash     string      // Hash of the contents for duplicate filtering
}

type Files []File

func (f Files) Len() int           { return len(f) }
func (f Files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Files) Less(i, j int) bool { return f[i].lines > f[j].lines }

type FileByLang struct{ Files }

func (f FileByLang) Less(i, j int) bool {
	return f.Files[i].lang.name < f.Files[j].lang.name
}

type FileByPath struct{ Files }

func (f FileByPath) Less(i, j int) bool {
	return f.Files[i].path < f.Files[j].path
}

type Language struct {
	name       string   // Print name
	extension  []string // File Extensions
	openblock  string   // Block comment opening
	closeblock string   // Block comment closing
	comment    string   // Line comment markers
	endmark    string   // End of code marker
}
type Languages []Language

var languages = Languages{
	Language{"Assembly", []string{".s"}, "", "", ";", ""},
	Language{"Batch", []string{".bat"}, "", "", "REM", ""},
	Language{"C", []string{".c"}, "/*", "*/", "//", ""},
	Language{"C++", []string{".cpp"}, "/*", "*/", "//", ""},
	Language{"C/C++ Header", []string{".h"}, "/*", "*/", "//", ""},
	Language{"CSS", []string{".css"}, "/*", "*/", "", ""},
	Language{"C#", []string{".cs"}, "/*", "*/", "//", ""},
	Language{"Go", []string{".go"}, "/*", "*/", "//", ""},
	Language{"HTML", []string{".html", ".htm"}, "", "", "", ""},
	Language{"Java", []string{".java"}, "/*", "*/", "//", ""},
	Language{"Javascript", []string{".js"}, "/*", "*/", "//", ""},
	Language{"JSON", []string{".json"}, "", "", "", ""},
	Language{"Markdown", []string{".md"}, "", "", "", ""},
	Language{"Perl", []string{".pl"}, "/*", "*/", "//", "__END__"},
	Language{"PHP", []string{".php"}, "/*", "*/", "//", "__halt_compiler()"},
	Language{"Python", []string{".py", ".pyw"}, "", "", "#", ""},
	Language{"RestructuredText", []string{".rst"}, "", "", "", ""},
	Language{"RPGLE", []string{".rpgle"}, "", "", "", ""},
	Language{"Ruby", []string{".rb"}, "/*", "*/", "#", "__END__"},
	Language{"Rust", []string{".rs"}, "/*", "*/", "//", ""},
	Language{"SQL", []string{".sql"}, "/*", "*/", "", ""},
	Language{"TCL", []string{".tcl"}, "", "", "#", ""},
	Language{"Text", []string{".txt"}, "", "", "", ""},
	Language{"VB", []string{".vb", ".mac", ".frm", ".frx", ".bas"}, "/*", "*/", "'", ""},
	Language{"XML", []string{".xml", ".xss", ".xsc", ".xsd", ".xsx"}, "", "", "", ""},
}

// Setup the set of extension types to scan
var extensions = func() map[string]Language {
	ext_set := map[string]Language{}
	for _, lang := range languages {
		for _, ext := range lang.extension {
			ext_set[ext] = lang
		}
	}
	return ext_set
}()

// States for scanning
const (
	NORMAL = iota
	BLOCK
	END
)

var files = []File{}

// Run the codecounter
func main() {
	file_count := 0
	blank_count := 0
	comment_count := 0
	code_count := 0
	line_count := 0
	start := time.Now()
	flag.Parse()
	args := flag.Args()
	if len(args) == 1 {
		ROOT = args[0]
	}

	if *ARG_VERSION {
		fmt.Printf("Codecount %s\n", VERSION)
		return
	}

	if *ARG_PROFILE != "" {
		f, err := os.Create(*ARG_PROFILE)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Collect the files or single file
	filepath.Walk(ROOT, walkFunc)

	// Total scanned files
	for i := 0; i < len(files); i++ {
		if files[i].scanned {
			file_count++
			blank_count = blank_count + files[i].blanks
			comment_count = comment_count + files[i].comments
			code_count = code_count + files[i].code
			line_count = line_count + files[i].lines
		}
	}

	if *ARG_JSON {
		json.NewEncoder(os.Stdout).Encode(files)
	} else {
		reportHeader()
		reportDetail(files)

		end := time.Now()
		fmt.Println(strings.Repeat("-", 79))
		fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
			"Totals",
			file_count,
			blank_count,
			comment_count,
			code_count,
			line_count)
		fmt.Println(strings.Repeat("-", 79))
		fmt.Println("Runtime: ", end.Sub(start))
	}

	if *ARG_MEMORY != "" {
		f, err := os.Create(*ARG_MEMORY)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
		return
	}
}

// Create the files
func walkFunc(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if path == "." {
			return nil
		}
		name := info.Name()
		if strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		} else if name == "__pycache__" {
			return filepath.SkipDir
		}
	} else {
		ext := filepath.Ext(path)
		if _, found := extensions[ext]; found {
			file := File{path: path, info: info}
			file.scan()
			files = append(files, file)
		}
	}
	return nil
}

// Scans a single file, recording the stats
func (file *File) scan() {
	state := NORMAL
	file.lang = extensions[strings.ToLower(filepath.Ext(file.path))]

	// Skip unknown files
	if file.lang.name == "" || file.info.Size() == 0 {
		file.scanned = false
		return
	}

	// Open the file to begin scanning
	f, err := os.Open(file.path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Read line by line of the file to classify
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		line_orig := scanner.Text()
		file.lines++

		line := strings.TrimSpace(line_orig)
		if line == "" {
			file.blanks++
			if *ARG_DEBUG {
				fmt.Printf("BLNK\t%s\n", line_orig)
			}
			continue
		}

		/* In each line take the current state and decide
		if conditions for another state have come up.
		Start with the NORMAL state, if a block is opened
		but not closed then the next line will resume in the
		BLOCK state.  If the end marker has been encountered,
		all further lines are in the END state.
		*/
		switch state {
		case NORMAL:
			if strings.HasPrefix(line, file.lang.comment) &&
				file.lang.comment != "" {

				file.comments++
				if *ARG_DEBUG {
					fmt.Printf("LCOM\t%s\n", line_orig)
				}
				continue
			}

			if strings.HasPrefix(line, file.lang.endmark) &&
				file.lang.endmark != "" {

				state = END
				file.comments++
				if *ARG_DEBUG {
					fmt.Printf("ECOM\t%s\n", line_orig)
				}
				continue
			}

			if file.lang.openblock != "" &&
				file.lang.closeblock != "" {

				spos := strings.LastIndex(line, file.lang.openblock)
				epos := strings.LastIndex(line, file.lang.closeblock)

				if spos > epos && spos > 1 {
					state = BLOCK
					file.code++
					if *ARG_DEBUG {
						fmt.Printf("COCM\t%s\n", line_orig)
					}
					continue
				} else if spos > epos {
					state = BLOCK
					file.comments++
					if *ARG_DEBUG {
						fmt.Printf("OCOM\t%s\n", line_orig)
					}
					continue
				} else if spos < epos && spos > -1 {
					state = NORMAL
					file.comments++
					if *ARG_DEBUG {
						fmt.Printf("BCOM\t%s\n", line_orig)
					}
					continue
				}
			}

			file.code++
			if *ARG_DEBUG {
				fmt.Printf("CODE\t%s\n", line_orig)
			}

		case BLOCK:
			spos := strings.LastIndex(line, file.lang.openblock)
			epos := strings.LastIndex(line, file.lang.closeblock)

			if spos < epos && epos != -1 {
				state = NORMAL
				if *ARG_DEBUG {
					fmt.Printf("CCOM\t%s\n", line_orig)
				}
			} else {
				if *ARG_DEBUG {
					fmt.Printf("BCOM\t%s\n", line_orig)
				}
			}
			file.comments++

		case END:
			file.comments++
			if *ARG_DEBUG {
				fmt.Printf("ECOM\t%s\n", line_orig)
			}
		}

	}
	file.scanned = true
}

func (file File) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Code     int    `json:"code"`
		Blanks   int    `json:"blanks"`
		Comments int    `json:"comments"`
		Lines    int    `json:"lines"`
		Language string `json:"language"`
	}{
		Name:     file.info.Name(),
		Path:     file.path,
		Code:     file.code,
		Lines:    file.lines,
		Blanks:   file.blanks,
		Comments: file.comments,
		Language: file.lang.name,
	})
}

// Print the report
func reportDetail(files Files) {
	if *ARG_BYFILE {
		sort.Sort(files)
		for i := 0; i < len(files); i++ {
			if !files[i].scanned {
				continue
			}
			name := files[i].info.Name()
			if len(name) > 29 {
				name = name[0:27] + ".."
			}
			fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
				name,
				1,
				files[i].blanks,
				files[i].comments,
				files[i].code,
				files[i].lines)
		}
	} else if *ARG_BYPATH {
		path := ""
		count, blanks, comments, code, lines := 0, 0, 0, 0, 0
		sort.Sort(FileByPath{files})
		for i := 0; i < len(files); i++ {
			if !files[i].scanned {
				continue
			}
			if path == "" {
				path = files[i].path
			}
			if path != files[i].path {
				if len(path) > 29 {
					path = path[0:10] + "..." + path[len(path)-16:]
				}
				fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
					path,
					count,
					blanks,
					comments,
					code,
					lines)
				path = files[i].path
				count = 1
				blanks = files[i].blanks
				comments = files[i].comments
				code = files[i].code
				lines = files[i].lines
			} else {
				count++
				blanks = files[i].blanks + blanks
				comments = files[i].comments + comments
				code = files[i].code + code
				lines = files[i].lines + lines
			}

		}
		fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
			path,
			count,
			blanks,
			comments,
			code,
			lines)
	} else {
		lang := ""
		count, blanks, comments, code, lines := 0, 0, 0, 0, 0
		sort.Sort(FileByLang{files})
		for i := 0; i < len(files); i++ {
			if !files[i].scanned {
				continue
			}
			if lang == "" {
				lang = files[i].lang.name
			}
			if lang != files[i].lang.name {
				fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
					lang,
					count,
					blanks,
					comments,
					code,
					lines)
				lang = files[i].lang.name
				count = 1
				blanks = files[i].blanks
				comments = files[i].comments
				code = files[i].code
				lines = files[i].lines
			} else {
				count++
				blanks = files[i].blanks + blanks
				comments = files[i].comments + comments
				code = files[i].code + code
				lines = files[i].lines + lines
			}

		}
		fmt.Printf("%-29s%10d%10d%10d%10d%10d\n",
			lang,
			count,
			blanks,
			comments,
			code,
			lines)
	}
}

func reportHeader() {
	fmt.Printf("Codecount - v %s\n", VERSION)
	fmt.Println(strings.Repeat("-", 79))
	fmt.Printf("%-29s%10s%10s%10s%10s%10s\n",
		"Grouping", "Files", "Blank", "Comment", "Code", "Lines")
	fmt.Println(strings.Repeat("-", 79))
}
