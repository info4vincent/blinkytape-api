package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"

	blinky "github.com/wI2L/blinkygo"
)

//const ledStripWindowsComPort = "COM17"
//const ledStripLinuxComPort = "/dev/tty.usbmodem1421"
//const ledStripRPiComPort = "/dev/ttyACM0"

type Page struct {
	Title string
	Body  []byte
}

func loadPage(title string) *Page {
	filename := title + ".txt"
	body, _ := ioutil.ReadFile(filename)
	return &Page{Title: title, Body: body}
}
func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// title := r.URL.Path[len("/view/"):]
	p := loadPage(title)
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func startHandler(w http.ResponseWriter, r *http.Request, title string) {
	//	cmd := exec.Command("/Program Files (x86)/Windows Media Player/wmplayer", "c:\\go\\static\\Halloween.mp3")
	cmd := exec.Command("/usr/bin/omxplayer", "--vol -1000 /home/pi/Halloween.mp3")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	// log.Printf("Waiting for command to finish...")
	// err = cmd.Wait()
	log.Printf("Command started ...: %v", err)
}

func stripHandler(w http.ResponseWriter, r *http.Request, title string) {
	command := r.URL.Path[len("/strip/"):]
	result := "working..."
	fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", command, result)

	bt, err := blinky.NewBlinkyTape(ledStripComPort, 60)
	if err != nil {
		// log.Fatal(err)
		return
	}
	defer bt.Close()

	pattern, err := blinky.NewPatternFromImage(command, 114)
	anim := &blinky.Animation{
		Name:    "clyon",
		Repeat:  10,
		Speed:   5,
		Pattern: pattern,
	}

	bt.Play(anim, nil)

	time.Sleep(22 * time.Second)

	result = "done..."
	fmt.Fprintf(w, "<div>%s</div>", result)
}

var validPath = regexp.MustCompile("^/(strip|start|view)/([a-zA-Z0-9_.]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	bt, err := blinky.NewBlinkyTape(ledStripComPort, 60)
	if err != nil {
		log.Fatal(err)
	}
	//	defer bt.Close()

	white := blinky.NewRGBColor(255, 255, 255)
	err = bt.SetColor(white)
	err = bt.Render()

	time.Sleep(1 * time.Second)
	// err = bt.SwitchOff()
	// err = bt.Render()
	bt.Close()

	fmt.Printf("Listening on port 8010...")
	fmt.Printf("api is /strip/{filename}")
	fmt.Printf("api is /view/{filename}")
	fmt.Printf("api is /start/ will play halloween sound")

	http.HandleFunc("/strip/", makeHandler(stripHandler))
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/start/", makeHandler(startHandler))
	http.ListenAndServe(":8010", nil)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	fmt.Printf("Input Char Is : %v", string([]byte(input)[0]))
}
