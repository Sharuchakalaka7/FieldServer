package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

/** Global Variables **/

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	metricCount int
	metricName []string
	metricUnit []string
	data []string
)




/** Common Helper Functions **/

func readWSMsg(conn *websocket.Conn) string {
	// Retrieve incomming data
	messageType, byteArray, err := conn.ReadMessage()

	if err != nil {
		// Quit from error state
		log.Println("ERROR --> ", err)
		return ""

	} else {
		// Parse and display recieved data on log and back to client
		text := string(byteArray)
		log.Println("RECEIVED --> ", text)
		if err := conn.WriteMessage(messageType, byteArray); err != nil { log.Println(err) }
		return text
	}
}

func writeWSMsg(conn *websocket.Conn, msg string) {
	// For now, only use TXT WebSocket messages to write - other types introduced when needed
	const TXT_TYPE int = 1
	err := conn.WriteMessage(TXT_TYPE, []byte(msg))
	if (err != nil) { log.Println(err) }
}

func JSONify(elems []string, quotify bool) string {
	if (len(elems) == 0) { return "" }
	var json string = "["
	if (quotify) { json += "\"" }
	json += elems[0]
	if (quotify) { json += "\"" }

	for i := 1; i < len(elems); i++ {
		json += ","
		if (quotify) { json += "\"" }
		json += elems[i]
		if (quotify) { json += "\"" }
	}
	json += "]\n"
	return json
}

func upgradeConnection(w http.ResponseWriter, r *http.Request, msg string) *websocket.Conn {
	// Upgrade HTTP to WebSocket Protocol
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	// Log any errors, then return connection ptr
	if (err != nil) { log.Println(err) }
	log.Println(msg)
	return conn
}





/** ESP Route **/

func espReader(conn *websocket.Conn) {
	// Read initial configurations

	// Initialize number of metrics, then make slices for name and unit
	metricCount, _ = strconv.Atoi(readWSMsg(conn))
	metricName = make([]string, metricCount)
	metricUnit = make([]string, metricCount)
	data = make([]string, metricCount)

	// Read the name and unit for each metric
	for i := 0; i < metricCount; i++ {
		metricName[i] = readWSMsg(conn)
		metricUnit[i] = readWSMsg(conn)
	}

	// Read data (loop)
	for {
		for i := 0; i < metricCount; i++ {
			data[i] = readWSMsg(conn)
		}
		readWSMsg(conn)		// remove conventional empty line from buffer
	}
}

func espEndpoint(w http.ResponseWriter, r *http.Request) {
	conn := upgradeConnection(w, r, "ESP Successfully connected...")
	espReader(conn)
}




/** Client Route **/

func clientWriter(conn *websocket.Conn) {
	// Keep sending data
	for {
		outerSlice := []string{ JSONify(metricName, true), JSONify(metricUnit, true), JSONify(data, true) }
		writeWSMsg(conn, JSONify(outerSlice, false))
	}
}

func clientEndpoint(w http.ResponseWriter, r *http.Request) {
	conn := upgradeConnection(w, r, "Client connected successfully...")
	clientWriter(conn)
}




/** Core Functions **/

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Data Viewer\n\n")
	for i := 0; i < metricCount; i++ {
		fmt.Fprintf(w, "%s (%s): %15s\n", metricName[i], metricUnit[i], data[i])
	}
}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/esp", espEndpoint)
	http.HandleFunc("/client", clientEndpoint)
}

func main() {
	fmt.Println("Go WebSockets")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
