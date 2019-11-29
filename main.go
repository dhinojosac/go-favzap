package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"github.com/Rhymen/go-whatsapp/binary/proto"
	"github.com/gookit/color"

	"github.com/gen2brain/beeep"
)

var ctime time.Time
var historyMessages []whatsapp.TextMessage //TODO:implement and order history messages
//Alerts
var soundAlert bool
var notifAlert bool

// ByTime implements sort.Interface based on the timestamp field.
type byTime []whatsapp.TextMessage

func (a byTime) Len() int           { return len(a) }
func (a byTime) Less(i, j int) bool { return a[i].Info.Timestamp < a[j].Info.Timestamp }
func (a byTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Toogle function
func toggleAlert(t int) {
	if t == 0 {
		if soundAlert {
			soundAlert = false
			fmt.Printf("Sound Alert: Disabled\n")
		} else {
			soundAlert = true
			fmt.Printf("Sound Alert: Enabled\n")
		}
	} else if t == 1 {
		if notifAlert {
			notifAlert = false
			fmt.Printf("Nofification: Disabled\n")
		} else {
			notifAlert = true
			fmt.Printf("Nofification: Enabled\n")
		}
	}
}

// Implement funtion to change color's text on console
func red(a string) string {
	f := color.FgRed.Render
	return f(a)
}

func green(a string) string {
	f := color.FgGreen.Render
	return f(a)
}

var number string // Phone number
var lastMsg string

type waHandler struct {
	c *whatsapp.Conn
}

func printHistory() {
	sort.Sort(byTime(historyMessages))
	fmt.Printf("History\n----------------------\n")
	for i := range historyMessages {
		message := historyMessages[i]
		t := time.Unix(int64(message.Info.Timestamp), 0)
		if message.Info.FromMe == true {
			fmt.Printf("%v %v\n", red("["+t.Format("01/02/2006 15:04:05")+"] >>"), message.Text)
		} else {
			fmt.Printf("%v %v\n", green("["+t.Format("01/02/2006 15:04:05")+"] <<"), message.Text)
		}
	}
	fmt.Printf("----------------------\n")
}

/*
func showNotification(s string) {
	notification := toast.Notification{
		AppID:   "WZP",
		Title:   "MSG",
		Message: s,
		//Icon:    "go.png", // This file must exist (remove this line if it doesn't)
		Actions: []toast.Action{
			{"protocol", "I'm a button", ""},
			{"protocol", "Me too!", ""},
		},
	}
	err := notification.Push()
	if err != nil {
		log.Fatalln(err)
	}
}
*/

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (*waHandler) HandleTextMessage(message whatsapp.TextMessage) {

	if message.Info.RemoteJid == number+"@s.whatsapp.net" {
		t := time.Unix(int64(message.Info.Timestamp), 0)
		lastMsg = message.Text
		diff := ctime.Sub(t)
		if diff < 0 { //new messages
			fmt.Printf("[new]")
			var err error
			if soundAlert {
				err = beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
				if err != nil {
					panic(err)
				}
			}
			if notifAlert {
				err = beeep.Alert("WZP", message.Text, "")
				if err != nil {
					panic(err)
				}
			}
			if message.Info.FromMe == true {
				fmt.Printf("%v %v\n", red("["+t.Format("01/02/2006 15:04:05")+"] >>"), message.Text)
			} else {
				fmt.Printf("%v %v\n", green("["+t.Format("01/02/2006 15:04:05")+"] <<"), message.Text)
			}
		} else {
			historyMessages = append(historyMessages, message) //Add to history message array
		}
	}
}

func SendMessage(w *whatsapp.Conn, m string) {

	previousMessage := "xD"
	quotedMessage := proto.Message{
		Conversation: &previousMessage,
	}

	ContextInfo := whatsapp.ContextInfo{
		QuotedMessage:   &quotedMessage,
		QuotedMessageID: "",
		Participant:     "", //Whot sent the original message
	}

	msg := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: number + "@s.whatsapp.net",
		},
		ContextInfo: ContextInfo,
		Text:        m,
	}

	//msgId, err := w.Send(msg)
	_, err := w.Send(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		os.Exit(1)
	} else {
		//fmt.Printf(red("\033[F\r[Sent] >> ")) //delete previous line
		fmt.Printf(red("[Sent] >> "))
		fmt.Println(m)
	}
}

func main() {
	fmt.Printf("WZP v0.1\n")

	if len(os.Args) != 2 {
		log.Fatalf("You MUST pass a number and message argument\n USAGE ./wsp <number>\n")
	}
	number = os.Args[1]

	fmt.Printf("[!] Favorite number: %s\n", number)

	ctime = time.Now()

	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(10 * time.Second)
	if err != nil {
		log.Fatalf("error creating connection: %v\n", err)
	}

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		log.Fatalf("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		log.Fatalf("error pinging in: %v\n", err)
	}

	// wait while chat jids are acquired through incoming initial messages
	fmt.Println("[!] Waiting for chats info...")
	<-time.After(6 * time.Second)
	printHistory()

	//generate buffer to send messages
	reader := bufio.NewReader(os.Stdin)

	go func() {
		for {
			text, _ := reader.ReadString('\n')
			if text == "*h\n" {
				printHistory()
			} else if text == "*s\n" {
				toggleAlert(0)
			} else if text == "*n\n" {
				toggleAlert(1)
			} else {
				SendMessage(wac, text)
			}

		}
	}()

	//wait signal to shut down application
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		log.Fatalf("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		log.Fatalf("error saving session: %v", err)
	}
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	fmt.Printf("[!] Session saved in: %v as a %v\n", os.TempDir(), "whatsappSession.gob")
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
