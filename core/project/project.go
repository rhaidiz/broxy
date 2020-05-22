package project

import (
	"encoding/json"
	"fmt"
	"strings"
	"os"
	"path/filepath"
	"io/ioutil"

	"github.com/rhaidiz/broxy/core/project/decoder"
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
	return ioutil.WriteFile(fileName, b, 0644)
}

func (p *PersistentProject) loadFromFile(m,t string, stg interface{}) error {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s%s.json",t,strings.ToLower(m)))
	jsonFile, err := os.Open(fileName)
	if err != nil {
		// file does not exist, create it
		return p.saveToFile(m,t,stg)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	return json.Unmarshal(byteValue, &stg)
}

// FileEncoder provides an Encoder to write stuff to file
func (p *PersistentProject) FileEncoder(m string) (*json.Encoder, error) {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(m)))
	jsonFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return json.NewEncoder(jsonFile), nil
}

// FileDecoder provides a Decoder to read stuff from file
func (p *PersistentProject) FileDecoder(m string) (*json.Decoder, error) {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(m)))
	jsonFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return json.NewDecoder(jsonFile), nil
}

type FileDecoder struct {
	decoder.Decoder
	actualDecoder	*json.Decoder
}

func (d *FileDecoder) Decode(i interface{}) error {
	return d.actualDecoder.Decode(i)
}

func (d *FileDecoder) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

type FileEncoder struct {
	decoder.Encoder
	actualEncoder	*json.Encoder
}

func (d *FileEncoder) Encode(i interface{}) error {
	return d.actualEncoder.Encode(i)
}

func (d *FileEncoder) Marshal(v interface{}) ([]byte, error){
	return json.Marshal(v)
}

func (p *PersistentProject) FileDecoder2(m string) (decoder.Decoder, error){
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(m)))
	jsonFile, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileDecoder{actualDecoder: json.NewDecoder(jsonFile)},nil
}

func (p *PersistentProject) FileEncoder2(m string) (decoder.Encoder, error) {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(m)))
	jsonFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileEncoder{actualEncoder: json.NewEncoder(jsonFile)},nil
}

func (p *PersistentProject) CreateFile(f string) (*os.File, error){
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(f)))
	return os.Create(fileName)
}

func (p *PersistentProject) DeleteFile(f string) error {
	fileName := filepath.Join(p.projectPath, fmt.Sprintf("%s.json",strings.ToLower(f)))
	return os.Remove(fileName)
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
