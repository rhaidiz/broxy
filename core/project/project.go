package project

import (
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"path/filepath"
	"io/ioutil"
)

// PersistentProject represents a persistent project
type PersistentProject struct {
	project			*Project
	projectPath		string // the path to the project to disk (including projectName)
	projectName		string // the name of the project to disk
	isPersistent	bool
}

// Project represents the basic information of a project
type Project struct {
	Title string
	Path  string
}

var fileExtension = "broxy"
var fileSuffix = "settings"

func NewPersistentProject(t string, p string) (*PersistentProject, error) {

	projectName := t //fmt.Sprintf("%s.%s",t,fileExtension)
	projectPath := filepath.Join(p,projectName)

	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil{
		return nil, err
	}
	return &PersistentProject{
		projectPath: projectPath,
		projectName: projectName,
		project: &Project{Title:t, Path:p},
		isPersistent: false,
	},nil
}

// TODO: add error in case of opening. How do I even know if this is a broxy project?
func OpenPersistentProject(t string, p string) (*PersistentProject, error) {
	projectName := t
	projectPath := filepath.Join(p,projectName)
	// check if the project path actually exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, err
	}
	return &PersistentProject{
		projectPath: projectPath,
		projectName: projectName,
		project: &Project{Title:t, Path:p},
		isPersistent: false,
	}, nil
}

// SaveToFile provides modules the possibility of saving something to file in a JSON format
func (p *PersistentProject) SaveToFile(m string, stg interface{}) error {
	return p.saveToFile(m,"_",stg)
}

// LoadModuleSettings provides modules the possibility of loading something in JSON
func (p *PersistentProject) LoadFromFile(m string, stg interface{}) error {
	return p.loadFromFile(m,"_",stg)
}

// SaveSettings saves a setting file
func (p *PersistentProject) SaveSettings(m string, stg interface{}) error {
	return p.saveToFile(m,"settings_",stg)
}

// LoadSettings loads a setting file
func (p *PersistentProject) LoadSettings(m string, stg interface{}) error {
	return p.loadFromFile(m,"settings_",stg)
}

func (p *PersistentProject) saveToFile(m,t string, stg interface{}) error {
	b, err := json.MarshalIndent(stg,""," ")
	if err != nil {
		return err
	}
	// write to file
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s%s.json",t,strings.ToLower(m)))
	fmt.Println(fileName)
	return ioutil.WriteFile(fileName, b, 0644)
}

func (p *PersistentProject) loadFromFile(m,t string, stg interface{}) error {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s%s.json",t,strings.ToLower(m)))
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return json.Unmarshal(byteValue, &stg)
}

// Persist persists the project to disk in location pa
func (p *PersistentProject) Persist(pn, pa string) error {
	dest := filepath.Join(pa,pn)
	err := os.Rename(p.projectPath, dest)
	if err != nil {
		return err
	}
	p.isPersistent = true
	p.projectName = pn
	p.projectPath = dest
	return nil
}
