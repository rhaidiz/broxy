package repeater

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rhaidiz/broxy/core"
	"github.com/rhaidiz/broxy/core/project/decoder"
	"github.com/rhaidiz/broxy/modules/coreproxy/model"
	"github.com/rhaidiz/broxy/util"
)

// Controller represents the controller of the repeater module
type Controller struct {
	core.ControllerModule
	Module				*Repeater
	Gui						*Gui
	Sess					*core.Session
	Tabs					map[int]*Tab
}

// Tab describes a tab
type Tab struct {
	ID				int
	Title			string
	Path			string
	history		[]*TabContent
	encoder		*decoder.Encoder
}

// TabContent describes the content of a tab
type TabContent struct {
	Host			string
	Timestamp	int64
	Request		string
	Response	string
}

// Entry describes one entry of the textfile containing the history
type Entry struct {
	ID					int
	Timestamp		int64
	Type				string
	Data				[]byte
	Host				string
}

// NewController creates a new controller for the repeater module
func NewController(module *Repeater, gui *Gui, s *core.Session) *Controller {

	c := &Controller{
		Module:			module,
		Gui:				gui,
		Sess:				s,
		Tabs:				make(map[int]*Tab),
	}
	tabNum = 1
	c.Gui.GoClick = c.GoClick
	c.Gui.NewTabEvent = c.NewTab
	c.Gui.Load = c.load
	c.Gui.RemoveTabEvent = c.removeTab
	c.Gui.GetStuff = c.getReqResp
	c.Gui.ChangeTabName = c.changeTabName
	return c
}

func (c *Controller) removeTab(id int){
	delete(c.Tabs, id)
	c.Sess.PersistentProject.SaveSettings("repeater", c.Tabs)
	c.Sess.PersistentProject.DeleteFile(fmt.Sprintf("tab_%d", id))
}

func (c *Controller) load(){
	// load settings
	var err error
	err = c.Sess.PersistentProject.LoadSettings("repeater", &c.Tabs)
	if err != nil {
		panic(fmt.Sprintf("Error while loading the repeater: %s", err))
	}
	for _,t := range c.Tabs{
		requestDec, err := c.Sess.PersistentProject.FileDecoder2(fmt.Sprintf("tab_%s", t.Path))
		if err != nil {
			// this if is meant to make sure that if there's an entry in the settings file
			// but no file associated, that tab is removed
			delete(c.Tabs, t.ID)
			err = c.Sess.PersistentProject.SaveSettings("repeater", c.Tabs)
			if err != nil {
				panic(fmt.Sprintf("Error while loading the repeater: %s", err))
			}
			continue
		}
		if t.ID > tabNum {
			tabNum = t.ID + 1
		}
		// load an encoder for this tab
		requestsEnc, err := c.Sess.PersistentProject.FileEncoder2(fmt.Sprintf("tab_%s", t.Path))
		if err != nil {
			panic(fmt.Sprintf("Error while loading the repeater: %s", err))
		}
		t.encoder = &requestsEnc
		for {
			e := &Entry{}
			if err := requestDec.Decode(&e); err != nil {
				break
			}
			if e.Type == "req"{
				// I'm reading a request
				if e.ID > len(t.history) - 1 {
					// first time I'm seeing this request
					tc := &TabContent{}
					tc.Host = e.Host
					tc.Request = string(e.Data)
					tc.Timestamp = e.Timestamp
					t.history = append(t.history, tc)
				}else{
					// second time I'm seeing this request, so overwrite.
					// this actually shoudn't happen, it's here just because shit happens
					req := string(e.Data)
					t.history[e.ID].Request = req
					t.history[e.ID].Host = e.Host
				}
			}else if e.Type == "resp" {
				// I'm reading a response
				resp := string(e.Data)
				t.history[e.ID].Response = resp
			}
		}
		c.new(t)
	}
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
	t := &Tab{Title: fmt.Sprintf("%d",tabNum), ID: tabNum, Path: fmt.Sprintf("%d",tabNum)}
	tabContent := &TabContent{ Host: host, Request: request, Timestamp: time.Now().Unix()}
	t.history = append(t.history, tabContent)
	c.Tabs[tabNum] = t
	tabNum = tabNum + 1
	c.new(t)
}

// this method creates a new GUI tab based on a Tab struct
func (c *Controller) new(t *Tab){

	if len(t.history) == 0 {
		t.history = append(t.history, &TabContent{})
	}

	lastItemIndex := len(t.history) - 1
		h := t.history[lastItemIndex].Host
		rq := t.history[lastItemIndex].Request
		rp := t.history[lastItemIndex].Response

		c.Gui.AddNewTab(t.Title, t.ID, h, rq, rp)
		for i, _ := range t.history {
		//for i := len(t.history)-1; i >= 0; i-- {
			tabContent := t.history[i]
			timeFormatted := time.Unix(tabContent.Timestamp, 0).Format("2006-01-02 15:04:05")
			c.Gui.AddToHistory(t.ID, i, fmt.Sprintf("%d. %s", i, timeFormatted ))
		}
}

func (c *Controller) getReqResp(idTab, idContent int)(string, string, string){
	t := c.Tabs[idTab].history[idContent]
	return t.Host, t.Request, t.Response
}


// ExecCommand execs commands submitted by other modules
func (c *Controller) ExecCommand(m string, args ...interface{}) {
	if m == "send-to" {
		r := args[0].(*model.Request)
		//c.Gui.AddNewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))
		c.NewTab(fmt.Sprintf("%s://%s", r.URL.Scheme, r.Host), fmt.Sprintf("%s\n", r.ToString()))

	}
}

func (c *Controller) changeTabName(id int,s string){
	c.Tabs[id].Title = s
	c.Sess.PersistentProject.SaveSettings("repeater", c.Tabs)
}

// GoClick is the event fired when clicking the Go button in a repeater tab
//func (c *Controller) GoClick(t *TabGui) {
func (c *Controller) GoClick(id int, host, request string, ch chan string) {
	t := c.Tabs[id]

	rRaw := util.NormalizeRequest(request)

	r := strings.NewReader(rRaw)
	buf := bufio.NewReader(r)

	req, err := http.ReadRequest(buf)
	if err != nil {
		c.Sess.Err(c.Module.Name(), fmt.Sprintf("ReadRequest %v", err))
		return
	}

	var tabContent *TabContent
	tabContent = &TabContent{ Host: host, Request: rRaw }

	timestamp := time.Now().Unix()
	timeFormatted := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")

	if t.history[0].Response == "" {
		// the tab has no history so it's a new thing to save to file
		// save c.Tabs to file as settings
		c.Sess.PersistentProject.SaveSettings("repeater", c.Tabs)

		requestsEnc, err := c.Sess.PersistentProject.FileEncoder2(fmt.Sprintf("tab_%d", t.ID))
		if err != nil {
			panic("Error while saving repeater history")
		}
		t.encoder = &requestsEnc

		t.history[0] = tabContent
	}else{
		t.history = append(t.history, tabContent)
		// I'm duplicating this part but whatever
		entryId := len(t.history) - 1
		c.Gui.AddToHistory(t.ID, entryId, fmt.Sprintf("%d. %s", entryId, timeFormatted))
	}

	entryId := len(t.history) - 1
	e := &Entry{ID: entryId, Type:"req", Host: host, Data: []byte(rRaw), Timestamp: timestamp}
	(*t.encoder).Encode(e)

	c.Sess.Debug(c.Module.Name(), req.Host)

	url, err := url.Parse(host)
	req.URL.Scheme = url.Scheme
	req.URL.Host = url.Host
	req.RequestURI = ""

	go func(i int, t *Tab, ch chan string) {
		resp, err := c.Module.RunRequest(req)
		var respRaw string
		if err != nil {
			c.Sess.Err(c.Module.Name(), fmt.Sprintf("RunRequest %v", err))
		} else {
			respRaw = util.ResponseToString(resp, false)
			ch <- respRaw
		}
		t.history[i].Response = respRaw

		entry := &Entry{ID: i, Data: []byte(respRaw), Type: "resp" }
		(*t.encoder).Encode(entry)

	}(entryId, t, ch)

}
