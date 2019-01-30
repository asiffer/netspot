// introspect.go
package stats

import (
	"errors"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

var (
	StatDir string = path.Dir("./")
)

type StatDescription struct {
	Name  string
	Brief string
}

func parseLine(line string) map[string]string {
	var key string = ""
	for strings.Contains(line, "  ") {
		line = strings.Replace(line, "  ", " ", -1)
	}
	m := make(map[string]string)
	splittedLine := strings.Split(line, " ")
	for _, s := range splittedLine {
		if s[0] == 0x40 { // == '@'
			key = s[1:]
			m[key] = ""
		} else if key != "" {
			if m[key] == "" {
				m[key] = s
			} else {
				m[key] += " " + s
			}
		}
	}
	return m
}

func cleanComment(c string) string {
	chunk := strings.Replace(c, "/*", "", -1)
	chunk = strings.Replace(chunk, "//", "", -1)
	chunk = strings.Replace(chunk, "*/", "", -1)
	chunk = strings.TrimLeft(chunk, " ")
	return chunk
}

func getCommentsBeforeImports(path string) (string, error) {
	// comments := make([]string, 0)
	var comments string
	var chunk string
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return comments, err
	}
	fileBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return comments, err
	}
	fileContent := string(fileBytes)
	for _, d := range f.Comments {
		if d.End() <= f.Package {
			chunk = cleanComment(fileContent[d.Pos()-1 : d.End()])
			// comments = append(comments, chunk)
			comments += chunk
		} else {
			return comments, nil
		}

	}
	return comments, errors.New("No package declaration")
}

func parseStat(comment string) (StatDescription, error) {
	m := parseLine(comment)
	name, exists := m["name"]
	if !exists {
		return StatDescription{}, errors.New("@name field missing")
	}
	brief, exists := m["brief"]
	if !exists {
		return StatDescription{}, errors.New("@brief field missing")
	}
	return StatDescription{Name: name, Brief: brief}, nil
}

func parseFile(path string) (StatDescription, error) {
	comment, err := getCommentsBeforeImports(path)
	if err != nil {
		return StatDescription{}, err
	}
	sd, err := parseStat(comment)
	if err != nil {
		return sd, err
	}
	return sd, nil
}

func parseDirectory(dirpath string) []StatDescription {
	var filepath string
	sdList := make([]StatDescription, 0, 20)
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		filepath = path.Join(dirpath, file.Name())
		sd, err := parseFile(filepath)
		if err == nil {
			sdList = append(sdList, sd)
		}
	}
	return sdList
}

func GetAvailableStats() []StatDescription {
	return parseDirectory(StatDir)
}
