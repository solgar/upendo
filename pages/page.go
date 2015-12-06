package pages

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"
	"upendo/router"
	"upendo/session"
	"upendo/settings"
)

const (
	stateChunk = iota
	statePrefix
	stateName
	stateSuffix

	PartTypeHTML     = 10
	PartTypeTemplate = 11
	PartTypeFunction = 12
	PartTypeInclude  = 13
)

var (
	funcMap       map[string]interface{} = template.FuncMap{}
	TemplatesRoot *template.Template
)

func LoadTemplates(directory string) {
	funcMap["roleOrHigher"] = session.RoleOrHigher
	funcMap["roleOrLower"] = session.RoleOrLower
	funcMap["redirect"] = router.Redirect
	var err error
	TemplatesRoot, err = template.New("root").Funcs(funcMap).ParseGlob(settings.StartDir + directory + "/*.*")
	if err != nil {
		panic(err)
	}
	TemplatesRoot.Funcs(funcMap)
}

func RegisterFunction(name string, function interface{}) {
	_, ok := funcMap[name]
	if ok {
		fmt.Printf("Error: function named \"%s\" already exists!", name)
	} else {
		funcMap[name] = function
	}
}

type PagePart struct {
	content string
	type_   int
	params  map[string]string
}

func (ptp *PagePart) Params() map[string]string {
	return ptp.params
}

func (ptp *PagePart) Content() string {
	return ptp.content
}

func (ptp *PagePart) Type() int {
	return ptp.type_
}

type Page struct {
	Title string
	Parts []PagePart
}

// <?go sampleFunction param1=value param2="value too" param3=some"string ?>
// <?include templateName ?>

func tidyUpWhitespaces(s string) string {
	s = strings.Trim(s, " \t\n")
	s = strings.Replace(s, "\t", " ", -1)
	s = strings.Replace(s, "\n", " ", -1)

	for strings.Contains(s, "  ") {
		s = strings.Replace(s, "  ", " ", -1)
	}

	return s
}

func extractParams(rawString string) (map[string]string, error) {
	params := make(map[string]string)
	kStart := -1
	kEnd := -1
	vStart := -1
	vEnd := -1
	expectEquals := false
	quotedValue := false
	i := 0
	for ; i < len(rawString); i++ {
		c := rawString[i]
		isWitespace := c == ' ' || c == '\n' || c == '\t'
		if kStart == -1 { // find start of a key
			if isWitespace {
				continue
			}
			kStart = i
			kEnd = vStart + 1
		} else if kStart != -1 && vStart == -1 && vEnd == -1 { // key start found, search for its end
			expectEquals = expectEquals || isWitespace
			if isWitespace {
				continue
			} else if c != '=' && expectEquals {
				return params, errors.New(`Error parsing params "` + rawString + `" around char ` + strconv.Itoa(i))
			} else if c == '=' {
				expectEquals = false
				vStart = -2
				kEnd += 1
			} else if !expectEquals {
				kEnd = i
			}
		} else if vStart == -2 {
			if isWitespace {
				continue
			} else if c == '"' {
				quotedValue = true
				vStart = i + 1
			} else {
				vStart = i
			}
		} else if vStart != -1 && vEnd == -1 {
			if quotedValue {
				if c == '\\' && (i+1 < len(rawString)) && rawString[i+1] == '"' {
					rawString = rawString[:i] + rawString[i+1:]
				} else if c == '"' {
					vEnd = i
				} else if i == len(rawString)-1 {
					return params, errors.New(`Quoted param "` + rawString[kStart:kEnd] + `" doesn't have terminating quote.`)
				}
			} else if isWitespace {
				vEnd = i
			} else if i == len(rawString)-1 {
				vEnd = i + 1
			}
		}

		if kStart != -1 && kEnd != -1 && vStart != -1 && vEnd != -1 {
			quotedValue = false
			key := rawString[kStart:kEnd]
			value := rawString[vStart:vEnd]
			_, exists := params[key]
			if exists {
				return params, errors.New(`Redefinition of value "` + key + `" around char ` + strconv.Itoa(i))
			}
			params[key] = value
			kStart = -1
			kEnd = -1
			vStart = -1
			vEnd = -1
		}
	}

	if kStart != -1 || kEnd != -1 || vStart != -1 || vEnd != -1 {
		return params, errors.New("Invalid state after parsing parameters.")
	}

	return params, nil
}

func processNormalChunk(chunk string) PagePart {
	if strings.Contains(chunk, "{{") && strings.Contains(chunk, "}}") {
		return PagePart{chunk, PartTypeTemplate, make(map[string]string)}
	} else {
		return PagePart{chunk, PartTypeHTML, make(map[string]string)}
	}
}

func LoadPageTemplate(name string) (*Page, error) {
	fmt.Println("Loading template:", name)
	rawData, err := ioutil.ReadFile(settings.StartDir + "templates/" + name + ".html")
	if err != nil {
		panic(err)
		return nil, err
	}

	parts := make([]PagePart, 0)
	html := string(rawData)

	chunks := strings.Split(html, "<?")

	for i, chunk := range chunks {
		if i == 0 && len(chunk) == 0 {
			continue
		}
		if i == 0 && len(chunk) > 0 {
			parts = append(parts, processNormalChunk(chunk))
			continue
		}

		subchunks := strings.Split(chunk, "?>")

		if len(subchunks) == 1 {
			if i == len(chunks)-1 {
				return nil, errors.New("Error while processing template \"" + name + "\" => No closing ?>")
			}
			return nil, errors.New("Error while processing template \"" + name + "\" => Nested <?")
		} else if len(subchunks) > 2 {
			return nil, errors.New("Error while processing template \"" + name + "\" => Orphaned ?>")
		}

		subchunks[0] = strings.TrimLeft(subchunks[0], " \t\n")
		firstSpaceIdx := strings.IndexAny(subchunks[0], " \t\n")
		actionName := subchunks[0][:firstSpaceIdx]

		if actionName == "go" {
			functionAndParameters := strings.Trim(subchunks[0][firstSpaceIdx+1:], " \t\n")
			firstSpaceIdx := strings.IndexAny(functionAndParameters, " \t\n")
			var parameters map[string]string
			var err error
			functionName := ""
			if firstSpaceIdx == -1 {
				functionName = functionAndParameters
			} else {
				functionName = functionAndParameters[:firstSpaceIdx]
				parameters, err = extractParams(functionAndParameters[firstSpaceIdx+1:])
			}

			if err != nil {
				panic(err)
			}

			parts = append(parts, PagePart{functionName, PartTypeFunction, parameters})
			parts = append(parts, processNormalChunk(subchunks[1]))
		} else if actionName == "include" {
			templateName := subchunks[0][firstSpaceIdx+1:]
			templateName = tidyUpWhitespaces(templateName)
			if len(templateName) == 0 || strings.Index(templateName, " ") != -1 {
				return nil, errors.New("Error while processing template \"" + name + "\" => Invalid include action.")
			}

			loadedTemplate, err := LoadPageTemplate(templateName)

			if err != nil {
				panic(err)
			}

			parts = append(parts, loadedTemplate.Parts...)
			parts = append(parts, processNormalChunk(subchunks[1]))

		} else {
			return nil, errors.New("Error while processing template \"" + name + "\" => Unknown action name \"" + actionName + "\"")
		}
	}

	return &Page{name, parts}, nil
}
