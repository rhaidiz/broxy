package repeater

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"log"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/core/project/decoder"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/rhaidiz/broxy/util"
)

// Controller represents the controller of the repeater module
type Controller struct {
	core.ControllerModule
	Module 		*Repeater
	Gui    		*Gui
	Sess   		*core.Session
	Tabs		map[int]*Tab
	TabEncoders map[int]*decoder.Encoder
}

// Tab describes a tab
type Tab struct {
	ID		int
	Title 	string
	Path 	string
	history []*TabContent
	count	int
}

// TabContent describes the content of a tab
type TabContent struct {
	Host		string
	Request 	string
	Response 	string
}

// Entry describes one entry of the textfile containing the history
type Entry struct {
	ID			int
	Timestamp	int64
	Type		string
	Data		[]byte
	Host		string
}

// NewController creates a new controller for the repeater module
func NewController(module *Repeater, gui *Gui, s *core.Session) *Controller {
	c := &Controller{
		Module: 		module,
		Gui:    		gui,
		Sess:   		s,
		Tabs: 			make(map[int]*Tab),
		TabEncoders: 	make(map[int]*decoder.Encoder),
	}
	c.Gui.GoClick = c.GoClick
	c.Gui.NewTabEvent = c.NewTab
	return c
}

// GetGui returns the Gui of the current controller
func (c *Controller) GetGui() core.GuiModule {
	return c.Gui
}

// GetModule returns the module of the current controller
func (c *Controller) GetModule() core.Module {
	return c.Module
}

func (c *Controller) NewTab(host, request string){
	t := &Tab{Title: fmt.Sprintf("%d",tabNum), ID: tabNum, Path: fmt.Sprintf("%d",tabNum), count: 1}
	tabContent := &TabContent{ Host: host, Request: request }
	t.history = append(t.history, tabContent)

	c.Tabs[tabNum] = t
	// save c.Tabs to file as settings
	c.Sess.PersistentProject.SaveSettings("repeater", c.Tabs)

	// get an encoder to save the history to file
	requestsEnc, _ := c.Sess.PersistentProject.FileEncoder2(fmt.Sprintf("tab_%d", t.ID))
	c.TabEncoders[t.ID] = &requestsEnc

	// save an entry in the file
	e := &Entry{ID:t.count, Type:"req", Host: host, Data: []byte(request)}
	(*c.TabEncoders[t.ID]).Encode(e)
	
	c.Gui.AddNewTab(t.ID, host, request)
	tabNum = tabNum + 1
}


// ExecCommand execs commands submitted by other modules
func (c *Controller) ExecCommand(m string, args ...interface{}) {
	if m == "send-to" {
		r := args[0].(*model.Request)
		//print(r.Host)
		//c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
		c.NewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
	}
}

// GoClick is the event fired when clicking the Go button in a repeater tab
func (c *Controller) GoClick(t *TabGui) {
	c.Sess.Debug(c.Module.Name(), "Go pressed")
	rRaw := util.NormalizeRequest(t.RequestEditor.ToPlainText())
	t.RequestEditor.SetPlainText(rRaw)

	r := strings.NewReader(rRaw)
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)
	var tabContent *TabContent
	tabContent = &TabContent{ Host: t.HostLine.Text(), Request: rRaw }
	if c.Tabs[t.id].history[0].Response == "" {
		// clear file content
		log.Println("create new file")
		c.Sess.PersistentProject.CreateFile(fmt.Sprintf("tab_%d", c.Tabs[t.id].ID))
		c.Tabs[t.id].history[0] = tabContent
	}else{
		c.Tabs[t.id].history = append(c.Tabs[t.id].history, tabContent)
	}
	e := &Entry{ID:c.Tabs[t.id].count, Type:"req", Host: tabContent.Host, Data: []byte(tabContent.Request)}
	(*c.TabEncoders[t.id]).Encode(e)

	if err != nil {
		c.Sess.Err(c.Module.Name(), fmt.Sprintf("ReadRequest %v", err))
		return
	}
	c.Sess.Debug(c.Module.Name(), req.Host)

	url, err := url.Parse(t.HostLine.Text())
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	req.RequestURI = ""

	go func(tab *Tab) {
		resp, err := c.Module.RunRequest(req)
		var respRaw string
		if err != nil {
			respRaw = fmt.Sprintf("RunRequest %v", err)
			c.Sess.Err(c.Module.Name(), fmt.Sprintf("RunRequest %v", err))
		} else {
			respRaw = util.ResponseToString(resp, false)
			t.ResponseEditor.SetPlainText(respRaw)
		}
		tab.history[tab.count-1].Response = respRaw
		
		entry := &Entry{ID: tab.count, Data: []byte(respRaw), Type: "resp" }
		(*c.TabEncoders[tab.ID]).Encode(entry)

		tab.count = tab.count + 1
	}(c.Tabs[t.id])
}
